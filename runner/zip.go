package runner

import (
	"archive/zip"
	"io/fs"
	"os"
	"path"
	"strings"
)

func zipDir(from fs.FS, to string) error {
	if err := os.MkdirAll(path.Dir(to), 0o770); err != nil {
		return err
	}

	toFile, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o660)
	if err != nil {
		return err
	}
	defer toFile.Close()

	writer := zip.NewWriter(toFile)
	defer writer.Close()

	return writer.AddFS(from)
}

func unzipDir(from string, to string) error {
	reader, err := zip.OpenReader(from)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, v := range reader.File {
		reader, err := v.Open()
		if err != nil {
			return err
		}
		defer reader.Close()

		toPath := path.Join(to, v.Name)

		// empty directory
		if strings.HasSuffix(v.Name, "/") {
			if err := os.MkdirAll(toPath, 0o770); err != nil {
				return err
			}
			continue
		}

		if err := writeFileLike(newZipFileLike(reader, v), toPath); err != nil {
			return err
		}
	}

	return nil
}
