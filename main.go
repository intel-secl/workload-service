/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package main

import (
	"crypto/x509/pkix"
	"fmt"
	commLog "intel/isecl/lib/common/v3/log"
	csetup "intel/isecl/lib/common/v3/setup"
	"intel/isecl/lib/common/v3/validation"
	"intel/isecl/workload-service/v3/config"
	"intel/isecl/workload-service/v3/constants"
	"intel/isecl/workload-service/v3/setup"
	"intel/isecl/workload-service/v3/version"
	"os"
	"os/exec"
	"strings"
	"syscall"

	// Import Postgres driver
	e "intel/isecl/lib/common/v3/exec"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "intel/isecl/workload-service/v3/swagger/docs"
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
					KeyFile:            config.Configuration.TLSKeyFile,
					CertFile:           config.Configuration.TLSCertFile,
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

			if args[1] == "download_cert" {
				if config.TakeOwnershipFileWLS(config.Configuration.TLSKeyFile) != nil {
					fmt.Fprintln(os.Stderr, "Error: Failed to set permissions on TLS Key file")
					os.Exit(-1)
				}

				if config.TakeOwnershipFileWLS(config.Configuration.TLSCertFile) != nil {
					fmt.Fprintln(os.Stderr, "Error: Failed to set permissions on TLS Cert file")
					os.Exit(-1)
				}
			}
		}

	case "start":
		config.LogConfiguration(config.Configuration.LogEnableStdout, true)
		err := start()
		if err != nil {
			fmt.Println("Failed to start service")
		}

	case "status":
		config.LogConfiguration(config.Configuration.LogEnableStdout, true)
		err := status()
		if err != nil {
			fmt.Println("Failed to check service status")
		}

	case "startserver":
		config.LogConfiguration(config.Configuration.LogEnableStdout, true)
		// this runs in attached mode
		err := startServer()
		if err != nil {
			fmt.Println("Failed to start service")
		}

	case "stop":
		config.LogConfiguration(config.Configuration.LogEnableStdout, true)
		err := stop()
		if err != nil {
			fmt.Println("Failed to stop service")
		}

	case "uninstall":
		config.LogConfiguration(false, false)
		fmt.Println("Uninstalling workload-service...")
		err := stop()
		if err != nil {
			fmt.Println("Failed to stop service")
		}
		err = removeService()
		if err != nil {
			fmt.Println("Failed to remove service")
		}
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

	case "-h", "--help":
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

	fmt.Fprintln(os.Stdout, "Usage:")
	fmt.Fprintln(os.Stdout, "    workload-service <command> [arguments]")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Available Commands:")
	fmt.Fprintln(os.Stdout, "    -h|--help            Show this help message")
	fmt.Fprintln(os.Stdout, "    -v|--version         Print version/build information")
	fmt.Fprintln(os.Stdout, "    start                Start workload-service")
	fmt.Fprintln(os.Stdout, "    stop                 Stop workload-service")
	fmt.Fprintln(os.Stdout, "    status               Determine if workload-service is running")
	fmt.Fprintln(os.Stdout, "    uninstall [--purge]  Uninstall workload-service. --purge option needs to be applied to remove configuration and data files")
	fmt.Fprintln(os.Stdout, "    setup                Run workload-service setup tasks")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Setup command usage:     workload-service setup [task] [--force]")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "Available tasks for setup:")
	fmt.Fprintln(os.Stdout, "   all                              Runs all setup tasks")
	fmt.Fprintln(os.Stdout, "                                    Required env variables:")
	fmt.Fprintln(os.Stdout, "                                        - get required env variables from all the setup tasks")
	fmt.Fprintln(os.Stdout, "                                    Optional env variables:")
	fmt.Fprintln(os.Stdout, "                                        - get optional env variables from all the setup tasks")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "   download_ca_cert                 Download CMS root CA certificate")
	fmt.Fprintln(os.Stdout, "                                    - Option [--force] overwrites any existing files, and always downloads new root CA cert")
	fmt.Fprintln(os.Stdout, "                                    Required env variables if WLS_NOSETUP=true or variables not set in config.yml:")
	fmt.Fprintln(os.Stdout, "                                        - AAS_API_URL=<url>                            : AAS API url")
	fmt.Fprintln(os.Stdout, "                                        - HVS_URL=<url>                                : HVS API Endpoint URL")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_USERNAME=<service username>      : WLS service username")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_PASSWORD=<service password>      : WLS service password")
	fmt.Fprintln(os.Stdout, "                                    Required env variables specific to setup task are:")
	fmt.Fprintln(os.Stdout, "                                        - CMS_BASE_URL=<url>                              : for CMS API url")
	fmt.Fprintln(os.Stdout, "                                        - CMS_TLS_CERT_SHA384=<CMS TLS cert sha384 hash>  : to ensure that WLS is talking to the right CMS instance")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "   download_cert TLS                Generates Key pair and CSR, gets it signed from CMS")
	fmt.Fprintln(os.Stdout, "                                    - Option [--force] overwrites any existing files, and always downloads newly signed WLS TLS cert")
	fmt.Fprintln(os.Stdout, "                                    Required env variables if WLS_NOSETUP=true or variable not set in config.yml:")
	fmt.Fprintln(os.Stdout, "                                        - CMS_TLS_CERT_SHA384=<CMS TLS cert sha384 hash>  : to ensure that WLS is talking to the right CMS instance")
	fmt.Fprintln(os.Stdout, "                                        - AAS_API_URL=<url>                            : AAS API url")
	fmt.Fprintln(os.Stdout, "                                        - HVS_URL=<url>                                : HVS API Endpoint URL")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_USERNAME=<service username>      : WLS service username")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_PASSWORD=<service password>      : WLS service password")
	fmt.Fprintln(os.Stdout, "                                    Required env variables specific to setup task are:")
	fmt.Fprintln(os.Stdout, "                                        - CMS_BASE_URL=<url>                       : for CMS API url")
	fmt.Fprintln(os.Stdout, "                                        - BEARER_TOKEN=<token>                     : for authenticating with CMS")
	fmt.Fprintln(os.Stdout, "                                        - SAN_LIST=<CSV List>                      : List of FQDNs to be added to the SAN field in TLS cert to override default specified in config")
	fmt.Fprintln(os.Stdout, "                                    Optional env variables specific to setup task are:")
	fmt.Fprintln(os.Stdout, "                                        - KEY_PATH=<key_path>                      : Path of file where TLS key needs to be stored")
	fmt.Fprintln(os.Stdout, "                                        - CERT_PATH=<cert_path>                    : Path of file/directory where TLS certificate needs to be stored")
	fmt.Fprintln(os.Stdout, "                                        - WLS_TLS_CERT_CN=<COMMON NAME>            : to override default specified in config")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "   database                         Setup workload-service database")
	fmt.Fprintln(os.Stdout, "                                    - Option [--force] overwrites existing database config")
	fmt.Fprintln(os.Stdout, "                                    Required env variables if WLS_NOSETUP=true or variable not set in config.yml:")
	fmt.Fprintln(os.Stdout, "                                        - CMS_BASE_URL=<url>                              : for CMS API url")
	fmt.Fprintln(os.Stdout, "                                        - CMS_TLS_CERT_SHA384=<CMS TLS cert sha384 hash>  : to ensure that WLS is talking to the right CMS instance")
	fmt.Fprintln(os.Stdout, "                                        - AAS_API_URL=<url>                               : AAS API url")
	fmt.Fprintln(os.Stdout, "                                        - HVS_URL=<url>                                   : HVS API Endpoint URL")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_USERNAME=<service username>         : WLS service username")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_PASSWORD=<service password>         : WLS service password")
	fmt.Fprintln(os.Stdout, "                                    Required env variables specific to setup task are:")
	fmt.Fprintln(os.Stdout, "                                        - WLS_DB_HOSTNAME=<db host name>                   : database host name")
	fmt.Fprintln(os.Stdout, "                                        - WLS_DB_PORT=<db port>                            : database port number")
	fmt.Fprintln(os.Stdout, "                                        - WLS_DB=<db name>                                 : database schema name")
	fmt.Fprintln(os.Stdout, "                                        - WLS_DB_USERNAME=<db user name>                   : database user name")
	fmt.Fprintln(os.Stdout, "                                        - WLS_DB_PASSWORD=<db password>                    : database password")
	fmt.Fprintln(os.Stdout, "                                    Optional env variables specific to setup task are:")
	fmt.Fprintln(os.Stdout, "                                        - WLS_DB_SSLMODE=<db sslmode>                      : database SSL Connection Mode <disable|allow|prefer|require|verify-ca|verify-full>")
	fmt.Fprintln(os.Stdout, "                                        - WLS_DB_SSLCERT=<ssl certificate path>            : database SSL Certificate target path. Only applicable for WLS_DB_SSLMODE=<verify-ca|verify-full>. If left empty, the cert will be copied to /etc/workload-service/wlsdbsslcert.pem")
	fmt.Fprintln(os.Stdout, "                                        - WLS_DB_SSLCERTSRC=<ssl certificate source path>  : database SSL Certificate source path. Mandatory if WLS_DB_SSLCERT does not already exist")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "   server                           Setup http server on given port")
	fmt.Fprintln(os.Stdout, "                                    - Option [--force] overwrites existing server config")
	fmt.Fprintln(os.Stdout, "                                    Required env variables if WLS_NOSETUP=true or variable not set in config.yml:")
	fmt.Fprintln(os.Stdout, "                                        - CMS_BASE_URL=<url>                              : for CMS API url")
	fmt.Fprintln(os.Stdout, "                                        - CMS_TLS_CERT_SHA384=<CMS TLS cert sha384 hash>  : to ensure that WLS is talking to the right CMS instance")
	fmt.Fprintln(os.Stdout, "                                        - AAS_API_URL=<url>                               : AAS API url")
	fmt.Fprintln(os.Stdout, "                                        - HVS_URL=<url>                                   : HVS API Endpoint URL")
	fmt.Fprintln(os.Stdout, "                                    Optional env variables specific to setup task are:")
	fmt.Fprintln(os.Stdout, "                                        - WLS_PORT=<port>    : WLS API listener port")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_USERNAME=<service username>         : WLS service username")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_PASSWORD=<service password>         : WLS service password")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "   hvsconnection                    Setup task for setting up the connection to the Host Verification Service(HVS)")
	fmt.Fprintln(os.Stdout, "                                    - Option [--force] overwrites existing HVS config")
	fmt.Fprintln(os.Stdout, "                                    Required env variables if WLS_NOSETUP=true or variable not set in config.yml:")
	fmt.Fprintln(os.Stdout, "                                        - CMS_BASE_URL=<url>                              : for CMS API url")
	fmt.Fprintln(os.Stdout, "                                        - CMS_TLS_CERT_SHA384=<CMS TLS cert sha384 hash>  : to ensure that WLS is talking to the right CMS instance")
	fmt.Fprintln(os.Stdout, "                                        - AAS_API_URL=<url>                               : AAS API url")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_USERNAME=<service username>         : WLS service username")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_PASSWORD=<service password>         : WLS service password")
	fmt.Fprintln(os.Stdout, "                                    Required env variable specific to setup task is:")
	fmt.Fprintln(os.Stdout, "                                        - HVS_URL=<url>      : HVS API Endpoint URL")
	fmt.Fprintln(os.Stdout, "")
	fmt.Fprintln(os.Stdout, "   download_saml_ca_cert            Setup to download SAML CA certificates from HVS")
	fmt.Fprintln(os.Stdout, "                                    - Option [--force] overwrites existing certificate")
	fmt.Fprintln(os.Stdout, "                                    Required env variables if WLS_NOSETUP=true or variable not set in config.yml:")
	fmt.Fprintln(os.Stdout, "                                        - CMS_BASE_URL=<url>                              : for CMS API url")
	fmt.Fprintln(os.Stdout, "                                        - CMS_TLS_CERT_SHA384=<CMS TLS cert sha384 hash>  : to ensure that WLS is talking to the right CMS instance")
	fmt.Fprintln(os.Stdout, "                                        - AAS_API_URL=<url>                               : AAS API url")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_USERNAME=<service username>         : WLS service username")
	fmt.Fprintln(os.Stdout, "                                        - WLS_SERVICE_PASSWORD=<service password>         : WLS service password")
	fmt.Fprintln(os.Stdout, "                                    Required env variables specific to setup task are:")
	fmt.Fprintln(os.Stdout, "                                        - HVS_URL=<url>      : HVS API Endpoint URL")
	fmt.Fprintln(os.Stdout, "                                        - BEARER_TOKEN=<token> for authenticating with HVS")
}

func printVersion() {
	fmt.Printf("Workload Service %s-%s\nBuilt %s\n", version.Version, version.GitHash, version.BuildDate)
}
