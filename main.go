package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ksrnnb/go-debugger/debugger"
	"github.com/ksrnnb/go-debugger/terminal"
)

var debuggeePath string

func init() {
	flag.StringVar(&debuggeePath, "path", "", "path of debuggee program")
}

func main() {
	flag.Parse()

	if debuggeePath == "" {
		log.Fatalf("path of debuggee program must be given")
	}

	absDebuggeePath, cleanup, err := buildDebuggeeProgram(debuggeePath)
	if err != nil {
		log.Fatalf("failed to build debuggee program: %s", err)
	}
	defer cleanup()

	d, err := debugger.NewDebugger(&debugger.Config{
		DebuggeePath: absDebuggeePath,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize debugger: %s", err)
		return
	}

	cmds := terminal.NewCommands()
	term := terminal.NewTerminal(d, cmds)

	if err := term.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run terminal: %s", err)
		return
	}

	fmt.Printf("go-debugger gracefully shut down\n")
}
