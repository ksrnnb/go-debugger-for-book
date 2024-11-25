package debugger

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
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
	locator     Locator
}

func NewDebugger(config *Config, locator Locator) (*Debugger, error) {
	d := &Debugger{
		config:      config,
		breakpoints: make(map[uint64]*Breakpoint),
		locator:     locator,
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
	if err := d.stepOverBreakpointIfNeeded(); err != nil {
		return fmt.Errorf("failed to step over breakpoint: %w", err)
	}

	if err := syscall.PtraceCont(d.pid, 0); err != nil {
		return fmt.Errorf("faield to execute ptrace cont: %w", err)
	}

	ws, err := d.wait()
	if err != nil {
		return err
	}

	if ws.Stopped() {
		switch ws.StopSignal() {
		case syscall.SIGTRAP:
			if err := d.onBreakpointHit(); err != nil {
				return err
			}
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

func (d *Debugger) DumpRegisters() error {
	return dumpRegisters(d.pid)
}

func (d *Debugger) wait() (syscall.WaitStatus, error) {
	var ws syscall.WaitStatus
	_, err := syscall.Wait4(d.pid, &ws, syscall.WALL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to wait pid %d", d.pid)
	}

	// ws.Exited() will be true when child process is finished.
	if ws.Exited() {
		return 0, ErrDebuggeeFinished
	}

	return ws, nil
}

func (d *Debugger) cleanup() error {
	if err := syscall.Kill(-d.pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to kill child process %d: %s", d.pid, err)
	}
	return nil
}

func (d *Debugger) getPC() (uint64, error) {
	return readRegister(d.pid, Rip)
}

func (d *Debugger) setPC(pc uint64) error {
	return writeRegister(d.pid, Rip, pc)
}

func (d *Debugger) onBreakpointHit() error {
	pc, err := d.getPC()
	if err != nil {
		return err
	}

	// PC is incremented by 1 when the INT3 instruction is executed,
	// so PC is restored when the breakpoint is hit.
	previousPC := pc - 1
	if err := d.setPC(previousPC); err != nil {
		return err
	}

	fmt.Printf("hit breakpoint at 0x%x\n", previousPC)

	if err := d.printSourceCode(); err != nil {
		return err
	}

	return nil
}

func (d *Debugger) stepOverBreakpointIfNeeded() error {
	pc, err := d.getPC()
	if err != nil {
		return err
	}

	bp, ok := d.breakpoints[pc]
	if !ok {
		return nil
	}

	if !bp.IsEnabled() {
		return nil
	}

	if err := bp.Disable(); err != nil {
		return err
	}

	if err := syscall.PtraceSingleStep(d.pid); err != nil {
		return err
	}

	if _, err := d.wait(); err != nil {
		return err
	}

	if err := bp.Enable(); err != nil {
		return err
	}

	return nil
}

func (d *Debugger) printSourceCode() error {
	pc, err := d.getPC()
	if err != nil {
		return err
	}

	filename, line := d.locator.PCToFileLine(pc)
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	printSourceCode(f, line)

	return nil

}
