package queue

import (
	"fmt"
	"github.com/spf13/afero"
	"os"
)

// MasterQ has two tasks: first, it's a repository of registered task
// instances; second, it's the primary interface to the file system
// saving and retrieving of task execution instances.
type MasterQ struct {
	root       Path
	fs         afero.Fs
	tasks      map[string]TaskQueue
	permission os.FileMode
}

var globalQ map[string]*MasterQ

func init() {
	globalQ = make(map[string]*MasterQ, 0)
}

// New will retrieve the queue if the path already exists, or else
// create a new queue and register it.
// **rootDir** must be a valid path.
func New(root string, fs afero.Fs, perm os.FileMode) (*MasterQ, error) {

	rootDir := NewPath(root, fs, perm)
	rootDir = rootDir.Resolve()

	// if the queue already exists
	q, ok := globalQ[rootDir.String()]
	if ok {
		return q, nil
	}

	// if this is a new queue, first create
	// the directory
	err := rootDir.MkDirs()
	if err != nil {
		return nil, err
	}

	newMasterQ := &MasterQ{
		fs:         fs,
		root:       rootDir,
		tasks:      make(map[string]TaskQueue, 0),
		permission: perm,
	}

	globalQ[rootDir.String()] = newMasterQ

	return newMasterQ, nil
}

func (q *MasterQ) Has(name string) bool {
	_, ok := q.tasks[name]
	return ok
}

func (q *MasterQ) Get(name string) (TaskQueue, error) {
	t, ok := q.tasks[name]
	if ok {
		return t, nil
	}
	return TaskQueue{}, fmt.Errorf("task '%s' is not registered", name)
}

func (q *MasterQ) Enqueue(name string) TaskQueue {
	taskQ, err := q.Get(name)
	if err != nil {
		// TODO: yeah, really?
		panic(err)
	}
	return taskQ
}

// Register the given instance of the task interface. The task is registered
// by the derrived name.
func (q *MasterQ) Register(task TaskExecutor, name string) error {
	//name := GetTaskName(task)
	_, ok := q.tasks[name]
	if ok {
		return fmt.Errorf("tasks '%s' is already registered", name)
	}
	fmt.Printf("TASK REGISTERING: %s\n", name)
	newTaskQ, err := NewTaskQueue(q.root, name, task)
	if err != nil {
		return err
	}
	q.tasks[name] = newTaskQ
	return nil
}

func (q *MasterQ) RunAllTasks() []error {
	var errors []error
	c := make(chan ExecuteTaskErr, 5)
	for name, queue := range q.tasks {
		fmt.Println(name)
		tasks, err := queue.GetTaskInstances()
		if err != nil {
			errors = append(errors, err)
			continue
		}
		for _, task := range tasks {
			if !task.IsReady() {
				continue
			}
			go ExecuteTask(queue.task, task, c)
		}
	}
	close(c)
	return UnwrapChannel(c)
}
