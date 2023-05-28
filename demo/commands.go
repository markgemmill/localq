package main

import "fmt"

type DemoOptions struct {
	QueueDir string
}

func SendCommand(options DemoOptions) {
	fmt.Println("Running the send command...")
}

func RunCommand(options DemoOptions) {
	fmt.Println("Running the run command...")
}
