package setup

import (
	"errors"
	"fmt"
	"intel/isecl/workload-service/config"
)

type Server struct{}

// Setup will configure the parameters for the WLS web service layer. This will be skipped if Validate() returns no errors
func (ss Server) Setup() error {
	if ss.Validate() == nil {
		fmt.Println("Webserver already setup, skipping ...")
		return nil
	}
	fmt.Println("Setting up webserver ...")
	config.Configuration.Port = getSetupInt(config.WLS_PORT, "Webserver Port")
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
