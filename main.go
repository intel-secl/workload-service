package main

import (
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
	// Import Postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
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
			err := setup.RunSetupTasks(args[1:]...)
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
	fmt.Println("Stopping Workload Service")
	pidData, err := ioutil.ReadFile("/var/run/workload-service/wls.pid")
	if err != nil {
		pid, _ := strconv.Atoi(string(pidData))
		syscall.Kill(pid, 9)
	}
}

func start() error {
	// spawn another process
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
    cmd.Process.Release()
    return nil
}

func startServer() {
	// source configuration from somewhere
	var sslMode string
	if config.Configuration.Postgres.SSLMode {
		sslMode = "enable"
	} else {
		sslMode = "disable"
	}
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s",
		config.Configuration.Postgres.Hostname, config.Configuration.Postgres.Port, config.Configuration.Postgres.User, config.Configuration.Postgres.DBName, config.Configuration.Postgres.Password, sslMode))
	defer db.Close()
	if err != nil {
		log.Fatal("could not open db: ", err)
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

	// store pid
	file, _ := os.Create("/var/run/workload-service/wls.pid")
	file.WriteString(strconv.Itoa(os.Getpid()))
	fmt.Println("Workload Service Started")
	http.ListenAndServe(fmt.Sprintf(":%d", config.Configuration.Port), r)
}
