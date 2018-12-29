package setup

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type SetupTask interface {
	Run() error
	Validate() error
}

func setupTasks() []SetupTask {
	return []SetupTask{
		Server{},
		Database{},
	}
}

func RunSetupTasks(tasks ...string) error {
	fmt.Println("Running setup ...")
	if len(tasks) == 0 {
		// run ALL the setup tasks
		for _, t := range setupTasks() {
			if err := t.Run(); err != nil {
				return err
			}
			if err := t.Validate(); err != nil {
				return err
			}
		}
	} else {
		// map each task ...string into a map[string]bool
		enabledTasks := make(map[string]bool)
		for _, t := range tasks {
			enabledTasks[strings.ToLower(t)] = true
		}
		// iterate through the proper order of tasks, and execute the ones listed in the parameters
		for _, t := range setupTasks() {
			taskName := strings.ToLower(reflect.TypeOf(t).Name())
			if _, ok := enabledTasks[taskName]; ok {
				if err := t.Run(); err != nil {
					return err
				}
				if err := t.Validate(); err != nil {
					return err
				}
			}
		}
	}
	fmt.Println("Setup finished successfully!")
	return nil
}

func getSetupInt(env string, description string) (int, error) {
	fmt.Printf("%s:\n", description)
	if intStr, ok := os.LookupEnv(env); ok {
		val, err := strconv.ParseInt(intStr, 10, 32)
		if err == nil {
			fmt.Println(intStr)
			return int(val), nil
		}
	}
	return 0, fmt.Errorf("%s is not defined", env)
}

func getSetupString(env string, description string) (string, error) {
	fmt.Printf("%s:\n", description)
	if str, ok := os.LookupEnv(env); ok {
		fmt.Println(str)
		return str, nil
	}
	return "", fmt.Errorf("%s is not defined", env)
}

func getSetupSecretString(env string, description string) (string, error) {
	fmt.Printf("%s:\n", description)
	if str, ok := os.LookupEnv(env); ok {
		fmt.Println("*****")
		return str, nil
	}
	return "", fmt.Errorf("%s is not defined", env)
}
