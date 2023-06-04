package main

import "github.com/alecthomas/kong"

type Globals struct {
	QueueDir string  `name:"queue-dir" required:""`
	Throttle float64 `name:"throttle" default:"1.0" help:"Number of seconds between processing tasks."`
}

type SendCmd struct {
	Globals
}

func (cmd *SendCmd) Run(ctx *kong.Context) error {
	SendCommand(DemoOptions{
		QueueDir: cmd.QueueDir,
		Throttle: cmd.Throttle,
	})
	return nil
}

type RunCmd struct {
	Globals
	MaxExecSec   int64   `name:"max-exec-sec" default:"5" help:"Max time task processes for."`
	RandomErrPer float64 `name:"rand-err-per" default:"0.0" help:"Percetage of task to error when running."`
}

func (cmd *RunCmd) Run(ctx *kong.Context) error {
	RunCommand(DemoOptions{
		QueueDir:       cmd.QueueDir,
		Throttle:       cmd.Throttle,
		MaxExecSeconds: cmd.MaxExecSec,
	})
	return nil
}

type FindCmd struct {
	Globals
	Status string `name:"status" enum:"all,error,orphaned,open" help:"What type of task to find."`
}

func (cmd *FindCmd) Run(ctx *kong.Context) error {
	return nil
}

type CLI struct {
	Send SendCmd `cmd:""`
	Run  RunCmd  `cmd:""`
	Find FindCmd `cmd:""`
}
