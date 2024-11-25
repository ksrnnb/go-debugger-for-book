package debugger

import (
	"debug/dwarf"
	"debug/elf"
	"debug/gosym"
	"errors"
	"fmt"
)

type Locator interface {
	PCToFileLine(pc uint64) (filename string, line int)
	FuncToAddr(funcSymbol string) (uint64, error)
	FileLineToAddr(filename string, line int) (uint64, error)
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

func (l *SourceCodeLocator) FuncToAddr(funcSymbol string) (uint64, error) {
	fn := l.st.LookupFunc(funcSymbol)
	if fn == nil {
		return 0, fmt.Errorf("failed to find function: %s", funcSymbol)
	}

	peAddr, err := l.getPrologueEndAddress(fn)
	if err != nil {
		return 0, err
	}

	return peAddr, nil
}

func (l *SourceCodeLocator) FileLineToAddr(filename string, line int) (uint64, error) {
	addr, fn, err := l.st.LineToPC(filename, line)
	if err != nil {
		return 0, fmt.Errorf("failed to get addr by filename %s and line %d: %s", filename, line, err)
	}

	if addr == fn.Entry {
		return l.getPrologueEndAddress(fn)
	}

	return addr, nil
}

func (l *SourceCodeLocator) getPrologueEndAddress(fn *gosym.Func) (uint64, error) {
	reader := l.dwf.Reader()
	for {
		entry, err := reader.Next()
		if err != nil {
			break
		}

		if entry.Tag != dwarf.TagCompileUnit {
			continue
		}

		lineReader, err := l.dwf.LineReader(entry)
		if err != nil {
			return 0, err
		}

		var lineEntry dwarf.LineEntry
		for lineReader.Next(&lineEntry) == nil {
			if lineEntry.Address == fn.Entry {
				for lineReader.Next(&lineEntry) == nil {
					if lineEntry.PrologueEnd {
						return lineEntry.Address, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("faield to get prologue end address for function %s", fn.Name)
}
