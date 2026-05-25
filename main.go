package main

import (
	"os"

	"github.com/Pumahawk/pumaci/internal/cmd"
	"github.com/Pumahawk/pumaci/internal/pumacicmd"
)

var AppCmd = cmd.Commands{
	CName: "app",
	Cmds: []cmd.Command{
		pumacicmd.Manifest,
	},
}

func main() {
	args := os.Args[1:]
	AppCmd.Run(args)
}
