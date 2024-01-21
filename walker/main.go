package walker

import (
	"io/fs"
	"os"
	"path/filepath"
)

func RecursiveWalk(targetDir string, shouldSkip func(path string, info fs.FileInfo) (bool, error)) error {
	info, err := os.Stat(targetDir)
	if err != nil {
		return err
	}

	skip, err := shouldSkip(targetDir, info)
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
