package debugger

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// ErrDebuggeeFinished is used when debuggee processs is finished.
var ErrDebuggeeFinished = errors.New("debuggee process is finished")

type Config struct {
	DebuggeePath string
}

type Debugger struct {
	config      *Config
	pid         int
	breakpoints map[uint64]*Breakpoint
}

func NewDebugger(config *Config) (*Debugger, error) {
	d := &Debugger{
		config:      config,
		breakpoints: make(map[uint64]*Breakpoint),
	}
	if err := d.Launch(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Debugger) Launch() error {
	cmd := exec.Command(d.config.DebuggeePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// trace debuggee program
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Ptrace:  true,
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch debuggee process: %w", err)
	}

	d.pid = cmd.Process.Pid

	if _, err := d.wait(); err != nil {
		return err
	}

	return nil
}

func (d *Debugger) Continue() error {
	if err := syscall.PtraceCont(d.pid, 0); err != nil {
		return fmt.Errorf("faield to execute ptrace cont: %w", err)
	}

	ws, err := d.wait()
	if err != nil {
		return err
	}

	// ws.Exited() will be true when child process is finished.
	if ws.Exited() {
		return ErrDebuggeeFinished
	}

	if ws.Stopped() {
		switch ws.StopSignal() {
		case syscall.SIGTRAP:
			fmt.Println("hit breakpoint!")
		default:
			// ignore SIGURG signal because it is not expected signal
			return d.Continue()
		}
	}

	return nil
}

func (d *Debugger) Quit() error {
	if err := d.cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup: %s", err)
	}

	return ErrDebuggeeFinished
}

func (d *Debugger) SetBreakpoint(addr uint64) error {
	bp, err := NewBreakpoint(d.pid, uintptr(addr))
	if err != nil {
		return err
	}

	d.breakpoints[addr] = bp

	return nil
}

func (d *Debugger) wait() (unix.WaitStatus, error) {
	var ws unix.WaitStatus
	_, err := unix.Wait4(d.pid, &ws, unix.WALL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to wait pid %d", d.pid)
	}

	return ws, nil
}

func (d *Debugger) cleanup() error {
	if err := syscall.Kill(-d.pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to kill child process %d: %s", d.pid, err)
	}
	return nil
}
