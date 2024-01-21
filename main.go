package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tronikelis/tdm/walker"
)

const SYNCED_PATH string = "synced"

func removeLastOccurrence(target string, occurrence string) string {
	index := strings.LastIndex(target, occurrence)

	if index == -1 {
		return target
	}

	return target[:index] + target[index+len(occurrence):]
}

func syncFromRemote(syncedDir string, homeDir string) error {
	return walker.AppendDir(
		homeDir,
		syncedDir,
		func(base *string, with *string) {
			*with = strings.Replace(*with, syncedDir, "", 1)
		},
		func(fileName string, reader *bufio.Reader, writer *bufio.Writer) error {
			log.Println("syncing", fileName)

			if _, err := writer.ReadFrom(reader); err != nil {
				return err
			}

			return nil
		},
	)

}

func addToRemote(localDir string, syncedDir string, homeDir string) error {
	return walker.AppendDir(
		syncedDir,
		localDir,
		func(base *string, with *string) {
			*with = removeLastOccurrence(*with, homeDir)
		},
		func(fileName string, reader *bufio.Reader, writer *bufio.Writer) error {
			log.Println("adding", fileName)

			if _, err := writer.ReadFrom(reader); err != nil {
				return err
			}

			return nil
		},
	)

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
