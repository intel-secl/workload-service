package setup

import (
	"errors"
	"fmt"
	csetup "intel/isecl/lib/common/setup"
	"intel/isecl/workload-service/config"
)

// HVSConnection is a setup task for setting up the connection to the Host Verification Service (HVS)
type HVSConnection struct{}

// Run will run the HVS Connection setup task, but will skip if Validate() returns no errors
func (hvs HVSConnection) Run(c csetup.Context) error {
	if hvs.Validate(c) == nil {
		fmt.Println("HVS connection already setup, skipping ...")
		return nil
	}
	fmt.Println("Setting up HVS connection ...")
	var err error
	if config.Configuration.HVS.URL, err = c.GetConfigString(config.HVS_URL, "Key Management Server URL"); err != nil {
		return err
	}
	if config.Configuration.HVS.User, err = c.GetConfigString(config.HVS_USER, "Key Management Server User"); err != nil {
		return err
	}
	if config.Configuration.HVS.Password, err = c.GetConfigSecretString(config.HVS_PASSWORD, "Key Management Server Password"); err != nil {
		return err
	}
	return config.Save()
}

// Validate checks whether or not the HVS Connection setup task was completed successfully
func (hvs HVSConnection) Validate(c csetup.Context) error {
	h := &config.Configuration.HVS
	if h.URL == "" {
		return errors.New("HVS Connection: URL is not set")
	}
	if h.User == "" {
		return errors.New("HVS Connection: User is not set")
	}
	if h.Password == "" {
		return errors.New("HVS Connection: Password is not set ")
	}
	return nil
}
