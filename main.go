package main

import (
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
		fmt.Println("Command not found. Usage below ", args[0])
		printUsage()
		return
	}

	switch arg := strings.ToLower(args[0]); arg {
	case "setup":
		if nosetup, err := strconv.ParseBool(os.Getenv("WLS_NOSETUP")); err != nil && nosetup == false {
			setup.RunSetupTasks(args[2:]...)
		} else {
			fmt.Println("WLS_NOSETUP is set, skipping setup")
		}
	case "start":
		startServer()
	case "stop":
		stopServer()
	default:
		fmt.Printf("Unrecognized option : %s\n", arg)
		fallthrough

	case "help", "-help", "--help":
		printUsage()
	}
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
	fmt.Printf("\t\t\t-Supported tasks - SampleSetupTask\n")
	fmt.Printf("\tExample :-\n")
	fmt.Printf("\t\t%s setup\n", os.Args[0])
	fmt.Printf("\t\t%s setup SampleSetupTask\n", os.Args[0])
}

func stopServer() {

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
		log.Fatal("could not open db", err)
	}
	wlsDb := postgres.PostgresDatabase{db}
	wlsDb.Migrate()
	r := mux.NewRouter().PathPrefix("/wls").Subrouter()
	// Set Resource Endpoints
	resource.SetFlavorsEndpoints(r.PathPrefix("/flavors").Subrouter(), wlsDb)
	// Set Image Endpoints
	resource.SetImagesEndpoints(r.PathPrefix("/images").Subrouter(), db)
	// Setup Report Endpoints
	resource.SetReportsEndpoints(r.PathPrefix("/reports").Subrouter(), db)
	// Setup Version Endpoint
	resource.SetVersionEndpoints(r, wlsDb)
	http.ListenAndServe(fmt.Sprintf(":%d", config.Configuration.Port), r)
}
