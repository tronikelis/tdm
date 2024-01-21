package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tronikelis/tdm/files"
	"github.com/Tronikelis/tdm/walker"
)

const SYNCED_PATH string = "synced"

func syncFromRemote(syncedDir string, homeDir string) error {
	return walker.WalkFiles(syncedDir, func(path string) error {
		syncedSuffix := strings.Replace(path, syncedDir, "", 1)
		fileToCreate := filepath.Join(homeDir, syncedSuffix)

		log.Println("syncing", fileToCreate)

		return files.MkDirCopyFile(path, fileToCreate)
	})
}

func addToRemote(localDir string, syncedDir string, homeDir string) error {
	return walker.WalkFiles(localDir, func(path string) error {
		localSuffix := strings.Replace(path, homeDir, "", 1)
		fileToCreate := filepath.Join(syncedDir, localSuffix)

		log.Println("syncing", fileToCreate)

		return files.MkDirCopyFile(path, fileToCreate)
	})
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	syncedDir := filepath.Join(homeDir, ".tdm", SYNCED_PATH)

	if err := os.MkdirAll(syncedDir, os.ModePerm); err != nil {
		log.Fatalln(err)
	}

	args := os.Args

	if len(args) == 1 {
		if err := syncFromRemote(syncedDir, homeDir); err != nil {
			log.Fatalln(err)
		}

		return
	}

	localDir := args[1]
	if !filepath.IsAbs(localDir) {
		localDir = filepath.Join(cwd, args[1])
	}

	if !strings.HasPrefix(localDir, homeDir) {
		log.Fatalln("adding non /home items is not supported")
	}

	if err := addToRemote(localDir, syncedDir, homeDir); err != nil {
		log.Fatalln(err)
	}
}
