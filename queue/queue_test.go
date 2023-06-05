package queue

import (
	"fmt"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type TaskOptions struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type ConcreteTask struct {
	Executed bool
	Errored  bool
}

func (t *ConcreteTask) Assert(opt any) error {
	options := opt.(TaskOptions)
	if options.Id == 0 {
		return fmt.Errorf("TaskOptions.Id is empty")
	}
	if options.Name == "" {
		return fmt.Errorf("TaskOptions.Name is empty")
	}
	return nil
}

func (t *ConcreteTask) Execute(jsonData []byte) error {
	fmt.Println("ConcreteTask.Execute....")
	options, err := ReadTaskData[TaskOptions](jsonData)
	fmt.Printf("ConcreteTask.Execute options: %v\n", options)
	if err != nil {
		fmt.Printf("ConcreteTask.Execute options err: %s\n", err)
		return err
	}
	t.Executed = true
	if t.Errored {
		return fmt.Errorf("ConcreteTask %d failed", options.Id)
	}
	fmt.Printf("ConcreteTask.Execute completed: %v\n", t.Executed)
	return nil
}

func MakeTaskQueue(err bool) *TaskQueue {
	tq, _ := NewTaskQueue(
		NewPath("/localq", afero.NewMemMapFs(), 07777),
		"concrete",
		&ConcreteTask{Errored: err},
	)
	return &tq
}

func TestTaskQueue_Initialize(t *testing.T) {
	tq := MakeTaskQueue(false)
	err := tq.Initialize()
	assert.Nil(t, err)
	assert.True(t, tq.root.Exists())

}

func TestTaskQueue_CreateTaskInstance(t *testing.T) {
	tq := MakeTaskQueue(false)
	err := tq.Initialize()
	assert.Nil(t, err)
	assert.True(t, tq.root.Exists())

	ti := tq.CreateTaskInstance()
	assert.Equal(t, "concrete", ti.name)
	assert.Equal(t, 15, len(ti.id), fmt.Sprintf("%s has a len of %d", ti.id, len(ti.id)))
	assert.Equal(t, fmt.Sprintf("/localq/concrete/%s", ti.id), ti.root.String())
	//assert.Equal(t, "*queue.ConcreteTask", reflect.TypeOf(ti.task).String())
	assert.False(t, ti.IsLocked())
}

func TestTaskQueue_LoadTaskInstance(t *testing.T) {
	tq := MakeTaskQueue(false)
	err := tq.Initialize()
	assert.Nil(t, err)
	assert.True(t, tq.root.Exists())

	ti := tq.CreateTaskInstance()
	ti2 := tq.LoadTaskInstance(ti.root)

	assert.Equal(t, ti.name, ti2.name)
	assert.Equal(t, ti.root, ti2.root)
	assert.Equal(t, ti.id, ti2.id)
	//assert.Equal(t, ti.task, ti2.task)
}

func TestTaskQueue_IterTaskInstances(t *testing.T) {

	tq := MakeTaskQueue(false)
	err := tq.Initialize()
	assert.Nil(t, err)
	assert.True(t, tq.root.Exists())

	_ = tq.CreateTaskInstance().Initialize()
	_ = tq.CreateTaskInstance().Initialize()
	_ = tq.CreateTaskInstance().Initialize()

	taskCount := 0

	err = tq.IterTaskInstances(func(t TaskInstance) {
		taskCount += 1
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, taskCount)

}

func TestTaskQueue_Send(t *testing.T) {

	tq := MakeTaskQueue(false)
	err := tq.Initialize()
	assert.Nil(t, err)
	assert.True(t, tq.root.Exists())

	ti, err := tq.Send(TaskOptions{
		Id:   1,
		Name: "Hello!",
	})

	assert.Nil(t, err)

	assert.Equal(t, ti.root.String(), ti.String())

	assert.True(t, ti.IsReady())
	assert.False(t, ti.IsLocked())
	assert.False(t, ti.HasError())
	assert.True(t, ti.TaskFile().Exists())
	assert.False(t, ti.LockFile().Exists())
	assert.False(t, ti.ErrorFile().Exists())

	data, err := ti.TaskFile().Read()
	assert.Nil(t, err)
	assert.Equal(t, `{"id":1,"name":"Hello!"}`, string(data))

}

//nolint:ireturn
func InterfaceToType[T any](inter any, inst any) T {
	v := reflect.ValueOf(inter)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Convert(reflect.TypeOf(inst)).Interface().(T)
}

func TestTaskQueue_Run(t *testing.T) {

	tq := MakeTaskQueue(false)
	err := tq.Initialize()
	assert.Nil(t, err)

	ti, err := tq.Run(TaskOptions{
		Id:   1,
		Name: "Hello!",
	})

	assert.Nil(t, err)
	task := InterfaceToType[ConcreteTask](tq.Task(), ConcreteTask{})

	assert.True(t, task.Executed)

	assert.Equal(t, ti.root.String(), ti.String())
	assert.False(t, ti.Exists())

}
