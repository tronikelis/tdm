package walker

import (
	"io/fs"
	"os"
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

func RecursiveWalk(targetDir string, shouldSkip func(path string) (bool, error)) error {
	info, err := os.Stat(targetDir)
	if err != nil {
		return err
	}

	skip, err := shouldSkip(targetDir)
	if err != nil {
		return err
	}

	if !info.IsDir() || skip {
		return nil
	}

	dirs, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		childDir := filepath.Join(targetDir, dir.Name())

		if err := RecursiveWalk(childDir, shouldSkip); err != nil {
			return err
		}
	}

	return nil
}
