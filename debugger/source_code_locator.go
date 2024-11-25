package debugger

import (
	"debug/dwarf"
	"debug/elf"
	"debug/gosym"
	"errors"
)

type Locator interface {
	PCToFileLine(pc uint64) (filename string, line int)
}

// SourceCodeLocator converts memory address to file name and line number.
type SourceCodeLocator struct {
	st  *gosym.Table
	dwf *dwarf.Data
}

// This implementation is based on the process in pclntab_test.go file.
// https://cs.opensource.google/go/go/+/refs/tags/go1.23.2:src/debug/gosym/pclntab_test.go;l=86
func NewSourceCodeLocator(debuggeePath string) (*SourceCodeLocator, error) {
	f, err := elf.Open(debuggeePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	s := f.Section(".gosymtab")
	if s == nil {
		return nil, errors.New(".gosymtab section is not found")
	}

	symdata, err := s.Data()
	if err != nil {
		return nil, err
	}

	pclndata, err := f.Section(".gopclntab").Data()
	if err != nil {
		return nil, err
	}

	pcln := gosym.NewLineTable(pclndata, f.Section(".text").Addr)

	table, err := gosym.NewTable(symdata, pcln)
	if err != nil {
		return nil, err
	}

	dwf, err := f.DWARF()
	if err != nil {
		return nil, err
	}

	return &SourceCodeLocator{
		st:  table,
		dwf: dwf,
	}, nil

}

func (l *SourceCodeLocator) PCToFileLine(pc uint64) (filename string, line int) {
	fn, ln, _ := l.st.PCToLine(pc)
	return fn, ln
}
