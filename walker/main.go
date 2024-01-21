package walker

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func WalkFiles(
	targetDir string,
	callback func(path string, relativePath string, info fs.FileInfo) error,
) error {
	return filepath.Walk(targetDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relativePath := strings.Replace(path, targetDir, "", 1)

		return callback(path, relativePath, info)
	})
}

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
