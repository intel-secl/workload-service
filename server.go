/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"intel/isecl/lib/common/middleware"
	"intel/isecl/workload-service/config"
	"intel/isecl/workload-service/constants"
	consts "intel/isecl/workload-service/constants"
	"intel/isecl/workload-service/repository/postgres"
	"intel/isecl/workload-service/resource"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/pkg/errors"
)

var cacheTime, _ = time.ParseDuration(constants.JWTCertsCacheTime)

//To be implemented if JWT certificate is needed from any other services
func fnGetJwtCerts() error {
	log.Trace("server:fnGetJwtCerts() Entering")
	defer log.Trace("server:fnGetJwtCerts() Leaving")
	return nil
}

func startServer() error {
	log.Trace("server:startServer() Entering")
	defer log.Trace("server:startServer() Leaving")
	var sslMode string
	if config.Configuration.Postgres.SSLMode {
		sslMode = "enable"
		log.Info("server:startServer() sslMode is enabled")
	} else {
		sslMode = "disable"
		log.Info("server:startServer() sslMode is disabled")
	}
	var db *gorm.DB
	var dbErr error
	var numAttempts = 4
	for i := 0; i < numAttempts; i = i + 1 {
		const retryTime = 5
		db, dbErr = gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			config.Configuration.Postgres.Hostname, config.Configuration.Postgres.Port, config.Configuration.Postgres.User, config.Configuration.Postgres.DBName, config.Configuration.Postgres.Password, sslMode))
		if dbErr != nil {
			secLog.Warnf("server:startServer() DB connection attempt %d of %d failed: %s", (i + 1), numAttempts, dbErr)
			fmt.Printf("Failed to connect to DB, retrying in %d seconds ...\n", retryTime)
		} else {
			break
		}
		time.Sleep(retryTime * time.Second)
	}
	defer db.Close()
	if dbErr != nil {
		secLog.Fatalf("server:startServer() Failed to establish DB connection: %s\n", dbErr.Error())
		return dbErr
	}
	wlsDb := postgres.PostgresDatabase{DB: db}
	wlsDb.Migrate()
	r := mux.NewRouter().PathPrefix("/wls").Subrouter()
	r.Use(middleware.NewTokenAuth(consts.TrustedJWTSigningCertsDir, consts.TrustedCaCertsDir, fnGetJwtCerts, cacheTime))
	// Set Resource Endpoints
	resource.SetFlavorsEndpoints(r.PathPrefix("/flavors").Subrouter(), wlsDb)
	// Setup Report Endpoints
	resource.SetReportsEndpoints(r.PathPrefix("/reports").Subrouter(), wlsDb)
	// Setup Images Endpoints
	resource.SetImagesEndpoints(r.PathPrefix("/images").Subrouter(), wlsDb)
	// Setup Version Endpoint
	resource.SetVersionEndpoints(r.PathPrefix("/version").Subrouter(), wlsDb)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	httpWriter := os.Stderr
	if httpLogFile, err := os.OpenFile(consts.HttpLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
		secLog.WithError(err).Errorf("server:startServer() Failed to open http log file: %s\n", err.Error())
		log.Tracef("%+v", err)
	} else {
		defer httpLogFile.Close()
		httpWriter = httpLogFile
	}
	tlsconfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
	}
	l := stdlog.New(httpWriter, "", 0)
	h := &http.Server{
		Addr:      fmt.Sprintf(":%d", config.Configuration.Port),
		Handler:   handlers.RecoveryHandler(handlers.RecoveryLogger(l), handlers.PrintRecoveryStack(true))(handlers.CombinedLoggingHandler(httpWriter, r)),
		ErrorLog:  l,
		TLSConfig: tlsconfig,
		ReadTimeout:       config.Configuration.ReadTimeout,
		ReadHeaderTimeout: config.Configuration.ReadHeaderTimeout,
		WriteTimeout:      config.Configuration.WriteTimeout,
		IdleTimeout:       config.Configuration.IdleTimeout,
		MaxHeaderBytes:    config.Configuration.MaxHeaderBytes,
	}

	// dispatch web server go routine
	fmt.Println("Starting Workload Service ...")
	go func() {
		tlsCert := consts.TLSCertPath
		tlsKey := consts.TLSKeyPath
		fmt.Println("Workload Service Started")
		if err := h.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
			secLog.WithError(err).Fatalf("server:startServer() Failed to start HTTPS server: %s\n", err.Error())
			log.Tracef("%+v", err)
			stop <- syscall.SIGTERM
		}
	}()

	fmt.Println("Workload Service is running")
	secLog.Infof("server:startServer() Workload Service is running. Listening at port %d", config.Configuration.Port)
	// TODO dispatch Service status checker goroutine
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.Shutdown(ctx); err != nil {
		fmt.Printf("Failed to gracefully shutdown webserver: %v\n", err)
		log.Tracef("%+v", err)
		return errors.Wrapf(err, "server:startServer() Failed to gracefully shutdown webserver: %v\n", err)
	}
	return nil
}
