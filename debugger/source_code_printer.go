package debugger

import (
	"bufio"
	"fmt"
	"io"
)

// how many lines from given line number
const lineRange = 5

// printSourceCode prints out source code passed as a reader, clealy emphasize the currrent line.
func printSourceCode(reader io.Reader, currentLine int) {
	startLine := 1
	if currentLine > lineRange {
		startLine = currentLine - lineRange
	}
	endLine := currentLine + lineRange
	scanLine := 1

	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if scanLine < startLine {
			scanLine++
			continue
		}
		if scanLine > endLine {
			break
		}

		text := scanner.Text()
		if scanLine == currentLine {
			text = fmt.Sprintf("> %d %s", scanLine, text)
		} else {
			text = fmt.Sprintf("  %d %s", scanLine, text)
		}
		lines = append(lines, text)
		scanLine++
	}

	for _, text := range lines {
		fmt.Printf("%s\n", text)
	}
}
