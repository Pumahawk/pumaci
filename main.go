package main

import (
	"fmt"
	"os"

	"github.com/Pumahawk/pumaci/internal/cmd"
)

var AppCmd = cmd.Commands{
	CName: "app",
	Cmds: []cmd.Command{
		HelloWorldCmd,
	},
}

var HelloWorldCmd = &cmd.Cmd{
	CName: "hello",
	CRun: func(s []string) error {
		fmt.Printf("hello world\n")
		return nil
	},
}

func main() {
	args := os.Args[1:]
	AppCmd.Run(args)
}
