package debugger

import (
	"encoding/binary"
	"syscall"
)

const Int3Instruction = 0xcc

type Breakpoint struct {
	pid                 int
	addr                uintptr
	originalInstruction []byte
	isEnabled           bool
}

func NewBreakpoint(pid int, addr uintptr) (*Breakpoint, error) {
	// originalInstruction must be allocated 8 bytes buffer to execute PtracePeekData
	bp := &Breakpoint{pid: pid, addr: addr, originalInstruction: make([]byte, 8)}
	if err := bp.Enable(); err != nil {
		return nil, err
	}

	return bp, nil
}

// Enable reads the instruction at the address of the breakpoint and rewrites it to an INT3 instruction.
func (bp *Breakpoint) Enable() error {
	_, err := syscall.PtracePeekData(bp.pid, bp.addr, bp.originalInstruction)
	if err != nil {
		return err
	}

	data := binary.LittleEndian.Uint64(bp.originalInstruction)
	// data & ^uint64(0xff) => data & 11111111 11111111 11111111 11111111 11111111 11111111 11111111 00000000
	newData := (data & ^uint64(0xff)) | Int3Instruction
	newInstruction := make([]byte, 8)
	binary.LittleEndian.PutUint64(newInstruction, newData)

	_, err = syscall.PtracePokeData(bp.pid, bp.addr, newInstruction)
	if err != nil {
		return err
	}

	bp.isEnabled = true
	return nil
}

// Disable updates the instruction at the address of the breakpoint to the original instruction
// before overwriting it with the INT3 instruction
func (bp *Breakpoint) Disable() error {
	_, err := syscall.PtracePokeData(bp.pid, bp.addr, bp.originalInstruction)
	if err != nil {
		return err
	}

	bp.isEnabled = false
	return nil
}

func (bp *Breakpoint) IsEnabled() bool {
	return bp.isEnabled
}
