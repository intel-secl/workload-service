package main

import (
	"crypto/x509/pkix"
	"fmt"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/lib/common/validation"
	"intel/isecl/workload-service/config"
	"intel/isecl/workload-service/constants"
	"intel/isecl/workload-service/setup"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	// Import Postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	e "intel/isecl/lib/common/exec"
	stdlog "log"
)

func main() {
	var context csetup.Context
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
		os.Exit(1)
	}

	err = config.SaveConfiguration(context)
	if err != nil {
		fmt.Println("Error: Unable to save configuration in config.yml")
		os.Exit(1)
	}
	switch arg := strings.ToLower(args[0]); arg {
	case "setup":
		flags := args
		if len(args) > 1 {
			flags = args[2:]
			if args[1] == "download_cert" && len(args) > 2 {
				flags = args[3:]
			}
		}
		setupRunner := &csetup.Runner{
			Tasks: []csetup.Task{
				csetup.Download_Ca_Cert{
					Flags:         flags,
					CmsBaseURL:    config.Configuration.CMS_BASE_URL,
					CaCertDirPath: constants.TrustedCaCertsDir,
					ConsoleWriter: os.Stdout,
				},
				csetup.Download_Cert{
					Flags:              flags,
					KeyFile:            constants.TLSKeyPath,
					CertFile:           constants.TLSCertPath,
					KeyAlgorithm:       constants.DefaultKeyAlgorithm,
					KeyAlgorithmLength: constants.DefaultKeyAlgorithmLength,
					CmsBaseURL:         config.Configuration.CMS_BASE_URL,
					Subject:            pkix.Name{
						Country:            []string{config.Configuration.Subject.Country},
						Organization:       []string{config.Configuration.Subject.Organization},
						Locality:           []string{config.Configuration.Subject.Locality},
						Province:           []string{config.Configuration.Subject.Province},
						CommonName:         config.Configuration.Subject.TLSCertCommonName,
					},
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
	fmt.Println("Usage:")
	fmt.Printf("    %s <command>\n\n", os.Args[0])
	fmt.Println("Available Commands:")
	fmt.Println("    help|-help|--help   Show this help message")
	fmt.Println("    start               Start workload-service")
	fmt.Println("    stop                Stop workload-service")
	fmt.Println("    status              Determine if workload-service is running")
	fmt.Printf("    setup               Setup workload-service for use\n\n")
	fmt.Printf("Setup command usage:  %s <command> [task...]\n", os.Args[0])
	fmt.Println("Available tasks for setup:")
	fmt.Println("    server              Setup http server on given port")
	fmt.Printf("                        Environment variable WLS_PORT=<port> should be set\n\n")
	fmt.Println("    database            Setup workload-service database")
	fmt.Println("                        Required env variables are:")
	fmt.Println("                        - WLS_DB_HOSTNAME  : database host name")
	fmt.Println("                        - WLS_DB_PORT      : database port number")
	fmt.Println("                        - WLS_DB_USERNAME  : database user name")
	fmt.Println("                        - WLS_DB_PASSWORD  : database password")
	fmt.Printf("                        - WLS_DB           : database schema name\n\n")
	fmt.Println("    hvsconnection       Setup task for setting up the connection to the Host Verification Service(HVS)")
	fmt.Println("                        Required env variables are:")
	fmt.Println("                        - HVS_URL      : HVS URL")
	fmt.Println("                        - HVS_USER     : HVS API user name")
	fmt.Printf("                        - HVS_PASSWORD : HVS API password\n\n")
	fmt.Println("    kmsconnection       Setup task for setting up the connection to the key management service(KMS)")
	fmt.Println("                        - KMS_URL      : KMS URL")
	fmt.Println("                        - KMS_USER     : KMS API user name")
	fmt.Printf("                        - KMS_PASSWORD : KMS API password\n\n")
	fmt.Println("    logs                Setup workload-service log level")
	fmt.Printf("                        Environment variable WLS_LOG_LEVEL=<log level> should be set\n\n")
	fmt.Println("    aas                 Setup to create workload service user roles in AAS")
	fmt.Printf("                        Environment variable AAS_API_URL=<aas URL> should be set\n\n")
 }
