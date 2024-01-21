package files

import (
	"archive/zip"
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tronikelis/tdm/walker"
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

func ZipToDir(dir string) (string, error) {
	file, err := os.CreateTemp("", "tdm_zip")
	if err != nil {
		return "", err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	if err := walker.RecursiveWalk(dir, func(path string, info fs.FileInfo) (bool, error) {
		if info.IsDir() {
			return false, nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return true, err
		}

		header.Name = strings.Replace(path, dir, "", 1)

		openedZip, err := zipWriter.CreateHeader(header)
		if err != nil {
			return true, err
		}

		openedFile, err := os.Open(path)
		if err != nil {
			return true, err
		}
		defer openedFile.Close()

		reader := bufio.NewReader(openedFile)
		writer := bufio.NewWriter(openedZip)

		defer writer.Flush()

		if _, err := writer.ReadFrom(reader); err != nil {
			return true, err
		}

		return false, nil
	}); err != nil {
		return "", err
	}

	return file.Name(), nil
}
