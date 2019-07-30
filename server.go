package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"intel/isecl/lib/common/middleware"
	"intel/isecl/workload-service/config"
	consts "intel/isecl/workload-service/constants"
	"intel/isecl/workload-service/repository/postgres"
	"intel/isecl/workload-service/resource"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/jinzhu/gorm"
	
	stdlog "log"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)


//To be implemented if JWT certificate is needed from any other services
func fnGetJwtCerts() error{
	
	return nil
}

func startServer() error {
	var sslMode string
	if config.Configuration.Postgres.SSLMode {
		sslMode = "enable"
	} else {
		sslMode = "disable"
	}
	var db *gorm.DB
	var dbErr error
	var numAttempts = 4
	for i := 0; i < numAttempts; i = i + 1 {
		const retryTime = 5
		db, dbErr = gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			config.Configuration.Postgres.Hostname, config.Configuration.Postgres.Port, config.Configuration.Postgres.User, config.Configuration.Postgres.DBName, config.Configuration.Postgres.Password, sslMode))
		if dbErr != nil {
			log.Warn(fmt.Sprintf("DB connection attempt %d of %d failed: %s", (i + 1), numAttempts, dbErr))
			fmt.Printf("Failed to connect to DB, retrying in %d seconds ...\n", retryTime)
		} else {
			break
		}
		time.Sleep(retryTime * time.Second)
	}
	defer db.Close()
	if dbErr != nil {
		fmt.Printf("Failed to connect to db after many attempts: %s\n", dbErr.Error())
		return dbErr
	}
	wlsDb := postgres.PostgresDatabase{DB: db}
	wlsDb.Migrate()
	r := mux.NewRouter().PathPrefix("/wls").Subrouter()
	r.Use(middleware.NewTokenAuth(consts.TrustedJWTSigningCertsDir, consts.TrustedCaCertsDir, fnGetJwtCerts))
	// Set Resource Endpoints
	resource.SetFlavorsEndpoints(r.PathPrefix("/flavors").Subrouter(), wlsDb)
	// Setup Report Endpoints
	resource.SetReportsEndpoints(r.PathPrefix("/reports").Subrouter(), wlsDb)
	// Setup Images Endpoints
	resource.SetImagesEndpoints(r.PathPrefix("/images").Subrouter(), wlsDb)
	// Setup Version Endpoint
	resource.SetVersionEndpoints(r, wlsDb)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	httpWriter := os.Stderr
	if httpLogFile, err := os.OpenFile("/var/log/workload-service/http.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err != nil {
		log.WithError(err).Info("Failed to open http log file")
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
		Addr:     fmt.Sprintf(":%d", config.Configuration.Port),
		Handler:  handlers.RecoveryHandler(handlers.RecoveryLogger(l), handlers.PrintRecoveryStack(true))(handlers.CombinedLoggingHandler(httpWriter, r)),
		ErrorLog: l,
		TLSConfig: tlsconfig,
	}

	// dispatch web server go routine
	fmt.Println("Starting Workload Service ...")
	go func() {
		tlsCert := consts.TLSCertPath
		tlsKey := consts.TLSKeyPath
		fmt.Println("Workload Service Started")
		if err := h.ListenAndServeTLS(tlsCert, tlsKey); err != nil {
			log.WithError(err).Info("Failed to start HTTPS server")
			stop <- syscall.SIGTERM
		}
	}()

	fmt.Println("Workload Service is running")
	// TODO dispatch Service status checker goroutine
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.Shutdown(ctx); err != nil {
		fmt.Printf("Failed to gracefully shutdown webserver: %v\n", err)
		return err
	}
	return nil
}
