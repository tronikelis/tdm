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

func RecursiveWalk(targetDir string, shouldContinue func(path string) bool) error {
	dirs, err := os.ReadDir(targetDir)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		childDir := filepath.Join(targetDir, dir.Name())

		if !shouldContinue(childDir) {
			continue
		}

		if !dir.IsDir() {
			continue
		}

		if err := RecursiveWalk(childDir, shouldContinue); err != nil {
			return err
		}
	}

	return nil
}
