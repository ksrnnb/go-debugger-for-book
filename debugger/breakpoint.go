package debugger

import (
	"encoding/binary"

	"golang.org/x/sys/unix"
)

const Int3Instruction = 0xcc

type Breakpoint struct {
	pid                 int
	addr                uintptr
	originalInstruction []byte
	isEnabled           bool
}

func NewBreakpoint(pid int, addr uintptr) (*Breakpoint, error) {
	bp := &Breakpoint{pid: pid, addr: addr}
	if err := bp.Enable(); err != nil {
		return nil, err
	}

	return bp, nil
}

// Enable reads the instruction at the address of the breakpoint and rewrites it to an INT3 instruction.
func (bp *Breakpoint) Enable() error {
	_, err := unix.PtracePeekData(bp.pid, bp.addr, bp.originalInstruction)
	if err != nil {
		return err
	}

	data := binary.LittleEndian.Uint64(bp.originalInstruction)
	// data & ^0xff => data & 11111111 11111111 11111111 00000000
	newData := (data & ^uint64(0xff)) | Int3Instruction
	newInstruction := make([]byte, 8)
	binary.LittleEndian.PutUint64(newInstruction, newData)

	_, err = unix.PtracePokeData(bp.pid, bp.addr, newInstruction)
	if err != nil {
		return err
	}

	bp.isEnabled = true
	return nil
}

// Disable updates the instruction at the address of the breakpoint to the original instruction
// before overwriting it with the INT3 instruction
func (bp *Breakpoint) Disable() error {
	_, err := unix.PtracePokeData(bp.pid, bp.addr, bp.originalInstruction)
	if err != nil {
		return err
	}

	bp.isEnabled = false
	return nil
}

func (bp *Breakpoint) IsEnabled() bool {
	return bp.isEnabled
}
