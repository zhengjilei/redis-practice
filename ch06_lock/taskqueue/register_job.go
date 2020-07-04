package taskqueue

import "errors"

var (
	task = map[string]func(){}
)

// RegisterJob ...
func RegisterJob(name string, callback func()) {
	task[name] = callback
}

// TriggerJob ...
func TriggerJob(name string) error {
	job, ok := task[name]
	if ok {
		job()
		return nil
	}
	return errors.New("non existed job")
}
