package main

import (
	"fmt"
	"github.com/markgemmill/localq/queue"
	"path"
	"time"
)

type PrintTask struct {
	queue *queue.Queue
}

type PrintTaskOptions struct {
	Name string `json:"name"`
}

func (t *PrintTask) Assert(opt any) error {
	options := opt.(PrintTaskOptions)
	if options.Name == "" {
		return fmt.Errorf("MyTask option.Name is empty")
	}
	return nil
}

func (t *PrintTask) Execute(filePath string, runId int) error {
	name := path.Base(filePath)
	opts, err := queue.ReadTaskFile[PrintTask](filePath)
	if err != nil {
		return err
	}

	time.Sleep(time.Second * 3)
	fmt.Printf("TASK EXE %d: %s: %s\n", runId, name, opts.Name)
	return nil
}
