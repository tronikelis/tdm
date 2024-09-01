package files

import (
	"archive/zip"
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tronikelis/tdm/walker"
)

func createNecessaryDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), os.ModePerm)
}

func pipe(reader io.Reader, writer io.Writer) error {
	bufReader := bufio.NewReader(reader)
	bufWriter := bufio.NewWriter(writer)

	defer bufWriter.Flush()

	if _, err := bufWriter.ReadFrom(bufReader); err != nil {
		return err
	}

	return nil
}

// copies "from" into "to" and creates necessary directories
func MkDirCopyFile(from, to string) error {
	if err := createNecessaryDir(to); err != nil {
		return err
	}

	fromFile, err := os.Open(from)
	if err != nil {
		return err
	}
	defer fromFile.Close()

	fromFileStat, err := os.Stat(from)
	if err != nil {
		return err
	}

	toFile, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_RDWR, fromFileStat.Mode().Perm())
	if err != nil {
		return err
	}
	defer toFile.Close()

	return pipe(fromFile, toFile)
}

func ZipDirTo(dir, filePath string) error {
	if err := createNecessaryDir(filePath); err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	return walker.RecursiveWalk(dir, func(path string, info fs.FileInfo) (bool, error) {
		if info.IsDir() {
			return walker.RContinue(nil)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return walker.RSkip(err)
		}

		header.Name = strings.Replace(path, dir, "", 1)

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return walker.RSkip(err)
		}

		reader, err := os.Open(path)
		if err != nil {
			return walker.RSkip(err)
		}
		defer reader.Close()

		if err := pipe(reader, writer); err != nil {
			return walker.RSkip(err)
		}

		return walker.RContinue(nil)
	})
}

func UnzipToDir(dir, filePath string) error {
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, zippedFile := range zipReader.File {
		reader, err := zippedFile.Open()
		if err != nil {
			return err
		}
		defer reader.Close()

		fileToCreate := filepath.Join(dir, zippedFile.Name)
		if err := createNecessaryDir(fileToCreate); err != nil {
			return err
		}

		writer, err := os.Create(fileToCreate)
		if err != nil {
			return err
		}
		defer writer.Close()

		if err := pipe(reader, writer); err != nil {
			return err
		}
	}

	return nil
}
