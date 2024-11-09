package terminal

import (
	"github.com/ksrnnb/go-debugger/debugger"
)

type cmdfunc func(dbg *debugger.Debugger, args string) error

type command struct {
	aliases []string
	cmdFn   cmdfunc
}

type Commands struct {
	cmds []command
}

func NewCommands() *Commands {
	return &Commands{
		cmds: []command{
			{
				aliases: []string{"continue", "c"},
				cmdFn:   cont,
			},
			{
				aliases: []string{"quit", "q"},
				cmdFn:   quit,
			},
		},
	}
}

func cont(dbg *debugger.Debugger, args string) error {
	return dbg.Continue()
}

func quit(dbg *debugger.Debugger, args string) error {
	return dbg.Quit()
}
