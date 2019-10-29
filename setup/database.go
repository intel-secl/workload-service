/*
 * Copyright (C) 2019 Intel Corporation
 * SPDX-License-Identifier: BSD-3-Clause
 */
package setup

import (
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"

	"github.com/pkg/errors"
)

// Database is a setup task for setting up the Postgres connection to use for WLS
// it expects you to set WLS_DB_HOSTNAME, WLS_DB_PORT, WLS_DB_USERNAME, WLS_DB_PASSWORD, and WLS_DB
type Database struct{}

// Run will run the database setup tasks, but will skip if Validate() returns no error
func (ds Database) Run(c csetup.Context) error {
	log.Trace("setup/database:Run() Entering")
	defer log.Trace("setup/database:Run() Leaving")
	if ds.Validate(c) == nil {
		log.Info("setup/database:Run() Database already setup, skipping ...")
		return nil
	}
	log.Info("setup/database:Run() Setting up database connection ...")
	var err error
	config.Configuration.Postgres.Hostname, err = c.GetenvString(config.WLS_DB_HOSTNAME, "Database Hostname")
	if err != nil {
		return errors.Wrap(err, "setup/database:Run() WLS_DB_HOSTNAME not set in environment")
	}
	config.Configuration.Postgres.Port, err = c.GetenvInt(config.WLS_DB_PORT, "Database Port")
	if err != nil {
		return errors.Wrap(err, "setup/database:Run() WLS_DB_PORT not set in environment")
	}
	config.Configuration.Postgres.User, err = c.GetenvString(config.WLS_DB_USERNAME, "Database Username")
	if err != nil {
		return errors.Wrap(err, "setup/database:Run() WLS_DB_USERNAME not set in environment")
	}
	config.Configuration.Postgres.Password, err = c.GetenvSecret(config.WLS_DB_PASSWORD, "Database Password")
	if err != nil {
		return errors.Wrap(err, "setup/database:Run() WLS_DB_PASSWORD not set in environment")
	}
	config.Configuration.Postgres.DBName, err = c.GetenvString(config.WLS_DB, "Database Schema")
	if err != nil {
		return errors.Wrap(err, "setup/database:Run() WLS_DB not set in environment")
	}
	log.Info("setup/database:Run() Database connection updated in config")
	return config.Save()
}

// Validate checks whether or not the Database task was completed successfully
func (ds Database) Validate(c csetup.Context) error {
	log.Trace("setup/database:Validate() Entering")
	defer log.Trace("setup/database:Validate() Leaving")
	if config.Configuration.Postgres.Hostname == "" {
		return errors.New("setup/database:Validate() Database: Hostname is not set")
	}
	if config.Configuration.Postgres.Port == 0 {
		return errors.New("setup/database:Validate() Database: Port is not set")
	}
	if config.Configuration.Postgres.User == "" {
		return errors.New("setup/database:Validate() Database: User is not set")
	}
	if config.Configuration.Postgres.Password == "" {
		return errors.New("setup/database:Validate() Database: Password is not set")
	}
	if config.Configuration.Postgres.DBName == "" {
		return errors.New("setup/database:Validate() Database: Schema is not set")
	}

	return nil
}
