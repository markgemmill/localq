package queue

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
)

type Queue struct {
	root       string
	tasks      map[string]Task
	permission os.FileMode
}

var Q *Queue

func GetQueue(rootDir string) *Queue {
	if Q != nil {
		return Q
	}

	q := Queue{
		root:       rootDir,
		tasks:      make(map[string]Task, 0),
		permission: 0755,
	}
	Q = &q

	return Q
}

func (q *Queue) getTaskName(task Task) string {
	name := reflect.TypeOf(task).String()
	name = strings.ToLower(name)
	name = strings.Replace(name, "-", "_", -1)
	name = strings.Replace(name, ".", "_", -1)
	name = strings.Replace(name, "*", "", -1)
	return name
}

func (q *Queue) makeQueueDir(pth ...string) (string, error) {
	dirPath := path.Join(q.root, path.Join(pth...))
	err := os.MkdirAll(dirPath, q.permission)
	if err != nil {
		return dirPath, err
	}
	return dirPath, nil
}

func (q *Queue) Initialize() error {
	_, err := q.makeQueueDir()
	return err
}

func (q *Queue) Has(name string) bool {
	_, ok := q.tasks[name]
	return ok
}

func (q *Queue) Get(name string) Task {
	t, ok := q.tasks[name]
	if ok {
		return t
	}
	return nil
}

func (q *Queue) Register(task Task) error {
	name := q.getTaskName(task)
	_, ok := q.tasks[name]
	if ok {
		return fmt.Errorf("tasks '%s' is already registered", name)
	}
	fmt.Printf("TASK REGISTERING: %s\n", name)
	q.tasks[name] = task
	return nil
}

func (q *Queue) Enqueue(taskType any) TaskQueue {
	name := q.getTaskName(taskType.(Task))
	//fmt.Printf("Enqueing: %s\n", name)
	_, ok := q.tasks[name]
	if !ok {
		panic(fmt.Errorf("task '%s' is not registered!", name))
	}
	return NewTaskQueue(q, name)
}

func (q *Queue) RunAllTasks(runId int) {
	runner := TaskRunner{Q: q, Id: runId}
	err := filepath.WalkDir(q.root, runner.HandleTaskFile)
	//fmt.Printf("TASKS IN USE: %d\n", runner.LockedTasks)
	//fmt.Printf("TASKS IN ERR: %d\n", runner.ErroredTasks)

	if err != nil {
		fmt.Println(err)
	}
}
