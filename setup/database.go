/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	"flag"
	"fmt"
	cos "intel/isecl/lib/common/v2/os"
	csetup "intel/isecl/lib/common/v2/setup"
	"intel/isecl/lib/common/v2/validation"
	"intel/isecl/workload-service/v2/config"
	"intel/isecl/workload-service/v2/constants"
	"os"
	"strings"

	"intel/isecl/workload-service/v2/repository/postgres"

	"github.com/pkg/errors"
)

// Database is a setup task for setting up the Postgres connection to use for WLS
// it expects you to set WLS_DB_HOSTNAME, WLS_DB_PORT, WLS_DB_USERNAME, WLS_DB_PASSWORD, and WLS_DB
type Database struct {
	Flags []string
}

// Run will run the database setup tasks, but will skip if Validate() returns no error
func (ds Database) Run(c csetup.Context) error {
	log.Trace("setup/database:Run() Entering")
	defer log.Trace("setup/database:Run() Leaving")
	var err error

	fmt.Println("Running setup task: database")

	fs := flag.NewFlagSet("database", flag.ExitOnError)
	force := fs.Bool("force", false, "force recreation, will overwrite any existing certificate")

	err = fs.Parse(ds.Flags)
	if err != nil {
		fmt.Println("WLS Database setup: Unable to parse flags")
		return fmt.Errorf("WLS Database setup: Unable to parse flags")
	}

	// task only runs if force flag is unset or
	if !*force && ds.Validate(c) == nil {
		fmt.Println("setup database: task already complete. Skipping...")
		log.Info("setup/database:Run() WLS Database already setup, skipping ...")
		return nil
	}

	log.Info("setup/database:Run() Setting up database connection ...")

	config.Configuration.Postgres.Hostname, _ = c.GetenvString("WLS_DB_HOSTNAME", "Database Hostname")
	config.Configuration.Postgres.Port, _ = c.GetenvInt("WLS_DB_PORT", "Database Port")
	config.Configuration.Postgres.UserName, _ = c.GetenvString("WLS_DB_USERNAME", "Database Username")
	config.Configuration.Postgres.Password, _ = c.GetenvSecret("WLS_DB_PASSWORD", "Database Password")
	config.Configuration.Postgres.DBName, _ = c.GetenvString("WLS_DB", "Database Name")
	config.Configuration.Postgres.SSLMode, _ = c.GetenvString("WLS_DB_SSLMODE", "Database SSLMode")
	config.Configuration.Postgres.SSLCert, _ = c.GetenvString("WLS_DB_SSLCERT", "Database SSL Certificate")
	envDBSSLCertSrc, _ := c.GetenvString("WLS_DB_SSLCERTSRC", "Database SSL Certificate source file")

	var validErr error

	validErr = validation.ValidateHostname(config.Configuration.Postgres.Hostname)
	if validErr != nil {
		return errors.Wrap(validErr, "setup database: Validation fail")
	}
	validErr = validation.ValidateAccount(config.Configuration.Postgres.UserName, config.Configuration.Postgres.Password)
	if validErr != nil {
		return errors.Wrap(validErr, "setup database: Validation fail")
	}
	validErr = validation.ValidateIdentifier(config.Configuration.Postgres.DBName)
	if validErr != nil {
		return errors.Wrap(validErr, "setup database: Validation fail")
	}

	config.Configuration.Postgres.SSLMode, config.Configuration.Postgres.SSLCert, validErr = configureDBSSLParams(
		config.Configuration.Postgres.SSLMode, envDBSSLCertSrc,
		config.Configuration.Postgres.SSLCert)
	if validErr != nil {
		return errors.Wrap(validErr, "setup database: Validation fail")
	}

	log.Info("setup/database:Run() Database connection updated in config")
	return config.Save()
}

func configureDBSSLParams(sslMode, sslCertSrc, sslCert string) (string, string, error) {
	sslMode = strings.TrimSpace(strings.ToLower(sslMode))
	sslCert = strings.TrimSpace(sslCert)
	sslCertSrc = strings.TrimSpace(sslCertSrc)

	if sslMode != "allow" && sslMode != "prefer" && sslMode != "verify-ca" && sslMode != "require" {
		sslMode = "verify-full"
	}

	if sslMode == "verify-ca" || sslMode == "verify-full" {
		// cover different scenarios
		if sslCertSrc == "" && sslCert != "" {
			if _, err := os.Stat(sslCert); os.IsNotExist(err) {
				return "", "", errors.Wrapf(err, "certificate source file not specified and sslcert %s does not exist", sslCert)
			}
			return sslMode, sslCert, nil
		}
		if sslCertSrc == "" {
			return "", "", errors.New("verify-ca or verify-full needs a source cert file to copy from unless db-sslcert exists")
		} else {
			if _, err := os.Stat(sslCertSrc); os.IsNotExist(err) {
				return "", "", errors.Wrapf(err, "certificate source file not specified and sslcert %s does not exist", sslCertSrc)
			}
		}
		// at this point if sslCert destination is not passed it, lets set to default
		if sslCert == "" {
			sslCert = constants.DefaultSSLCertFilePath
		}
		// lets try to copy the file now. If copy does not succeed return the file copy error
		if err := cos.Copy(sslCertSrc, sslCert); err != nil {
			return "", "", errors.Wrap(err, "failed to copy file")
		}
		// set permissions so that non root users can read the copied file
		if err := os.Chmod(sslCert, 0644); err != nil {
			return "", "", errors.Wrapf(err, "could not apply permissions to %s", sslCert)
		}
	}
	return sslMode, sslCert, nil
}

// Validate checks whether or not the Database task was completed successfully
func (ds Database) Validate(c csetup.Context) error {
	log.Trace("setup/database:Validate() Entering")
	defer log.Trace("setup/database:Validate() Leaving")

	if config.Configuration.Postgres.Hostname == "" {
		return errors.New("Database: Hostname is not set in configuration")
	}
	if config.Configuration.Postgres.Port == 0 {
		return errors.New("Database: Port is not set in configuration")
	}
	if config.Configuration.Postgres.UserName == "" {
		return errors.New("Database: User is not set in configuration")
	}
	if config.Configuration.Postgres.Password == "" {
		return errors.New("Database: Password is not set in configuration")
	}
	if config.Configuration.Postgres.DBName == "" {
		return errors.New("Database: Schema is not set in configuration")
	}

	// let's test the configuration by making a connection to the DB instance
	wlsDB, err := postgres.Open(config.Configuration.Postgres.Hostname, config.Configuration.Postgres.Port, config.Configuration.Postgres.DBName,
		config.Configuration.Postgres.UserName, config.Configuration.Postgres.Password, config.Configuration.Postgres.SSLMode, config.Configuration.Postgres.SSLCert)
	if err != nil {
		return errors.Wrap(err,"setup/database:Validate() Failed to connect to database with the provided configuration")
	}
	defer wlsDB.Close()

	return nil
}
