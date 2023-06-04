package queue

import (
	"encoding/json"
	"fmt"
)

// TaskExecutor interface is what must be implemented
// for loq to register and execute a defined task
type TaskExecutor interface {
	Assert(any) error
	Execute([]byte) error
}

func ReadTaskData[T any](jsonData []byte) (T, error) {
	var opts T
	err := json.Unmarshal(jsonData, &opts)
	if err != nil {
		return opts, err
	}
	return opts, nil
}

// TaskInstance represents a unique task and is represented
// on a file system as a directory that contains at least a single
// json file with that task's execution arguments.
// The task directory can also contain a lock file and/or an
// error file.
type TaskInstance struct {
	task  TaskExecutor
	root  Path
	id    string
	runId int
	name  string
}

func (ti TaskInstance) Initialize() error {
	err := ti.root.MkDirs()
	if err != nil {
		return err
	}
	return ti.ApplyLock()
}

func (tq TaskInstance) Remove() error {
	return tq.root.Remove()
}

func (tq TaskInstance) Exists() bool {
	return tq.TaskDir().Exists()
}

func (tq TaskInstance) IsReady() bool {
	return tq.TaskFile().Exists() && !tq.IsLocked() && !tq.HasError()
}

func (tq TaskInstance) IsLocked() bool {
	return tq.LockFile().Exists()
}

func (tq TaskInstance) ApplyLock() error {
	err := tq.LockFile().Write([]byte{})
	if err != nil {
		return err
	}
	return nil
}

func (tq TaskInstance) ReleaseLock() error {
	return tq.LockFile().Remove()
}

func (tq TaskInstance) HasError() bool {
	return tq.ErrorFile().Exists()
}

func (tq TaskInstance) WriteError(msg string) error {
	return tq.ErrorFile().Write([]byte(msg))
}

func (tq TaskInstance) TaskDir() Path {
	return tq.root
}

func (tq TaskInstance) TaskFile() Path {
	return tq.TaskDir().Join(fmt.Sprintf("%s.json", tq.id))
}

func (tq TaskInstance) ErrorFile() Path {
	return tq.TaskDir().Join(fmt.Sprintf("%s.error", tq.id))
}

func (tq TaskInstance) LockFile() Path {
	return tq.TaskDir().Join(fmt.Sprintf("%s.lock", tq.id))
}

//func (tq TaskInstance) Execute(task TaskExecutor, taskFile Path) {
//	err := task.Execute(taskFile.String(), tq.runId)
//	if err != nil {
//		err = tq.WriteError(err.Error())
//		if err != nil {
//			// TODO: must be a better way to recover
//			panic(err)
//		}
//		return
//	}
//	err = tq.Remove()
//	if err != nil {
//		panic(err)
//	}
//}

func (tq TaskInstance) String() string {
	return tq.root.String()
}

//type TaskFinder interface {
//	Tasks() []TaskInstance
//	Handler(Path, fs.FileInfo, error) error
//}
//
//// TaskRunner handles assessing the status of a given task. It determines
//// the task's status and will execute if conditions allow (i.e. it's not
//// already locked by another process or has errors.
//type TaskRunner struct {
//	Q            *MasterQ
//	Id           int
//	LockedTasks  int
//	ErroredTasks int
//}
//
//func (tr *TaskRunner) Tasks() []TaskExecutor {
//	return nil
//}
//
//func (tr *TaskRunner) Handler(filePath Path, fileInfo fs.FileInfo, err error) error {
//	isDir, err := filePath.IsDir()
//	if err != nil {
//		return err
//	}
//	if isDir {
//		return nil
//	}
//
//	if strings.HasSuffix(filePath.Name(), ".json") {
//		tq := LoadTaskQueue(tr.Q, filePath, tr.Id)
//		if tq.IsLocked() {
//			tr.LockedTasks += 1
//			return fs.SkipDir
//		}
//		err := tq.ApplyLock()
//		if err != nil {
//			return fs.SkipDir
//		}
//
//		if tq.HasError() {
//			tr.ErroredTasks += 1
//			return fs.SkipDir
//		}
//
//		task, ok := tr.Q.tasks[tq.name]
//		if ok {
//			fmt.Printf("TASK RUN %d: %s\n", tr.Id, tq.String())
//			go tq.Execute(task, filePath)
//			fmt.Printf("TASK SUB %d: %s\n", tr.Id, tq.String())
//		}
//	}
//	return nil
//}
//
//// ErroredTaskFinder collects all tasks that have errored.
//type ErroredTaskFinder struct {
//	Q     *MasterQ
//	tasks []TaskInstance
//}
//
//func (tf *ErroredTaskFinder) Tasks() []TaskInstance {
//	return tf.tasks
//}
//
//func (tf *ErroredTaskFinder) Handler(filePath Path, fileInfo fs.FileInfo, err error) error {
//	isDir, err := filePath.IsDir()
//	if err != nil {
//		return err
//	}
//	if isDir {
//		return nil
//	}
//
//	if strings.HasSuffix(filePath.Name(), ".error") {
//		tq := LoadTaskQueue(tf.Q, filePath, 0)
//		tf.tasks = append(tf.tasks, tq)
//	}
//	return nil
//}
//
//// OrphanedTaskFinder collects all tasks that have a lock file
//// that is older than the number seconds in MaxAge.
//type OrphanedTaskFinder struct {
//	Q      *MasterQ
//	MaxAge float64
//	tasks  []TaskInstance
//}
//
//func (tf *OrphanedTaskFinder) Tasks() []TaskInstance {
//	return tf.tasks
//}
//
//func (tf *OrphanedTaskFinder) Handler(filePath Path, fileInfo fs.FileInfo, err error) error {
//	isDir, err := filePath.IsDir()
//	if err != nil {
//		return err
//	}
//	if isDir {
//		return nil
//	}
//
//	if strings.HasSuffix(filePath.Name(), ".lock") {
//		fileInfo, err := filePath.Stat()
//		if err != nil {
//			return err
//		}
//
//		fileAge := time.Now().Sub(fileInfo.ModTime())
//
//		if fileAge.Seconds() < tf.MaxAge {
//			return fs.SkipDir
//		}
//
//		tq := LoadTaskQueue(tf.Q, filePath, 0)
//		tf.tasks = append(tf.tasks, tq)
//	}
//	return nil
//}
