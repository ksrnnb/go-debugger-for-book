package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
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

	pid, err := executeDebuggeeProcess(absDebuggeePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to execute debugee program: %s\n", err)
		return
	}

	fmt.Printf("pid of debuggee program is %d\n", pid)

	var ws unix.WaitStatus
	_, err = unix.Wait4(pid, &ws, unix.WALL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to wait pid %d\n", pid)
		return
	}

	if err := syscall.PtraceCont(pid, 0); err != nil {
		fmt.Fprintf(os.Stderr, "faield to execute ptrace cont: %s\n", err)
		return
	}

	_, err = unix.Wait4(pid, &ws, unix.WALL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to wait pid %d\n", pid)
		return
	}
}
