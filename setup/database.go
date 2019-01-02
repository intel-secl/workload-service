package setup

import (
	"errors"
	"fmt"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"
)

type Database struct{}

func setDbHostname() {

}

// Run will run the database setup tasks, but will skip if Validate() returns no error
func (ds Database) Run(c csetup.Context) error {
	if ds.Validate(c) == nil {
		fmt.Println("Database already setup, skipping ...")
		return nil
	}
	fmt.Println("Setting up database connection ...")
	var err error
	config.Configuration.Postgres.Hostname, err = c.GetConfigString(config.WLS_DB_HOSTNAME, "Database Hostname")
	if err != nil {
		return err
	}
	config.Configuration.Postgres.Port, err = c.GetConfigInt(config.WLS_DB_PORT, "Database Port")
	if err != nil {
		return err
	}
	config.Configuration.Postgres.User, err = c.GetConfigString(config.WLS_DB_USERNAME, "Database Username")
	if err != nil {
		return err
	}
	config.Configuration.Postgres.Password, err = c.GetConfigSecretString(config.WLS_DB_PASSWORD, "Database Password")
	if err != nil {
		return err
	}
	config.Configuration.Postgres.DBName, err = c.GetConfigString(config.WLS_DB, "Database Schema")
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
