package terminal

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/ksrnnb/go-debugger/debugger"
)

const prompt = "go-debugger>"

type Terminal struct {
	debugger *debugger.Debugger
	cmds     *Commands
}

func NewTerminal(debugger *debugger.Debugger, cmds *Commands) *Terminal {
	return &Terminal{debugger, cmds}
}

func (t *Terminal) Run() error {
	sc := bufio.NewScanner(os.Stdin)

	fmt.Printf("%s ", prompt)

	for sc.Scan() {
		input := sc.Text()

		cmdFn, err := t.Find(input)
		if err != nil {
			fmt.Printf("failed to parse command: %s\n", err)
			fmt.Printf("%s ", prompt)
			continue
		}

		// input: "<command> <arg1> <arg2>"
		// -> args: [<arg1>, <arg2>]
		s := strings.SplitN(input, " ", 2)
		var args string
		if len(s) == 2 {
			args = s[1]
		}

		if err := cmdFn(t.debugger, args); err != nil {
			if errors.Is(err, debugger.ErrDebuggeeFinished) {
				break
			}
			return err
		}
		fmt.Printf("\n%s ", prompt)
	}

	return nil
}

func (t *Terminal) Find(commandWithArgs string) (cmdfunc, error) {
	s := strings.SplitN(commandWithArgs, " ", 2)
	command := s[0]
	for _, cmd := range t.cmds.cmds {
		if slices.Contains(cmd.aliases, command) {
			return cmd.cmdFn, nil
		}
	}

	return nil, fmt.Errorf("command %s is not found", command)
}
