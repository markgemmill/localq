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

//nolint:ireturn
func (tq TaskQueue) Task() TaskExecutor {
	return tq.task
}

func (tq TaskQueue) CreateTaskInstance() TaskInstance {
	id := NewTaskId()
	ti := TaskInstance{
		id:   id,
		name: tq.name,
		root: tq.root.Join(id),
	}
	return ti
}

func (tq TaskQueue) LoadTaskInstance(taskDir Path) TaskInstance {
	return TaskInstance{
		id:   taskDir.Name(),
		name: taskDir.Parent().Name(),
		root: taskDir,
	}
}

func (tq TaskQueue) GetTaskInstances() ([]TaskInstance, error) {
	tasks := []TaskInstance{}

	dirs, err := tq.root.ReadDir()
	if err != nil {
		return tasks, err
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		fmt.Println(dir.Name())
		taskInst := tq.LoadTaskInstance(dir)
		tasks = append(tasks, taskInst)
	}
	return tasks, nil
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

	c := make(chan ExecuteTaskErr, 5)
	ExecuteTask(tq.task, ti, c)
	close(c)

	errors := UnwrapChannel(c)

	if errors != nil {
		return ti, errors[0]
	}

	return ti, nil
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
