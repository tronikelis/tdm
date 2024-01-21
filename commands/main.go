package commands

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tronikelis/tdm/files"
	"github.com/Tronikelis/tdm/walker"
)

func SyncFromRemote(syncedDir, homeDir string) error {
	return walker.WalkFiles(syncedDir, func(path string) error {
		syncedSuffix := strings.Replace(path, syncedDir, "", 1)
		fileToCreate := filepath.Join(homeDir, syncedSuffix)

		log.Println("syncing", fileToCreate)

		return files.MkDirCopyFile(path, fileToCreate)
	})
}

func AddToRemote(localDir, syncedDir, homeDir string) error {
	return walker.WalkFiles(localDir, func(path string) error {
		localSuffix := strings.Replace(path, homeDir, "", 1)
		fileToCreate := filepath.Join(syncedDir, localSuffix)

		log.Println("syncing", fileToCreate)

		return files.MkDirCopyFile(path, fileToCreate)
	})
}

func SyncToRemote(syncedDir, homeDir string) error {
	return walker.WalkFiles(syncedDir, func(path string) error {
		syncedSuffix := strings.Replace(path, syncedDir, "", 1)
		syncFrom := filepath.Join(homeDir, syncedSuffix)

		log.Println("syncing", path)

		_, err := os.Stat(syncFrom)
		if err != nil && os.IsNotExist(err) {
			// if upstream has no file, delete it from tracked as well
			log.Println("removing", path)

			if err := os.Remove(path); err != nil {
				return err
			}

			return nil
		}

		return files.MkDirCopyFile(syncFrom, path)
	})
}
