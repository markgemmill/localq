package main

import (
	"fmt"
	"os"
	"time"
)
import "github.com/markgemmill/localq/queue"

type DemoOptions struct {
	QueueDir string
}

func InitializeTasks(options DemoOptions) *queue.Queue {
	tasks := queue.GetQueue(options.QueueDir)
	tasks.Register(&PrintTask{})
	return tasks
}

func SendCommand(options DemoOptions) {
	fmt.Println("Running the send command...")
	tasks := InitializeTasks(options)
	count := 0
	for {
		err := tasks.Enqueue(new(PrintTask)).Send(PrintTaskOptions{Name: fmt.Sprintf("Hello foo #%d", count)})
		if err != nil {
			fmt.Println(err)
			os.Exit(5)
		}
		count += 1
		time.Sleep(time.Second * 1)
	}
}

func RunCommand(options DemoOptions) {
	fmt.Println("Running the run command...")
	tasks := InitializeTasks(options)
	count := 0
	for {
		fmt.Println("Scan Tasks...")
		count += 1
		tasks.RunAllTasks(count)
		time.Sleep(time.Second * 10)
	}
}
