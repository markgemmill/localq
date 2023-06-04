package queue

import (
	"fmt"
	"github.com/spf13/afero"
	"os"
)

//// GetTaskName generates the unique name of a TaskExecutor that the
//// MasterQ uses to register and retrieve the instance.
//func GetTaskName(task TaskExecutor) string {
//	name := reflect.TypeOf(task).String()
//	name = strings.ToLower(name)
//	name = strings.Replace(name, "-", "_", -1)
//	name = strings.Replace(name, ".", "_", -1)
//	name = strings.Replace(name, "*", "", -1)
//	return name
//}

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

func (q *MasterQ) Get(name string) TaskQueue {
	t, ok := q.tasks[name]
	if ok {
		return t
	}
	return TaskQueue{}
}

// Register the given instance of the task interface. The task is registered
// by the derrived name
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

func (q *MasterQ) Enqueue(name string) TaskQueue {
	taskQ, ok := q.tasks[name]
	if !ok {
		panic(fmt.Errorf("task '%s' is not registered!", name))
	}
	return taskQ
}

//func (q *MasterQ) RunAllTasks(runId int) {
//	walk, err := pathlib.NewWalkWithOpts(q.root, &pathlib.WalkOpts{
//		Depth:           -1,
//		Algorithm:       pathlib.AlgorithmBasic,
//		FollowSymlinks:  false,
//		MinimumFileSize: -1,
//		MaximumFileSize: -1,
//		VisitFiles:      true,
//		VisitDirs:       true,
//		VisitSymlinks:   false,
//	})
//
//	runner := TaskRunner{Q: q, Id: runId}
//	err = walk.Walk(runner.Handler)
//
//	if err != nil {
//		fmt.Println(err)
//	}
//}

//func (q *MasterQ) FindTasks(finder TaskFinder) []TaskInstance {
//	walk, err := pathlib.NewWalkWithOpts(q.root, &pathlib.WalkOpts{
//		Depth:           -1,
//		Algorithm:       pathlib.AlgorithmBasic,
//		FollowSymlinks:  false,
//		MinimumFileSize: -1,
//		MaximumFileSize: -1,
//		VisitFiles:      true,
//		VisitDirs:       true,
//		VisitSymlinks:   false,
//	})
//
//	err = walk.Walk(finder.Handler)
//
//	if err != nil {
//		fmt.Println(err)
//		os.Exit(0)
//	}
//
//	return finder.Tasks()
//}
