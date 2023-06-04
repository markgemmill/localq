package main

import (
	"fmt"
	"github.com/markgemmill/localq/queue"
	"math/rand"
	"path"
	"time"
)

func CalcExecutionTime(maxSeconds int64) time.Duration {
	max := time.Second * time.Duration(maxSeconds)
	percent := rand.Float64()
	nanoSeconds := float64(max) * percent
	fmt.Printf("execution time: max of %d * %f %% = %f nanoSeconds\n", max, percent, nanoSeconds)
	return time.Duration(nanoSeconds)
}

type PrintTask struct {
	MaxExecutionSeconds int64
	RandomErrors        float64
}

type PrintTaskOptions struct {
	Name string `json:"name"`
}

func (t *PrintTask) Assert(opt any) error {
	options := opt.(PrintTaskOptions)
	if options.Name == "" {
		return fmt.Errorf("PrintTaskOptions.Name is empty")
	}
	return nil
}

func (t *PrintTask) RaiseError() bool {
	return t.RandomErrors > 0.0 && rand.Float64() <= t.RandomErrors
}

func (t *PrintTask) Execute(filePath string, runId int) error {
	name := path.Base(filePath)
	opts, err := queue.ReadTaskFile[PrintTaskOptions](filePath)
	if err != nil {
		return err
	}

	if t.RaiseError() {
		fmt.Printf("TASK ERR %d: %s: %s\n", runId, name, opts.Name)
		return fmt.Errorf("A randomly selected error occurred!")
	}

	time.Sleep(CalcExecutionTime(t.MaxExecutionSeconds))
	fmt.Printf("TASK EXE %d: %s: %s\n", runId, name, opts.Name)
	return nil
}
