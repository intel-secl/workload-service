/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package main

import (
	"crypto/x509/pkix"
	"fmt"
	commLog "intel/isecl/lib/common/log"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/lib/common/validation"
	"intel/isecl/workload-service/config"
	"intel/isecl/workload-service/constants"
	"intel/isecl/workload-service/setup"
	"intel/isecl/workload-service/version"
	"os"
	"os/exec"
	"strings"
	"syscall"

	// Import Postgres driver
	e "intel/isecl/lib/common/exec"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var log = commLog.GetDefaultLogger()
var secLog = commLog.GetSecurityLogger()

func main() {
	log.Trace("main:main() Entering")
	defer log.Trace("main:main() Leaving")

	var context csetup.Context

	/* PARSE COMMAND LINE OPTIONS */
	args := os.Args[1:]

	if len(args) <= 0 {
		fmt.Println("Command not found. Usage below ", os.Args[0])
		printUsage()
		return
	}

	if err := validation.ValidateStrings(args); err != nil {
		secLog.WithError(err).Error("main:main() Invalid Input")
		log.Tracef("%+v", err)
		fmt.Fprintln(os.Stderr, "Invalid input")
		printUsage()
		os.Exit(1)
	}

	// force all args to lowercase
	for i, x := range args {
		args[i] = strings.ToLower(x)
	}

	switch arg := args[0]; arg {
	case "setup":
		args := os.Args[1:]
		if len(args) <= 1 {
			printUsage()
			return
		}

		flags := args
		if len(args) >= 2 &&
			args[1] != "download_ca_cert" &&
			args[1] != "download_cert" &&
			args[1] != "server" &&
			args[1] != "download_saml_ca_cert" &&
			args[1] != "database" &&
			args[1] != "hvsconnection" &&
			args[1] != "all" {
			fmt.Fprintln(os.Stderr, "Error: Unknown setup task ", args[1])
			printUsage()
			os.Exit(1)
		}

		// check if the WLS_NOSETUP env flag is set to "true", if so then skip setup
		nosetup, err := context.GetenvString(config.WLS_NOSETUP, "WLS Setup Skip Flag")
		if err == nil && strings.ToLower(nosetup) == "true" {
			fmt.Println(config.WLS_NOSETUP, " is true, skipping setup")
			os.Exit(0)
		}

		if len(args) > 2 {
			// check if TLS cert type was specified for downloiad
			if args[1] == "download_cert" {
				if args[2] != "tls" {
					fmt.Println("Invalid cert type provided for download_cert setup task: Only TLS cert type is supported. Aborting.")
					os.Exit(1)
				} else if len(args) > 3 {
					// flags will be post the tls arg
					flags = args[3:]
				}
			} else {
			// flags for arguments
			flags = args[2:]
		}
		}

		// log initialization
		config.LogConfiguration(config.Configuration.LogEnableStdout, true)

		err = config.SaveConfiguration(context)
		if err != nil {
			log.WithError(err).Error("main:main() Error processing WLS config: " + err.Error())
			log.Tracef("%+v", err)
			fmt.Fprintln(os.Stderr, "Error processing WLS config: "+err.Error())
			os.Exit(1)
		}

		setupRunner := &csetup.Runner{
			Tasks: []csetup.Task{
				csetup.Download_Ca_Cert{
					Flags:                flags,
					CmsBaseURL:           config.Configuration.CMS_BASE_URL,
					CaCertDirPath:        constants.TrustedCaCertsDir,
					TrustedTlsCertDigest: config.Configuration.CmsTlsCertDigest,
					ConsoleWriter:        os.Stdout,
				},
				csetup.Download_Cert{
					Flags:              flags,
					KeyFile:            constants.TLSKeyPath,
					CertFile:           constants.TLSCertPath,
					KeyAlgorithm:       constants.DefaultKeyAlgorithm,
					KeyAlgorithmLength: constants.DefaultKeyAlgorithmLength,
					CmsBaseURL:         config.Configuration.CMS_BASE_URL,
					Subject: pkix.Name{
						CommonName: config.Configuration.Subject.TLSCertCommonName,
					},
					SanList:       config.Configuration.CertSANList,
					CertType:      "TLS",
					CaCertsDir:    constants.TrustedCaCertsDir,
					BearerToken:   "",
					ConsoleWriter: os.Stdout,
				},
				setup.Server{
					Flags: flags,
				},
				setup.Database{
					Flags: flags,
				},
				setup.HVSConnection{
					Flags: flags,
				},
				setup.Download_Saml_Ca_Cert{
					Flags: flags,
				},
			},
			AskInput: false,
		}
		tasklist := []string{}
		if args[1] != "all" {
			tasklist = args[1:]
		}
		err = setupRunner.RunTasks(tasklist...)
		if err != nil {
			log.WithError(err).Error("main:main() Error in running setup tasks")
			log.Tracef("%+v", err)
			fmt.Fprintf(os.Stderr, "Error running setup tasks. %v\n", err.Error())
			os.Exit(1)
		} else {
			// when successful, we update the ownership of the config/certs updated by the setup tasks
			// all of them are likely to be found in /etc/workload-service/ path
			if config.TakeOwnershipFileWLS(constants.ConfigDir) != nil {
				fmt.Fprintln(os.Stderr, "Error: Failed to set permissions on WLS configuration files")
				os.Exit(-1)
			}
		}

	case "start":
		config.LogConfiguration(config.Configuration.LogEnableStdout, true)
		start()

	case "status":
		config.LogConfiguration(config.Configuration.LogEnableStdout, true)
		status()

	case "startserver":
		config.LogConfiguration(config.Configuration.LogEnableStdout, true)
		// this runs in attached mode
		startServer()

	case "stop":
		config.LogConfiguration(config.Configuration.LogEnableStdout, true)
		stop()

	case "uninstall":
		config.LogConfiguration(false, false)
		fmt.Println("Uninstalling workload-service...")
		stop()
		removeService()
		deleteFile("/opt/workload-service/")
		deleteFile("/usr/local/bin/workload-service")
		deleteFile("/var/log/workload-service/")
		if len(args) > 1 && strings.ToLower(args[1]) == "--purge" {
			deleteFile("/etc/workload-service/")
		}
		fmt.Println("workload-service successfully uninstalled")

	case "--version", "-v":
		printVersion()

	default:
		fmt.Printf("Unrecognized option : %s\n", arg)
		fallthrough

	case "help", "-help", "--help":
		printUsage()

	}
}

func start() error {
	log.Trace("main:start() Entering")
	defer log.Trace("main:start() Leaving")

	fmt.Println("Starting Workload Service")
	log.Info("main:start() Starting Workload Service")
	systemctl, err := exec.LookPath("systemctl")
	if err != nil {
		log.WithError(err).Error("main:start() Error trying to look up for systemctl path")
		log.Tracef("%+v", err)
		fmt.Fprintln(os.Stderr, "Error trying to look up for systemctl path")
		os.Exit(1)
	}
	return syscall.Exec(systemctl, []string{"systemctl", "start", "workload-service"}, os.Environ())
}

func stop() error {
	log.Trace("main:stop() Entering")
	defer log.Trace("main:stop() Leaving")

	fmt.Println("Stopping Workload Service")

	_, _, err := e.RunCommandWithTimeout("systemctl stop workload-service", 5)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not stop Workload-service")
		fmt.Println("Error : ", err)
		log.WithError(err).Error("main:stop() Could not stop Workload-service")
		log.Tracef("%+v", err)
	}
	return err
}

func status() error {
	log.Trace("main:status() Entering")
	defer log.Trace("main:status() Leaving")

	fmt.Println("Forwarding to systemctl status workload-service")
	log.Info("main:status() Workload-service status")
	systemctl, err := exec.LookPath("systemctl")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error trying to look up for systemctl path")
		log.WithError(err).Error("main:status() Error trying to look up for systemctl path")
		log.Tracef("%+v", err)
		os.Exit(1)
	}
	return syscall.Exec(systemctl, []string{"systemctl", "status", "workload-service"}, os.Environ())
}

func removeService() error {
	log.Trace("main:removeService() Entering")
	defer log.Trace("main:removeService() Leaving")

	fmt.Println("Removing Workload Service")
	log.Info("main:removeService() Removing Workload Service")
	_, _, err := e.RunCommandWithTimeout("systemctl disable workload-service", 5)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not remove Workload-service")
		fmt.Fprintf(os.Stderr, "Error: %v", err.Error())
		log.WithError(err).Error("main:removeService() Could not remove Workload-service")
		log.Tracef("%+v", err)
	}
	return err
}

func deleteFile(path string) {
	log.Trace("main:deleteFile() Entering")
	defer log.Trace("main:deleteFile() Leaving")

	fmt.Println("Deleting : ", path)
	log.Infof("main:deleteFile() Deleting : %s", path)
	// delete file
	var err = os.RemoveAll(path)
	if err != nil {
		fmt.Println(err.Error())
		log.WithError(err).Error("main:deleteFile() Error in deleting file")
		log.Tracef("%+v", err)
	}
}

func printUsage() {
	log.Trace("main:printUsage() Entering")
	defer log.Trace("main:printUsage() Leaving")

	fmt.Println("Usage:")
	fmt.Printf("    %s <command>\n\n", os.Args[0])
	fmt.Println("Available Commands:")
	fmt.Println("    help|-help|--help    Show this help message")
	fmt.Println("    -v|--version Print version/build information")
	fmt.Println("    start                Start workload-service")
	fmt.Println("    stop                 Stop workload-service")
	fmt.Println("    status               Determine if workload-service is running")
	fmt.Println("    uninstall  [--purge] Uninstall workload-service. --purge option needs to be applied to remove configuration and data files")
	fmt.Printf("    setup                Run workload-service setup tasks\n\n")

	fmt.Printf("Setup command usage:  %s <command> [task...]\n", os.Args[0])
	fmt.Println("Available tasks for setup:")
	fmt.Printf("    all                    Runs all setup tasks\n\n")
	fmt.Println("   download_ca_cert     Download CMS root CA certificate")
	fmt.Printf("\t\t                     - Option [--force] overwrites any existing files, and always downloads new root CA cert\n")
	fmt.Printf("\t\t                     Required env variables:\n")
	fmt.Println("                        - Environment variable CMS_BASE_URL=<url> for CMS API url")
	fmt.Println("                        - Environment variable CMS_TLS_CERT_SHA384=<CMS TLS cert sha384 hash> to ensure that WLS is talking to the right CMS instance")
	fmt.Printf("\n")

	fmt.Println("   download_cert TLS    Generates Key pair and CSR, gets it signed from CMS")
	fmt.Printf("\t\t                     - Option [--force] overwrites any existing files, and always downloads newly signed WLS TLS cert\n")
	fmt.Printf("\t\t                     Required env variables:\n")
	fmt.Println("                        - Environment variable CMS_BASE_URL=<url> for CMS API url")
	fmt.Println("                        - Environment variable BEARER_TOKEN=<token> for authenticating with CMS")
	fmt.Println("                        - Environment variable KEY_PATH=<key_path> Path of file where TLS key needs to be stored")
	fmt.Println("                        - Environment variable CERT_PATH=<cert_path> Path of file/directory where TLS certificate needs to be stored")
	fmt.Printf("\t\t                     Optional env variables:\n")
	fmt.Println("                        - Environment variable WLS_TLS_CERT_CN=<COMMON NAME> to override default specified in config")
	fmt.Println("                        - Environment variable SAN_LIST=<CSV List> List of FQDNs to be added to the SAN field in TLS cert to override default specified in config")
	fmt.Printf("\n")

	fmt.Println("    server              Setup http server on given port")
	fmt.Printf("\t\t                     - Option [--force] overwrites existing server config\n")
	fmt.Printf("                       -Environment variable WLS_PORT=<port> : WLS API listener port\n\n")
	fmt.Printf("\n")

	fmt.Println("    database            Setup workload-service database")
	fmt.Printf("\t\t                     - Option [--force] overwrites existing database config\n")
	fmt.Printf("\t\t                     Required env variables:\n")
	fmt.Println("                        - Environment variable WLS_DB_HOSTNAME  : database host name")
	fmt.Println("                        - Environment variable WLS_DB_PORT      : database port number")
	fmt.Println("                        - Environment variable WLS_DB           : database schema name")
	fmt.Println("                        - Environment variable WLS_DB_USERNAME  : database user name")
	fmt.Println("                        - Environment variable WLS_DB_PASSWORD  : database password")
	fmt.Println("                        - Environment variable WLS_DB_SSLMODE  : database SSL Connection Mode")
	fmt.Println("                        - Environment variable WLS_DB_SSLCERT  : database SSL Certificate target path")
	fmt.Printf("                        - Environment variable WLS_DB_SSLCERTSRC  : database SSL Certificate source path\n\n")
	fmt.Printf("\n")

	fmt.Println("    hvsconnection       Setup task for setting up the connection to the Host Verification Service(HVS)")
	fmt.Printf("\t\t                     - Option [--force] overwrites existing HVS config\n")
	fmt.Printf("                        - Environment variable HVS_URL=<url>      : HVS API Endpoint URL\n\n")
	fmt.Printf("\n")

	fmt.Printf("\n")
	fmt.Println("    download_saml_ca_cert   Setup to download SAML CA certificates from HVS")
	fmt.Printf("\t\t                     - Option [--force] overwrites existing HVS config\n")
	fmt.Println("                        - Environment variable HVS_URL=<url>      : HVS API Endpoint URL")
	fmt.Println("                        - Environment variable BEARER_TOKEN=<token> for authenticating with HVS")
}

func printVersion() {
	fmt.Printf("Workload Service Version %s\nBuild %s at %s - %s\n", version.Version, version.Branch, version.Time, version.GitHash)
}
