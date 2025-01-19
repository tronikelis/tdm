package runner

import (
	"archive/zip"
	"io"
	"os"
	"path"
)

type fileLike interface {
	io.Reader
	Stat() (os.FileInfo, error)
}

type zipFileLike struct {
	io.Reader
	*zip.File
}

func newZipFileLike(r io.Reader, f *zip.File) zipFileLike {
	return zipFileLike{Reader: r, File: f}
}

func (z zipFileLike) Stat() (os.FileInfo, error) {
	return z.FileInfo(), nil
}

func writeFileLike(f fileLike, to string) error {
	if err := os.MkdirAll(path.Dir(to), 0o770); err != nil {
		return err
	}

	fStat, err := f.Stat()
	if err != nil {
		return err
	}

	t, err := os.OpenFile(to, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, fStat.Mode())
	if err != nil {
		return err
	}
	defer t.Close()

	_, err = io.Copy(t, f)
	if err != nil {
		return err
	}

	return nil
}

// opens file at `from` and calls `writeFileLike`
func write(from string, to string) error {
	f, err := os.Open(from)
	if err != nil {
		return err
	}
	defer f.Close()

	return writeFileLike(f, to)
}
