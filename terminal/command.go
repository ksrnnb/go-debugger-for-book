package terminal

import (
	"errors"
	"strconv"

	"github.com/ksrnnb/go-debugger/debugger"
)

type cmdfunc func(dbg *debugger.Debugger, args []string) error

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
			{
				aliases: []string{"break", "b"},
				cmdFn:   breakpoint,
			},
		},
	}
}

func cont(dbg *debugger.Debugger, args []string) error {
	return dbg.Continue()
}

func quit(dbg *debugger.Debugger, args []string) error {
	return dbg.Quit()
}

func breakpoint(dbg *debugger.Debugger, args []string) error {
	if len(args) == 0 {
		return errors.New("length of args must be greater than 0")
	}

	addr, err := strconv.ParseUint(args[0], 16, 64)
	if err != nil {
		return errors.New("breakpoint address must be parsed as uint64")
	}

	return dbg.SetBreakpoint(addr)
}
