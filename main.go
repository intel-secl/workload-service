package main

import (
	"intel/isecl/lib/common/logger"
	"time"
	"context"
	"os/signal"
	"os/exec"
	"io/ioutil"
	"syscall"
	"intel/isecl/workload-service/setup"
	"fmt"
	"intel/isecl/workload-service/repository/postgres"
	"intel/isecl/workload-service/resource"
	"log"
	"net/http"
	"os"
	"strings"
	"strconv"

	"github.com/jinzhu/gorm"

	"intel/isecl/workload-service/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/handlers"
	// Import Postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
	csetup "intel/isecl/lib/common/setup"
)

func main() {
	args := os.Args[1:]
	if len(args) <= 0 {
		fmt.Println("Command not found. Usage below ", os.Args[0])
		printUsage()
		return
	}

	switch arg := strings.ToLower(args[0]); arg {
	case "setup":
		if nosetup, err := strconv.ParseBool(os.Getenv("WLS_NOSETUP")); err != nil && nosetup == false {
			setupRunner := &csetup.Runner {
				Tasks: []csetup.Task{
					new(setup.Server),
					new(setup.Database),
					new(setup.HVSConnection),
					new(setup.KMSConnection),
				},
				AskInput: false,
			}
			err := setupRunner.RunTasks(args[1:]...)
			if err != nil {
				fmt.Println("Error running setup: ", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("WLS_NOSETUP is set, skipping setup")
			os.Exit(1)
		}
	case "start":
		err := start() // this starts server detached
		if err != nil {
			fmt.Println("Failed to start server")
			os.Exit(1)
		}
	case "status":
		if s := status(); s == Running {
			fmt.Println("Workload Service is running")
		} else {
			fmt.Println("Workload Service is stopped")
		}
	case "startserver":
		// this runs in attached mode
		startServer()
	case "stop":
		stopServer()
	case "uninstall":
		uninstall()
	default:
		fmt.Printf("Unrecognized option : %s\n", arg)
		fallthrough

	case "help", "-help", "--help":
		printUsage()
	}
}

const pidPath = "/var/run/workload-service/wls.pid"

// Status indicate the process status of WLS
type Status bool

const (
	Stopped Status = false
	Running Status = true
)

func status() Status {
	pid, err := readPid() 
	if err != nil {
		os.Remove(pidPath)
		return Stopped
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return Stopped
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		return Stopped
	}
	return Running
}

func uninstall() {
	fmt.Println("Not yet supported")
}

func printUsage() {
	fmt.Printf("Work Load Service\n")
	fmt.Printf("===============\n\n")
	fmt.Printf("usage : %s <command> [<args>]\n\n", os.Args[0])
	fmt.Printf("Following are the list of commands\n")
	fmt.Printf("\tsetup\n\n")
	fmt.Printf("setup command is used to run setup tasks\n")
	fmt.Printf("\tusage : %s setup [<tasklist>]\n", os.Args[0])
	fmt.Printf("\t\t<tasklist>-space seperated list of tasks\n")
	fmt.Printf("\t\t\t-Supported tasks - server database\n")
	fmt.Printf("\tExample :-\n")
	fmt.Printf("\t\t%s setup\n", os.Args[0])
	fmt.Printf("\t\t%s setup database\n", os.Args[0])
}

func stopServer() {
	pid, err := readPid()
	if err != nil {
		log.Printf("Error: %v\n", err)
	}
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		log.Printf("Error: %v\n", err)
	}
	log.Println("Workload Service Stopped")
}

func readPid() (int, error) {
	pidData, err := ioutil.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(pidData))
	if err != nil {
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
			log.Printf("Failed to connect to DB, retrying in %d seconds ...\n", retryTime)
		}
		time.Sleep(retryTime*time.Second)
	}
	defer db.Close()
	if dbErr != nil {
		log.Fatal("could not open db: ", dbErr)
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
	h := &http.Server {
		Addr: fmt.Sprintf(":%d", config.Configuration.Port),
		Handler: handlers.RecoveryHandler(handlers.RecoveryLogger(logger.Error), handlers.PrintRecoveryStack(true))(r),
	}
	// dispatch http listener on separate go routine
	log.Println("Starting Workload Service ...")
	go func() {
		log.Println("Workload Service Started")
		if err := h.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
	// wait for a signal on the stop channel
	<-stop // swallow the value, as we don't really care what it is

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.Shutdown(ctx); err != nil {
		log.Printf("Failed to gracefully shutdown webserver: %v\n", err)
	} else {
		log.Println("Workload Service stopped")
	}
}
