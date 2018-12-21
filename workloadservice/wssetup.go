package workloadservice

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

type SetupTask interface {
	Setup() error
	Validate() error
}

type SampleSetupTask struct{}

func (SampleSetupTask) Setup() error {
	fmt.Println("SampleSetupTask Setup Done")
	return nil
}

func (SampleSetupTask) Validate() error {
	fmt.Println("SampleSetupTask Validate")
	return nil
}

func GetSetupTasks(commandargs []string) map[string]SetupTask {

	//tasks = ParseSetupTasks(commandargs)
	if len(commandargs) < 1 || strings.ToLower(commandargs[0]) != "setup" {
		panic (fmt.Errorf("GetSetupTasks need at least one parameter with command \"setup\". Arguments : %v\n", commandargs))
	}

	m := make(map[string]SetupTask)

	if len(commandargs) > 1  {
		// Todo - we should be able to find structs using reflection in this
		// package that implements the SetupTask Interface and add elements to theNo arguments passed in
		//  map. For now, we are just going to hardcode the setup tasks that we have

		// First argument is "setup" - the rest should be list of tasks
		for _, task := range commandargs[1:] {
			switch strings.ToLower(task) {
			case "SampleSetupTask":
				m["SampleSetupTask"] = SampleSetupTask{}
			default:
				log.Printf("Unknown Setup Task in list : %s", task)
			}
		}
	} else {
		fmt.Println("No arguments passed in")
		// no specific tasks passed in. We will return a list of all tasks
		m[reflect.TypeOf(SampleSetupTask{}).Name()] = SampleSetupTask{}
	}
	return m
}

func ParseSetupTasks(commandargs ...[]string) []string{
	if len(commandargs) > 1{
		log.Println("Expecting a slice of string as argument.")
	}
	fmt.Println(commandargs)
	return commandargs[0]
}

func RunTasks(commandargs []string ){
}
