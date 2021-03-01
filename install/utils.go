package install

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/posener/script"
)

func lineInFile(path string, line string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// Splits on newlines by default.
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), line) {
			return true
		}
	}
	return false
}

func createFile(path string, content string) error {
	return script.Echo(content).ToFile(path)
}

func appendFile(path string, content string) error {
	return script.Echo(content).AppendFile(path)
}

func removeFromFile(path string, line string) error {
	backupPath := path + ".bck"
	err := script.Cat(path).ToFile(backupPath)
	if err != nil {
		return fmt.Errorf("creating backup file: %s", err)
	}

	tmp, err := script.Cat(path).Modify(NewSimpleSearcher(line)).ToTempFile()
	if err != nil {
		return fmt.Errorf("failed remove: %s", err)
	}
	defer os.Remove(tmp)
	err = script.Cat(tmp).ToFile(path)
	if err != nil {
		restoreErr := script.Cat(backupPath).ToFile(path)
		if restoreErr != nil {
			return fmt.Errorf("failed write: %s, and failed restore: %s", err, restoreErr)
		}
	}
	return nil
}

// SimpleSearcher is a modifier that filters only line that equal line we are searching for. If Invert was set only line that did
type SimpleSearcher struct {
	line []byte
}

func (g SimpleSearcher) Modify(line []byte) (modifed []byte, err error) {
	if line == nil {
		return nil, nil
	}

	if !bytes.Equal(line, g.line) {
		return append(line, '\n'), nil
	}
	return nil, nil
}

func (g SimpleSearcher) Name() string {
	return fmt.Sprintf("Simple searcher. Line: %s", g.line)
}

func NewSimpleSearcher(line string) script.Modifier {
	return SimpleSearcher{line: []byte(line)}
}
