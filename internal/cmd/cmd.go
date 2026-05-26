package cmd

import (
	"fmt"
	"os"
)

type Commands struct {
	CName string
	Cmds  []Command
}

func (c *Commands) Name() string {
	return c.CName
}

func (c *Commands) Run(args []string) error {
	if len(args) > 0 {
		for _, cm := range c.Cmds {
			if args[0] == cm.Name() {
				return cm.Run(args[1:])
			}
		}
		fmt.Fprintf(os.Stderr, "command %q not found\n", c.Name())
	}
	for _, cm := range c.Cmds {
		fmt.Fprintf(os.Stderr, "%s\n", cm.Name())
	}
	return nil
}

type Command interface {
	Name() string
	Run(args []string) error
}

type Cmd struct {
	CName string
	CRun  func([]string) error
}

func (c *Cmd) Name() string {
	return c.CName
}

func (c *Cmd) Run(args []string) error {
	return c.CRun(args)
}
