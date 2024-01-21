package files

import (
	"bufio"
	"os"
	"path/filepath"
)

// copies "from" into "to" and creates necessary directories
func MkDirCopyFile(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), os.ModePerm); err != nil {
		return err
	}

	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	toFile, err := os.Create(to)
	if err != nil {
		return err
	}
	defer toFile.Close()

	reader := bufio.NewReader(fromFile)
	writer := bufio.NewWriter(toFile)

	defer writer.Flush()

	if _, err := writer.ReadFrom(reader); err != nil {
		return err
	}

	return nil
}
