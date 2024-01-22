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

func removeLast(str string, target string) string {
	index := strings.LastIndex(str, target)
	if index == -1 {
		return str
	}

	return str[:index] + str[index+len(target):]
}

// home <- remote
func SyncFromRemote(syncedDir, homeDir string) error {
	return walker.RecursiveWalk(syncedDir, func(path string, info fs.FileInfo) (bool, error) {
		if info.IsDir() {
			return walker.RContinue(nil)
		}

		syncedSuffix := strings.Replace(path, syncedDir, "", 1)
		fileToCreate := filepath.Join(homeDir, syncedSuffix)

		if strings.HasSuffix(path, ".git.zip") {
			log.Println("unzipping", path)

			newDir := removeLast(fileToCreate, ".zip")

			if err := files.UnzipToDir(newDir, path); err != nil {
				return walker.RSkip(err)
			}

			return walker.RSkip(nil)
		}

		log.Println("syncing", fileToCreate)

		if err := files.MkDirCopyFile(path, fileToCreate); err != nil {
			return walker.RSkip(err)
		}

		return walker.RContinue(nil)
	})
}

// dir -> remote
func AddToRemote(localDir, syncedDir, homeDir string) error {
	return walker.RecursiveWalk(localDir, func(path string, info fs.FileInfo) (bool, error) {
		localSuffix := strings.Replace(path, homeDir, "", 1)
		fileToCreate := filepath.Join(syncedDir, localSuffix)

		if strings.HasSuffix(path, ".git") {
			log.Println("zipping", path)

			if err := files.ZipDirTo(path, fileToCreate+".zip"); err != nil {
				return walker.RSkip(err)
			}

			// skip this .git directory
			return walker.RSkip(nil)
		}

		if info.IsDir() {
			return walker.RContinue(nil)
		}

		log.Println("syncing", fileToCreate)

		if err := files.MkDirCopyFile(path, fileToCreate); err != nil {
			return walker.RSkip(err)
		}

		return walker.RContinue(nil)
	})
}

// home -> remote
func SyncToRemote(syncedDir, homeDir string) error {
	return walker.RecursiveWalk(syncedDir, func(path string, info fs.FileInfo) (bool, error) {
		if info.IsDir() {
			return walker.RContinue(nil)
		}

		syncedSuffix := strings.Replace(path, syncedDir, "", 1)
		syncFrom := filepath.Join(homeDir, syncedSuffix)

		log.Println("syncing", path)

		_, err := os.Stat(syncFrom)
		if err != nil && os.IsNotExist(err) {
			// if upstream has no file, delete it from tracked as well
			log.Println("removing", path)

			if err := os.Remove(path); err != nil {
				return walker.RSkip(err)
			}

			return walker.RContinue(nil)
		}

		if strings.HasSuffix(syncedSuffix, ".git.zip") {
			if err := files.ZipDirTo(removeLast(syncFrom, ".zip"), path); err != nil {
				return walker.RSkip(err)
			}

			return walker.RSkip(nil)
		}

		if err := files.MkDirCopyFile(syncFrom, path); err != nil {
			return walker.RSkip(err)
		}

		return walker.RContinue(nil)
	})
}
