package setup

import (
	"errors"
	"fmt"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"
)

// Database is a setup task for setting up the Postgres connection to use for WLS
// it expects you to set WLS_DB_HOSTNAME, WLS_DB_PORT, WLS_DB_USERNAME, WLS_DB_PASSWORD, and WLS_DB
type Database struct{}

// Run will run the database setup tasks, but will skip if Validate() returns no error
func (ds Database) Run(c csetup.Context) error {
	if ds.Validate(c) == nil {
		fmt.Println("Database already setup, skipping ...")
		return nil
	}
	fmt.Println("Setting up database connection ...")
	var err error
	config.Configuration.Postgres.Hostname, err = c.GetenvString(config.WLS_DB_HOSTNAME, "Database Hostname")
	if err != nil {
		return err
	}
	config.Configuration.Postgres.Port, err = c.GetenvInt(config.WLS_DB_PORT, "Database Port")
	if err != nil {
		return err
	}
	config.Configuration.Postgres.User, err = c.GetenvString(config.WLS_DB_USERNAME, "Database Username")
	if err != nil {
		return err
	}
	config.Configuration.Postgres.Password, err = c.GetenvSecret(config.WLS_DB_PASSWORD, "Database Password")
	if err != nil {
		return err
	}
	config.Configuration.Postgres.DBName, err = c.GetenvString(config.WLS_DB, "Database Schema")
	if err != nil {
		return err
	}
	return config.Save()
}

// Validate checks whether or not the Database task was completed successfully
func (ds Database) Validate(c csetup.Context) error {
	if config.Configuration.Postgres.Hostname == "" {
		return errors.New("Database: Hostname is not set")
	}
	if config.Configuration.Postgres.Port == 0 {
		return errors.New("Database: Port is not set")
	}
	if config.Configuration.Postgres.User == "" {
		return errors.New("Database: User is not set")
	}
	if config.Configuration.Postgres.Password == "" {
		return errors.New("Database: Password is not set")
	}
	if config.Configuration.Postgres.DBName == "" {
		return errors.New("Database: Schema is not set")
	}

	return nil
}
