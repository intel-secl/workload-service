package main

import (
	"fmt"
	"intel/isecl/workload-service/config"
	"intel/isecl/workload-service/setup"
	"intel/isecl/workload-service/constants"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"os/exec"
	"intel/isecl/lib/common/validation"
	csetup "intel/isecl/lib/common/setup"
	// Import Postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
	stdlog "log"
	log "github.com/sirupsen/logrus"
	e "intel/isecl/lib/common/exec"
)

func main() {
	/* BEGIN LOG CONFIGURATION */
	wlsLogFile, err := os.OpenFile("/var/log/workload-service/wls.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	defer wlsLogFile.Close()
	if err != nil {
		log.WithError(err).Info("Failed to open log file, using stderr")
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(io.MultiWriter(os.Stderr, wlsLogFile))
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

	inputStringArr := os.Args[0:]
	if err := validation.ValidateStrings(inputStringArr); err != nil {
		fmt.Println("Invalid input")
		printUsage()
		os.Exit(0)
	}
	
	switch arg := strings.ToLower(args[0]); arg {
	case "setup":
		if nosetup, err := strconv.ParseBool(os.Getenv("WLS_NOSETUP")); err != nil && nosetup == false {
			setupRunner := &csetup.Runner{
				Tasks: []csetup.Task{
					csetup.Download_Ca_Cert{
						Flags:         args,
						CaCertDirPath: constants.TrustedCaCertsDir,
						ConsoleWriter: os.Stdout,
					},
					csetup.Download_Cert{
						Flags:              args,
						KeyFile:            constants.TLSKeyPath,
						CertFile:           constants.TLSCertPath,
						KeyAlgorithm:       constants.DefaultKeyAlgorithm,
						KeyAlgorithmLength: constants.DefaultKeyAlgorithmLength,
						CommonName:         constants.DefaultWlsTlsCn,
						SanList:            constants.DefaultWlsTlsSan,
						CertType:           "TLS",
						CaCertsDir:         constants.TrustedCaCertsDir,
						BearerToken:        "",
						ConsoleWriter:      os.Stdout,
					},
					new(setup.Server),
					new(setup.Database),
					new(setup.HVSConnection),
					new(setup.KMSConnection),
					new(setup.AASConnection),
					new(setup.Logs),
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
		start()

	case "status":
		status()

	case "startserver":
		// this runs in attached mode
		startServer()

	case "stop":
		stop()

	case "uninstall":
		stop()
		removeService()
		deleteFile("/opt/workload-service/")
		deleteFile("/usr/local/bin/workload-service")
		deleteFile("/var/log/workload-service/")
		if len(args) > 1 && strings.ToLower(args[1]) == "--purge" {
			deleteFile("/etc/workload-service/")
		}		

	default:
		fmt.Printf("Unrecognized option : %s\n", arg)
		fallthrough

	case "help", "-help", "--help":
		printUsage()
	}
}

func start() error {
	fmt.Println("Starting Workload Service")
	systemctl, err := exec.LookPath("systemctl")
	if err != nil {
		fmt.Println("Error trying to look up for systemctl path")
		os.Exit(1)
	}
	return syscall.Exec(systemctl, []string{"systemctl", "start", "workload-service"}, os.Environ())
}

func stop() error {
	fmt.Println("Stopping Workload Service")
	_, _, err := e.RunCommandWithTimeout("systemctl stop workload-service", 5)
	if err != nil {
		fmt.Println("Could not stop Workload-service")
		fmt.Println("Error : ", err)
	}
	return err
}

func status() error {
	fmt.Println("Forwarding to systemctl status workload-service")
	systemctl, err := exec.LookPath("systemctl")
	if err != nil {
		fmt.Println("Error trying to look up for systemctl path")
		os.Exit(1)
	}
	return syscall.Exec(systemctl, []string{"systemctl", "status", "workload-service"}, os.Environ())
}

func removeService() error {
	fmt.Println("Removing Workload Service")
	_, _, err := e.RunCommandWithTimeout("systemctl disable workload-service", 5)
	if err != nil {
		fmt.Println("Could not remove Workload-service")
		fmt.Println("Error : ", err)
	}
	return err
}

func deleteFile(path string) {
	fmt.Println("Deleting : ", path)
	// delete file
	var err = os.RemoveAll(path)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func printUsage() {
	fmt.Printf("Usage:\n\n")
	fmt.Printf("    %s <command>\n\n", os.Args[0])
	fmt.Printf("Available Commands:\n")
	fmt.Printf("    help|-help|--help   Show this help message\n")
	fmt.Printf("    start               Start workload-service\n")
	fmt.Printf("    stop                Stop workload-service\n")
	fmt.Printf("    status              Determine if workload-service is running\n")
	fmt.Printf("    setup               Setup workload-service for use\n\n")
	fmt.Printf("Available tasks for setup:\n")
	fmt.Printf("    server\n")
	fmt.Printf("    database\n")
	fmt.Printf("    hvsconnection\n")
	fmt.Printf("    kmsconnection\n")
	fmt.Printf("    logs\n")
 }
