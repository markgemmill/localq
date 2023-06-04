package queue

import (
	"encoding/json"
	"fmt"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

type TaskHandler func(instance TaskInstance)

func NewTaskId() string {
	return gonanoid.Must(15)
}

type TaskQueue struct {
	root Path
	name string
	task TaskExecutor
}

func (tq TaskQueue) Initialize() error {
	return tq.root.MkDirs()
}

func (tq TaskQueue) Task() TaskExecutor {
	return tq.task
}

func (tq TaskQueue) CreateTaskInstance() TaskInstance {
	id := NewTaskId()
	ti := TaskInstance{
		id:   id,
		name: tq.name,
		root: tq.root.Join(id),
		task: tq.task,
	}
	return ti
}

func (tq TaskQueue) LoadTaskInstance(taskDir Path) TaskInstance {
	return TaskInstance{
		id:   taskDir.Name(),
		name: taskDir.Parent().Name(),
		root: taskDir,
		task: tq.task,
	}
}

func (tq TaskQueue) IterTaskInstances(handler TaskHandler) error {
	dirs, err := tq.root.ReadDir()
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		fmt.Println(dir.Name())
		taskInst := tq.LoadTaskInstance(dir)
		handler(taskInst)
	}
	return nil
}

// Send creates a new TaskInstance on disk with the given
// task arguments.
func (tq TaskQueue) Send(opt any) (TaskInstance, error) {

	err := tq.task.Assert(opt)
	if err != nil {
		return TaskInstance{}, fmt.Errorf("invalid %s task options: %w", tq.name, err)
	}

	// serialize the tasks arguments to json
	serializedTaskArgs, err := json.Marshal(opt)
	if err != nil {
		return TaskInstance{}, err
	}

	ti := tq.CreateTaskInstance()
	err = ti.Initialize()
	if err != nil {
		return ti, err
	}

	err = ti.TaskFile().Write(serializedTaskArgs)
	if err != nil {
		return ti, err
	}

	err = ti.ReleaseLock()
	if err != nil {
		return ti, err
	}

	return ti, nil
}

// Run creates a new TaskInstance on disk with the given
// task arguments, and then immediately executes.
func (tq TaskQueue) Run(opt any) (TaskInstance, error) {
	ti, err := tq.Send(opt)
	if err != nil {
		return ti, err
	}

	err = tq.execute(ti)

	if err != nil {
		return ti, err
	}

	return ti, nil
}

func (tq TaskQueue) execute(ti TaskInstance) error {
	ti.LockFile()
	defer func() {
		// TODO: returning from a defer???
		_ = ti.ReleaseLock()
	}()

	data, err := ti.TaskFile().Read()
	if err != nil {
		return err
	}
	err = tq.task.Execute(data)
	if err != nil {
		err = ti.WriteError(err.Error())
		if err != nil {
			// TODO: must be a better way to recover
			panic(err)
		}
		return nil
	}
	err = ti.Remove()
	if err != nil {
		panic(err)
	}
	return nil
}

func NewTaskQueue(master Path, name string, task TaskExecutor) (TaskQueue, error) {
	tq := TaskQueue{
		root: master.Join(name),
		name: name,
		task: task,
	}
	err := tq.Initialize()
	if err != nil {
		return tq, err
	}
	return tq, nil
}
