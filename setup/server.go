package setup

import (
	"errors"
	"fmt"
	"intel/isecl/workload-service/config"
)

type Server struct{}

// Run will configure the parameters for the WLS web service layer. This will be skipped if Validate() returns no errors
func (ss Server) Run() error {
	if ss.Validate() == nil {
		fmt.Println("Webserver already setup, skipping ...")
		return nil
	}
	fmt.Println("Setting up webserver ...")
	var err error
	config.Configuration.Port, err = getSetupInt(config.WLS_PORT, "Webserver Port")
	if err != nil {
		fmt.Println("Using default webserver port: 5000")
		config.Configuration.Port = 5000
	}
	return config.Save()
}

// Validate checks whether or not the Server task configured successfully or not
func (ss Server) Validate() error {
	// validate that the port variable is not the zero value of its type
	if config.Configuration.Port == 0 {
		return errors.New("Server: Port is not set")
	}
	return nil
}
