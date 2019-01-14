package main

import (
	"fmt"
	"intel/isecl/workload-service/config"
	"intel/isecl/workload-service/setup"
	"os"
	"strconv"
	"strings"
	"syscall"

	csetup "intel/isecl/lib/common/setup"
	// Import Postgres driver

	_ "github.com/jinzhu/gorm/dialects/postgres"

	stdlog "log"

	log "github.com/sirupsen/logrus"
)

func main() {
	/* BEGIN LOG CONFIGURATION */
	wlsLogFile, err := os.OpenFile("/var/log/workload-service/wls.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer wlsLogFile.Close()
	if err != nil {
		log.WithError(err).Info("Failed to open log file, using stderr")
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(wlsLogFile)
	}
	log.SetLevel(config.Configuration.LogLevel)
	// override standard golang log
	w := log.StandardLogger().WriterLevel(config.Configuration.LogLevel)
	defer w.Close()
	stdlog.SetOutput(w)
	/* END LOG CONFIGURATION */

	/* PARSE COMMAND LINE OPTIONS */
	args := os.Args[1:]
	if len(args) <= 0 {
		fmt.Println("Command not found. Usage below ", os.Args[0])
		printUsage()
		return
	}
	switch arg := strings.ToLower(args[0]); arg {
	case "setup":
		if nosetup, err := strconv.ParseBool(os.Getenv("WLS_NOSETUP")); err != nil && nosetup == false {
			setupRunner := &csetup.Runner{
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
