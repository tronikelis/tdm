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
			return walker.RContinue(nil)
		}

		syncedSuffix := strings.Replace(path, syncedDir, "", 1)
		fileToCreate := filepath.Join(homeDir, syncedSuffix)

		if filepath.Base(path) == ".git.zip" {
			gitPath, _ := strings.CutSuffix(fileToCreate, ".zip")

			log.Println("unzipping", gitPath)

			if err := files.UnzipToDir(gitPath, path); err != nil {
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

		if filepath.Base(path) == ".git" {
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

		if filepath.Base(syncFrom) == ".git.zip" {
			syncFrom, _ = strings.CutSuffix(syncFrom, ".zip")
		}

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

		if filepath.Base(syncFrom) == ".git" {
			log.Println("zipping", syncFrom)

			if err := files.ZipDirTo(syncFrom, path); err != nil {
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
