package main

import (
	"fmt"
	"github.com/alecthomas/kong"
)

func main() {

	cli := &CLI{}
	ctx := kong.Parse(cli,
		kong.Name("queue-demo"),
		kong.Description("queue demo app."),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Vars{
			"version": "0.0.0",
		})

	err := ctx.Run(&kong.Context{})
	if err != nil {
		fmt.Println(err)
	}

}
