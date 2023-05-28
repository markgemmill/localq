package main

import "github.com/alecthomas/kong"

type Globals struct {
	QueueDir string `name:"queue-dir"`
}

type SendCmd struct {
	Globals
	Verbose int `short:"v" type:"counter" help:"Verbosity can have a value of 1-3. Example: --verbose=3 or -vvv."`
}

func (cmd *SendCmd) Run(ctx *kong.Context) error {
	SendCommand(DemoOptions{
		QueueDir: cmd.QueueDir,
	})
	return nil
}

type RunCmd struct {
	Globals
}

func (cmd *RunCmd) Run(ctx *kong.Context) error {
	RunCommand(DemoOptions{
		QueueDir: cmd.QueueDir,
	})
	return nil
}

type CLI struct {
	Send SendCmd `cmd:""`
	Run  RunCmd  `cmd:""`
}
