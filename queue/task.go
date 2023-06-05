package queue

import (
	"encoding/json"
	"fmt"
	"time"
)

// TaskExecutor interface is what must be implemented
// for loq to register and execute a defined task.
type TaskExecutor interface {
	Assert(any) error
	Execute([]byte) error
}

//nolint:ireturn
func ReadTaskData[T any](jsonData []byte) (T, error) {
	var opts T
	err := json.Unmarshal(jsonData, &opts)
	if err != nil {
		return opts, err
	}
	return opts, nil
}

type TaskExecutionError struct {
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
	Traceback string    `json:"traceback"`
}

type TaskErrors struct {
	Errors []TaskExecutionError `json:"errors"`
}

func (te *TaskErrors) Count() int {
	return len(te.Errors)
}

func (te *TaskErrors) Add(err TaskExecutionError) {
	te.Errors = append(te.Errors, err)
}

func (te *TaskErrors) ReadFrom(errFile Path) error {
	if errFile.Exists() {
		data, _ := errFile.Read()
		err := json.Unmarshal(data, te)
		if err != nil {
			return err
		}
	}
	return nil
}

func (te *TaskErrors) WriteTo(errFile Path) error {
	data, err := json.Marshal(te)
	if err != nil {
		return err
	}
	err = errFile.Write(data)
	if err != nil {
		return err
	}
	return nil
}

const FOO = 100

// TaskInstance represents a unique task and is represented
// on a file system as a directory that contains at least a single
// json file with that task's execution arguments.
// The task directory can also contain a lock file and/or an
// error file.
type TaskInstance struct {
	root Path
	name string
	id   string
}

// Initialize creates the task directory and applies a lock.
func (ti TaskInstance) Initialize() error {
	err := ti.root.MkDirs()
	if err != nil {
		return err
	}
	return ti.ApplyLock()
}

// Actions

// Remove deletes the task folder and all it's files.
func (tq TaskInstance) Remove() error {
	return tq.root.Remove()
}

// ApplyLock create a .lock file in the task folder.
func (tq TaskInstance) ApplyLock() error {
	err := tq.LockFile().Write([]byte{})
	if err != nil {
		return err
	}
	return nil
}

// ReleaseLock deletes the .lock file in the task folder.
func (ti TaskInstance) ReleaseLock() error {
	if ti.LockFile().Exists() {
		return ti.LockFile().Remove()
	}
	return nil
}

func (tq TaskInstance) GetErrors() (TaskErrors, error) {
	errors := TaskErrors{}
	err := errors.ReadFrom(tq.ErrorFile())
	if err != nil {
		return errors, err
	}
	return errors, nil
}

// WriteError writes an error message to the tasks error file.
func (tq TaskInstance) WriteError(msg string, traceback string) error {
	errMsg := TaskExecutionError{
		Timestamp: time.Now(),
		Error:     msg,
		Traceback: traceback,
	}
	errFile := tq.ErrorFile()

	errors, err := tq.GetErrors()
	if err != nil {
		return err
	}

	errors.Add(errMsg)
	err = errors.WriteTo(errFile)
	if err != nil {
		return err
	}
	return nil
}

// Status

// Exists is true if the task folder exists.
func (tq TaskInstance) Exists() bool {
	return tq.TaskDir().Exists()
}

// IsReady is true if the task folder has a task file and
// is not locked and has no errors.
func (tq TaskInstance) IsReady() bool {
	return tq.TaskFile().Exists() && !tq.IsLocked() && !tq.HasError()
}

// IsLocked is true if the task folder contains a .lock file.
func (tq TaskInstance) IsLocked() bool {
	return tq.LockFile().Exists()
}

// HasError is true if the task folder contains a .error ffile.
func (tq TaskInstance) HasError() bool {
	return tq.ErrorFile().Exists()
}

// TaskDir returns the Path object for the task dir.
func (tq TaskInstance) TaskDir() Path {
	return tq.root
}

// TaskFile returns the Path object fof the task file.
func (tq TaskInstance) TaskFile() Path {
	return tq.TaskDir().Join(fmt.Sprintf("%s.json", tq.id))
}

// ErrorFile returns the Path object of the task error file.
func (tq TaskInstance) ErrorFile() Path {
	return tq.TaskDir().Join(fmt.Sprintf("%s.error", tq.id))
}

// LockFile returns the Path object of the task lock file.
func (tq TaskInstance) LockFile() Path {
	return tq.TaskDir().Join(fmt.Sprintf("%s.lock", tq.id))
}

func (tq TaskInstance) String() string {
	return tq.root.String()
}

type ExecuteTaskErr struct {
	Name  string
	Error error
}

func NewExecutTaskErr(name string, err error) ExecuteTaskErr {
	return ExecuteTaskErr{
		Name:  name,
		Error: err,
	}
}

func UnwrapChannel(c chan ExecuteTaskErr) []error {
	var errors []error
	for err := range c {
		errors = append(errors, err.Error)
	}
	if len(errors) > 0 {
		return errors
	}
	return nil
}

func ExecuteTask(task TaskExecutor, instance TaskInstance, c chan<- ExecuteTaskErr) {
	instance.LockFile()
	defer func() {
		err := instance.ReleaseLock()
		c <- NewExecutTaskErr(instance.name, err)
	}()

	data, err := instance.TaskFile().Read()
	if err != nil {
		c <- NewExecutTaskErr(instance.name, err)
		return
	}

	err = task.Execute(data)
	if err != nil {
		err = instance.WriteError(err.Error(), "")
		if err != nil {
			c <- NewExecutTaskErr(instance.name, err)
			return
		}
		return
	}

	err = instance.Remove()
	if err != nil {
		c <- NewExecutTaskErr(instance.name, err)
	}
	return
}
