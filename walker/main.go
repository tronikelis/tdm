package walker

import (
	"bufio"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func AppendDir(
	base string,
	with string,
	pathChanger func(base *string, with *string),
	copier func(fileName string, reader *bufio.Reader, writer *bufio.Writer) error,
) error {
	return filepath.Walk(with, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		targetDir := filepath.Join(base, path)

		pathChanger(&path, &targetDir)

		if err := os.MkdirAll(filepath.Dir(targetDir), os.ModePerm); err != nil {
			return err
		}

		fromFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fromFile.Close()

		log.Println("opening", path)

		toFile, err := os.Create(targetDir)
		if err != nil {
			return err
		}
		defer toFile.Close()

		log.Println("creating", targetDir)

		reader := bufio.NewReader(fromFile)
		writer := bufio.NewWriter(toFile)

		defer writer.Flush()

		return copier(info.Name(), reader, writer)
	})
}
