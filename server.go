package main

import (
	"context"
	"fmt"
	"intel/isecl/workload-service/config"
	"intel/isecl/workload-service/repository/postgres"
	"intel/isecl/workload-service/resource"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
	stdlog "log"
)

func stopServer() {
	pid, err := readPid()
	if err != nil {
		log.WithError(err).Error("Failed to stop server")
	}
	if err := syscall.Kill(pid, syscall.SIGQUIT); err != nil {
		log.WithError(err).Error("Failed to kill server")
	}
	fmt.Println("Workload Service Stopped")
}

func readPid() (int, error) {
	pidData, err := ioutil.ReadFile(pidPath)
	if err != nil {
		log.WithError(err).Debug("Failed to read pidfile")
		return 0, err
	}
	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
		log.WithError(err).WithField("pid", pidData).Debug("Failed to convert pidData string to int")
		return 0, err
	}
	return pid, nil
}

func start() error {
	// first check to see if the pid specified in /var/run is already running
	// spawn another process
	fmt.Println("Starting Workload Service ...")
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd := exec.Command(os.Args[0], "startServer")
	cmd.Dir = cwd
	err = cmd.Start()
	if err != nil {
		return err
	}
	// store pid
	file, _ := os.Create(pidPath)
	file.WriteString(strconv.Itoa(cmd.Process.Pid))
	cmd.Process.Release()
	fmt.Println("Workload Service started")
	return nil
}

func startServer() {
	var sslMode string
	if config.Configuration.Postgres.SSLMode {
		sslMode = "enable"
	} else {
		sslMode = "disable"
	}
	var db *gorm.DB
	var dbErr error
	for i := 0; i < 4; i = i + 1 {
		const retryTime = 5
		db, dbErr = gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
			config.Configuration.Postgres.Hostname, config.Configuration.Postgres.Port, config.Configuration.Postgres.User, config.Configuration.Postgres.DBName, config.Configuration.Postgres.Password, sslMode))
		if dbErr != nil {
			log.WithError(dbErr).Info("Failed to connect to DB")
			fmt.Printf("Failed to connect to DB, retrying in %d seconds ...\n", retryTime)
		} else {
			break
		}
		time.Sleep(retryTime * time.Second)
	}
	defer db.Close()
	if dbErr != nil {
		log.Fatal("Failed to connect to db after many attempts: ", dbErr)
	}
	wlsDb := postgres.PostgresDatabase{DB: db}
	wlsDb.Migrate()
	r := mux.NewRouter().PathPrefix("/wls").Subrouter()
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
	l := stdlog.New(httpWriter, "", 0)
	h := &http.Server{
		Addr:     fmt.Sprintf(":%d", config.Configuration.Port),
		Handler:  handlers.RecoveryHandler(handlers.RecoveryLogger(l), handlers.PrintRecoveryStack(true))(handlers.CombinedLoggingHandler(httpWriter, r)),
		ErrorLog: l,
	}
	// dispatch http listener on separate go routine
	fmt.Println("Starting Workload Service ...")
	go func() {
		fmt.Println("Workload Service Started")
		if err := h.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	// wait for a signal on the stop channel
	<-stop // swallow the value, as we don't really care what it is

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.Shutdown(ctx); err != nil {
		fmt.Printf("Failed to gracefully shutdown webserver: %v\n", err)
	} else {
		fmt.Println("Workload Service stopped")
	}
}
