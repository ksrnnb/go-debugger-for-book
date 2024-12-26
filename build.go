package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// buildDebuggeeProgram build go program to debug and returns cleanup function.
func buildDebuggeeProgram(path string) (absPath string, cleanup func() error, err error) {
	// Name the file to build in the format __debug__1730159170
	name := fmt.Sprintf("__debug__%d", time.Now().Unix())
	cmd := exec.Command("go", "build", "-o", name, "-gcflags", "all=-N -l", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", nil, fmt.Errorf("failed to build go program: %w", err)
	}

	// Get the absolute path of the built file
	absPath, err = filepath.Abs(name)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get absolute path of debuggee program: %w", err)
	}

	return absPath, func() error {
		if err := os.Remove(absPath); err != nil {
			return fmt.Errorf("failed to remove debuggee program: %w", err)
		}

		return nil
	}, nil
}
