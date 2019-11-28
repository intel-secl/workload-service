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
	config.LogConfiguration(false, true)
	/* PARSE COMMAND LINE OPTIONS */
	args := os.Args[1:]
	if len(args) <= 0 {
		fmt.Println("Command not found. Usage below ", os.Args[0])
		printUsage()
		return
	}

	inputStringArr := os.Args[0:]
	if err := validation.ValidateStrings(inputStringArr); err != nil {
		secLog.WithError(err).Error("main:main() Invalid Input")
		log.Tracef("%+v", err)
		fmt.Fprintln(os.Stderr, "Invalid input")
		printUsage()
		os.Exit(1)
	}

	switch arg := strings.ToLower(args[0]); arg {
	case "setup":
		config.LogConfiguration(false, true)
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
			args[1] != "database" &&
			args[1] != "hvsconnection" &&
			args[1] != "aasconnection" &&
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

		if len(args) > 1 {
			flags = args[2:]
			if args[1] == "download_cert" && len(args) > 2 {
				flags = args[3:]
			}
		}

		err = config.SaveConfiguration(context)
		if err != nil {
			log.WithError(err).Error("main:main() Unable to save configuration in config.yml")
			log.Tracef("%+v", err)
			fmt.Fprintln(os.Stderr, "Error: Unable to save configuration in config.yml")
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
						Country:      []string{config.Configuration.Subject.Country},
						Organization: []string{config.Configuration.Subject.Organization},
						Locality:     []string{config.Configuration.Subject.Locality},
						Province:     []string{config.Configuration.Subject.Province},
						CommonName:   config.Configuration.Subject.TLSCertCommonName,
					},
					SanList:       config.Configuration.CertSANList,
					CertType:      "TLS",
					CaCertsDir:    constants.TrustedCaCertsDir,
					BearerToken:   "",
					ConsoleWriter: os.Stdout,
				},
				new(setup.Server),
				new(setup.Database),
				new(setup.HVSConnection),
				new(setup.AASConnection),
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
			fmt.Fprintf(os.Stderr, "Error running setup tasks. %s\n", err.Error())
			os.Exit(1)
		}

	case "start":
		start()

	case "status":
		status()

	case "startserver":
		config.LogConfiguration(true, true)
		// this runs in attached mode
		startServer()

	case "stop":
		config.LogConfiguration(true, true)
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
	log.Info("main:stop() Stopping Workload Service")
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
	fmt.Println("    start                Start workload-service")
	fmt.Println("    stop                 Stop workload-service")
	fmt.Println("    status               Determine if workload-service is running")
	fmt.Println("    uninstall  [--purge] Uninstall workload-service. --purge option needs to be applied to remove configuration and data files")
	fmt.Printf("    setup all               Setup workload-service for use\n\n")
	fmt.Printf("Setup command usage:  %s <command> [task...]\n", os.Args[0])
	fmt.Println("Available tasks for setup:")
	fmt.Printf("    all                    Runs all setup tasks\n\n")
	fmt.Println("    download_ca_cert")
	fmt.Println("                        - Download CMS root CA certificate")
	fmt.Println("                        - Environment variable CMS_BASE_URL=<url> for CMS API url")
	fmt.Println("                        - Environment variable CMS_TLS_CERT_SHA384=<CMS TLS cert sha384 hash> to ensure that WLS is talking to the right CMS instance")
	fmt.Println("                        - Environment variable AAS_API_URL=<url> for AAS API url")
	fmt.Println("    download_cert TLS")
	fmt.Println("                        - Generates Key pair and CSR, gets it signed from CMS")
	fmt.Println("                        - Environment variable CMS_BASE_URL=<url> for CMS API url")
	fmt.Println("                        - Environment variable CMS_TLS_CERT_SHA384=<CMS TLS cert sha384 hash> to ensure that WLS is talking to the right CMS instance")
	fmt.Println("                        - Environment variable BEARER_TOKEN=<token> for authenticating with CMS")
	fmt.Println("                        - Environment variable KEY_PATH=<key_path> to override default specified in config")
	fmt.Println("                        - Environment variable CERT_PATH=<cert_path> to override default specified in config")
	fmt.Println("                        - Environment variable WLS_TLS_CERT_CN=<COMMON NAME> to override default specified in config")
	fmt.Println("                        - Environment variable WLS_CERT_ORG=<CERTIFICATE ORGANIZATION> to override default specified in config")
	fmt.Println("                        - Environment variable WLS_CERT_COUNTRY=<CERTIFICATE COUNTRY> to override default specified in config")
	fmt.Println("                        - Environment variable WLS_CERT_LOCALITY=<CERTIFICATE LOCALITY> to override default specified in config")
	fmt.Println("                        - Environment variable WLS_CERT_PROVINCE=<CERTIFICATE PROVINCE> to override default specified in config")
	fmt.Println("                        - Environment variable WLS_CERT_SAN=<CSV List of alternative names to be added to the SAN field in TLS cert> to override default specified in config")
	fmt.Println("    server              Setup http server on given port")
	fmt.Printf("                        -Environment variable WLS_PORT=<port> should be set\n\n")
	fmt.Println("    database            Setup workload-service database")
	fmt.Println("                        Required env variables are:")
	fmt.Println("                        - WLS_DB_HOSTNAME  : database host name")
	fmt.Println("                        - WLS_DB_PORT      : database port number")
	fmt.Println("                        - WLS_DB_USERNAME  : database user name")
	fmt.Println("                        - WLS_DB_PASSWORD  : database password")
	fmt.Printf("                         - WLS_DB           : database schema name\n\n")
	fmt.Println("    hvsconnection       Setup task for setting up the connection to the Host Verification Service(HVS)")
	fmt.Println("                        Required env variables are:")
	fmt.Printf("                        - HVS_URL      : HVS URL\n\n")
	fmt.Println("    aasconnection       Setup to create workload service user roles in AAS")
	fmt.Println("                        - AAS_API_URL      : AAS API URL")
	fmt.Println("                        - BEARER_TOKEN     : Bearer Token for authenticating with AAS")
}
