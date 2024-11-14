package debugger

import (
	"fmt"
	"reflect"
	"syscall"
)

type Register string

const (
	Rbp Register = "Rbp" // base pointer
	Rip Register = "Rip" // instruction pointer (= program counter)
	Rsp Register = "Rsp" // stack pointer
)

type RegisterClient struct {
	pid int
}

func NewRegisterClient(pid int) RegisterClient {
	return RegisterClient{pid: pid}
}

func (c RegisterClient) GetRegisterValue(register Register) (uint64, error) {
	regs := &syscall.PtraceRegs{}
	if err := syscall.PtraceGetRegs(c.pid, regs); err != nil {
		return 0, fmt.Errorf("failed to get register values for %s and pid %d: %s", register, c.pid, err)
	}

	v := reflect.ValueOf(regs).Elem()
	field := v.FieldByName(string(register))
	if !field.IsValid() {
		return 0, fmt.Errorf("no '%s' field in syscall.PtraceRegs", register)
	}
	if field.Kind() != reflect.Uint64 {
		return 0, fmt.Errorf("field %s is not of type uint64", register)
	}

	return field.Uint(), nil
}

func (c RegisterClient) SetRegisterValue(register Register, value uint64) error {
	regs := &syscall.PtraceRegs{}
	if err := syscall.PtraceGetRegs(c.pid, regs); err != nil {
		return err
	}

	v := reflect.ValueOf(regs).Elem()
	field := v.FieldByName(string(register))
	if !field.IsValid() {
		return fmt.Errorf("no '%s' field in syscall.PtraceRegs", register)
	}
	if field.Kind() != reflect.Uint64 {
		return fmt.Errorf("field %s is not of type uint64", register)
	}
	if !field.CanSet() {
		return fmt.Errorf("field %s cannot set", register)
	}
	field.SetUint(value)

	return syscall.PtraceSetRegs(c.pid, regs)
}

func (c RegisterClient) DumpRegisters() error {
	regs := &syscall.PtraceRegs{}
	if err := syscall.PtraceGetRegs(c.pid, regs); err != nil {
		return fmt.Errorf("failed to get regs for pid %d: %s", c.pid, err)
	}

	v := reflect.ValueOf(regs).Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fmt.Printf("%s: 0x%x\n", v.Type().Field(i).Name, field.Uint())
	}

	return nil
}
