package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
)

func executeDebuggeeProcess(debuggeePath string) (pid int, err error) {
	cmd := exec.Command(debuggeePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// trace debuggee program
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Ptrace: true,
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("failed to start command: %s", err)
	}

	return cmd.Process.Pid, nil
}
