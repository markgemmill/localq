package main

import (
	"fmt"
	"github.com/markgemmill/localq/queue"
	"os"
	"time"
)

type DemoOptions struct {
	QueueDir       string
	MaxExecSeconds int64
	Throttle       float64
	RandErrPer     float64
}

func InitializeTasks(options DemoOptions) (*queue.Queue, error) {
	tasks := queue.GetQueue(options.QueueDir)
	err := tasks.Register(&PrintTask{
		MaxExecutionSeconds: options.MaxExecSeconds,
	})
	if err != nil {
		return nil, err
	}
	return tasks, err
}

func SendCommand(options DemoOptions) {
	fmt.Println("localq-demo send...")
	tasks, err := InitializeTasks(options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	count := 0
	for {
		err := tasks.Enqueue(new(PrintTask)).Send(PrintTaskOptions{Name: fmt.Sprintf("Hello foo #%d", count)})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		count += 1
		time.Sleep(time.Second * time.Duration(options.Throttle))
	}
}

func RunCommand(options DemoOptions) {
	fmt.Println("localq-demo run...")
	tasks, err := InitializeTasks(options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	count := 0
	for {
		count += 1
		fmt.Println("Scan Tasks...")
		tasks.RunAllTasks(count)
		time.Sleep(time.Second * time.Duration(options.Throttle))
	}
}

func FindCommand(options DemoOptions) {
	tasks, err := InitializeTasks(options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	//finder := queue.OrphanedTaskFinder{}
	//err = tasks.FindTasks(finder)
	//for _, task := tasks.FindTasks(finder)
}
