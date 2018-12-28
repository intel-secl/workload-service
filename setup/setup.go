package setup

import "reflect"

type SetupTask interface {
	Setup() error
	Validate() error
}

func setupTasks() []SetupTask {
	return []SetupTask{
		SetupServer{},
	}
}

func RunSetupTasks(tasks ...string) error {
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
		var enabledTasks map[string]bool
		for _, t := range tasks {
			enabledTasks[t] = true
		}
		// iterate through the proper order of tasks, and execute the ones listed in the parameters
		for _, t := range setupTasks() {
			taskName := reflect.TypeOf(t).Name()
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
	return nil
}
