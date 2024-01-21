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

// home <- remote
func SyncFromRemote(syncedDir, homeDir string) error {
	return walker.RecursiveWalk(syncedDir, func(path string, info fs.FileInfo) (bool, error) {
		if info.IsDir() {
			return false, nil
		}

		syncedSuffix := strings.Replace(path, syncedDir, "", 1)
		fileToCreate := filepath.Join(homeDir, syncedSuffix)

		log.Println("syncing", fileToCreate)

		if err := files.MkDirCopyFile(path, fileToCreate); err != nil {
			return true, err
		}

		return false, nil
	})
}

// dir -> remote
func AddToRemote(localDir, syncedDir, homeDir string) error {
	return walker.RecursiveWalk(localDir, func(path string, info fs.FileInfo) (bool, error) {
		localSuffix := strings.Replace(path, homeDir, "", 1)
		fileToCreate := filepath.Join(syncedDir, localSuffix)

		if strings.HasSuffix(path, ".git") {
			log.Println("zipping", path)

			zipFileDir, err := files.ZipToDir(path)
			if err != nil {
				return true, nil
			}

			defer os.Remove(zipFileDir)

			if err := files.MkDirCopyFile(zipFileDir, fileToCreate+".zip"); err != nil {
				return true, nil
			}

			// skip this .git directory
			return true, nil
		}

		if info.IsDir() {
			return false, nil
		}

		log.Println("syncing", fileToCreate)

		if err := files.MkDirCopyFile(path, fileToCreate); err != nil {
			return true, err
		}

		return false, nil
	})
}

// home -> remote
func SyncToRemote(syncedDir, homeDir string) error {
	return walker.RecursiveWalk(syncedDir, func(path string, info fs.FileInfo) (bool, error) {
		if info.IsDir() {
			return false, nil
		}

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
