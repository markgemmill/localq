package queue

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/fs"
	"os"
	"path"
)

// Task interface is what must be implemented
// for loq to register and execute a defined task
type Task interface {
	//Name() string
	//Id() string
	Assert(any) error
	Execute(string, int) error
}

// TaskQueue represents a unique task entry
type TaskQueue struct {
	queue *Queue
	task  Task
	id    string
	runId int
	name  string
}

// Send validates the task options and serializes them to disk.
func (tq TaskQueue) Send(opt any) error {
	t := tq.queue.Get(tq.name)
	if t == nil {
		return fmt.Errorf("no registered task %s", tq.name)
	}

	err := t.Assert(opt)
	if err != nil {
		return fmt.Errorf("invalid %s task options: %w", tq.name, err)
	}

	// serialize the tasks arguments to json
	serializedTaskArgs, err := json.Marshal(opt)
	if err != nil {
		return err
	}

	// TODO should we be using tq.Path() here?
	qpth, err := tq.queue.makeQueueDir(tq.name, tq.id)
	if err != nil {
		return err
	}

	taskFilePath := path.Join(qpth, tq.TaskFileName())
	fmt.Printf("TASK ENQUEUED: %s\n", taskFilePath)

	// open file for writing task args
	file, err := os.Create(taskFilePath)
	defer func() {
		_ = file.Close()
	}()

	if err != nil {
		return err
	}

	_, err = file.Write(serializedTaskArgs)
	if err != nil {
		return err
	}

	return nil
}

func (tq TaskQueue) Remove() error {
	return os.RemoveAll(tq.Path())
}

func (tq TaskQueue) Path() string {
	// TODO: is this necessary - should this be the queue or the task-queue
	return path.Join(tq.queue.root, tq.name, tq.id)
}

func (tq TaskQueue) IsLocked() bool {
	pth := path.Join(tq.queue.root, tq.name, tq.id, tq.TaskLockFileName())
	_, err := os.Stat(pth)
	if err != nil {
		return false
	}
	return true
}

func (tq TaskQueue) ApplyLock() error {
	pth := path.Join(tq.queue.root, tq.name, tq.id, tq.TaskLockFileName())
	err := os.WriteFile(pth, []byte{}, tq.queue.permission)
	if err != nil {
		return err
	}
	return nil
}

func (tq TaskQueue) HasError() bool {
	pth := path.Join(tq.queue.root, tq.name, tq.id, tq.TaskErrorFileName())
	_, err := os.Stat(pth)
	if err != nil {
		return false
	}
	return true
}

func (tq TaskQueue) WriteError(msg string) error {
	pth := path.Join(tq.queue.root, tq.name, tq.id, tq.TaskErrorFileName())
	err := os.WriteFile(pth, []byte(msg), tq.queue.permission)
	if err != nil {
		return err
	}
	return nil
}

func (tq TaskQueue) TaskFileName() string {
	return fmt.Sprintf("%s.json", tq.id)
}

func (tq TaskQueue) TaskErrorFileName() string {
	return fmt.Sprintf("%s.error", tq.id)
}

func (tq TaskQueue) TaskLockFileName() string {
	return fmt.Sprintf("%s.lock", tq.id)
}

func (tq TaskQueue) Execute(task Task, taskFile string) {
	err := task.Execute(taskFile, tq.runId)
	if err != nil {
		err = tq.WriteError(err.Error())
		if err != nil {
			// TODO: must be a better way to recover
			panic(err)
		}
		return
	}
	err = tq.Remove()
	if err != nil {
		panic(err)
	}
}

func (tq TaskQueue) String() string {
	return fmt.Sprintf("%s %s", path.Base(tq.Path()), tq.name)
}

// NewTaskQueue initializes a new task with a unique id
func NewTaskQueue(queue *Queue, taskName string) TaskQueue {
	return TaskQueue{
		queue: queue,
		id:    uuid.New().String(),
		name:  taskName,
	}
}

// LoadTaskQueue "loads" a task from disk
func LoadTaskQueue(queue *Queue, taskPath string, runID int) TaskQueue {
	taskInstDir := path.Dir(taskPath)
	taskDir := path.Dir(taskInstDir)
	taskName := path.Base(taskDir)
	taskId := path.Base(taskInstDir)
	return TaskQueue{queue: queue, name: taskName, id: taskId, runId: runID}
}

// TaskRunner handles assessing the status of a given task. It determines
// the task's status and will execute if conditions allow (i.e. it's not
// already locked by another process or has errors.
type TaskRunner struct {
	Q            *Queue
	Id           int
	LockedTasks  int
	ErroredTasks int
}

func (tr *TaskRunner) HandleTaskFile(filePath string, dirEntry fs.DirEntry, err error) error {
	if path.Ext(filePath) == ".json" {
		tq := LoadTaskQueue(tr.Q, filePath, tr.Id)
		if tq.IsLocked() {
			tr.LockedTasks += 1
			return fs.SkipDir
		}
		err := tq.ApplyLock()
		if err != nil {
			return fs.SkipDir
		}

		if tq.HasError() {
			tr.ErroredTasks += 1
			return fs.SkipDir
		}

		task, ok := tr.Q.tasks[tq.name]
		if ok {
			fmt.Printf("TASK RUN %d: %s\n", tr.Id, tq.String())
			go tq.Execute(task, filePath)
			fmt.Printf("TASK SUB %d: %s\n", tr.Id, tq.String())
		}
	}
	return nil
}
