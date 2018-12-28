package setup

import (
	"errors"
	"fmt"
	"intel/isecl/workload-service/config"
)

type Database struct{}

func setDbHostname() {

}

// Setup will run the database setup tasks, but will skip if Validate() returns no error
func (ds Database) Setup() error {
	if ds.Validate() == nil {
		fmt.Println("Database already setup, skipping ...")
		return nil
	}
	fmt.Println("Setting up database connection ...")
	config.Configuration.Postgres.Hostname = getSetupString(config.WLS_DB_HOSTNAME, "Database Hostname")
	config.Configuration.Postgres.Port = getSetupInt(config.WLS_DB_PORT, "Database Port")
	config.Configuration.Postgres.User = getSetupString(config.WLS_DB_USERNAME, "Database Username")
	config.Configuration.Postgres.Password = getSetupSecretString(config.WLS_DB_PASSWORD, "Database Password")
	config.Configuration.Postgres.DBName = getSetupString(config.WLS_DB, "Database Schema")
	return config.Save()
}

// Validate checks whether or not the Database task was completed successfully
func (ds Database) Validate() error {
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
