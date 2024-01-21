package walker

import (
	"io/fs"
	"path/filepath"
)

func WalkFiles(
	targetDir string,
	callback func(path string) error,
) error {
	return filepath.Walk(targetDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		return callback(path)
	})
}
