package setup

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type SetupTask interface {
	Setup() error
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
			if err := t.Setup(); err != nil {
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
				if err := t.Setup(); err != nil {
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

func getSetupInt(env string, description string) int {
	fmt.Printf("Enter %s:\n", description)
	var intValue int
	if intStr, ok := os.LookupEnv(env); ok {
		val, err := strconv.ParseInt(intStr, 10, 32)
		if err == nil {
			fmt.Println(intStr)
			return int(val)
		}
	}
	for {
		if scanned, err := fmt.Scanf("%d", &intValue); scanned == 1 && err == nil {
			break
		}
		fmt.Printf("Error parsing %s, try again\n", description)
		fmt.Printf("Enter %s:", description)
	}
	return intValue
}

func getSetupString(env string, description string) string {
	fmt.Printf("Enter %s:\n", description)
	if str, ok := os.LookupEnv(env); ok {
		fmt.Println(str)
		return str
	}
	var str string
	fmt.Scanln(&str)
	return str
}

func getSetupSecretString(env string, description string) string {
	fmt.Printf("Enter %s:\n", description)
	if str, ok := os.LookupEnv(env); ok {
		fmt.Println("*****")
		return str
	}
	var str string
	fmt.Scanln(&str)
	return str
}
