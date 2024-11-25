package terminal

import (
	"errors"
	"fmt"
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
				cmdFn:   setBreakpoint,
			},
			{
				aliases: []string{"dump", "d"},
				cmdFn:   dumpRegisters,
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

func setBreakpoint(dbg *debugger.Debugger, args []string) error {
	if len(args) == 0 {
		return errors.New("length of args must be greater than 0")
	}

	addr, err := strconv.ParseUint(args[0], 16, 64)
	if err == nil {
		return dbg.SetBreakpoint(debugger.SetBreakpointArgs{Addr: addr})
	}

	if len(args) == 1 {
		return dbg.SetBreakpoint(debugger.SetBreakpointArgs{FunctionSymbol: args[0]})
	}

	if len(args) == 2 {
		line, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("failed to parse line number: %w", err)
		}

		return dbg.SetBreakpoint(debugger.SetBreakpointArgs{
			Filename: args[0],
			Line:     line,
		})
	}

	return errors.New("length of args must be 1 or 2")
}

func dumpRegisters(dbg *debugger.Debugger, args []string) error {
	return dbg.DumpRegisters()
}
