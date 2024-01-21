package commands

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tronikelis/tdm/files"
	"github.com/Tronikelis/tdm/walker"
)

func SyncFromRemote(syncedDir, homeDir string) error {
	return walker.WalkFiles(syncedDir, func(path string, info fs.FileInfo) (bool, error) {
		syncedSuffix := strings.Replace(path, syncedDir, "", 1)
		fileToCreate := filepath.Join(homeDir, syncedSuffix)

		log.Println("syncing", fileToCreate)

		if err := files.MkDirCopyFile(path, fileToCreate); err != nil {
			return true, err
		}

		return false, nil
	})
}

func AddToRemote(localDir, syncedDir, homeDir string) error {
	return walker.WalkFiles(localDir, func(path string, info fs.FileInfo) (bool, error) {
		localSuffix := strings.Replace(path, homeDir, "", 1)
		fileToCreate := filepath.Join(syncedDir, localSuffix)

		log.Println("syncing", fileToCreate)

		if err := files.MkDirCopyFile(path, fileToCreate); err != nil {
			return true, err
		}

		return false, nil
	})
}

func SyncToRemote(syncedDir, homeDir string) error {
	return walker.WalkFiles(syncedDir, func(path string, info fs.FileInfo) (bool, error) {
		syncedSuffix := strings.Replace(path, syncedDir, "", 1)
		syncFrom := filepath.Join(homeDir, syncedSuffix)

		log.Println("syncing", path)

		_, err := os.Stat(syncFrom)
		if err != nil && os.IsNotExist(err) {
			// if upstream has no file, delete it from tracked as well
			log.Println("removing", path)

			if err := os.Remove(path); err != nil {
				return true, err
			}

			return false, nil
		}

		if err := files.MkDirCopyFile(syncFrom, path); err != nil {
			return true, err
		}

		return false, nil
	})
}
