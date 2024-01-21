package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tronikelis/tdm/commands"
)

const SYNCED_PATH string = "synced"

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	syncedDir := filepath.Join(homeDir, ".tdm", SYNCED_PATH)

	if err := os.MkdirAll(syncedDir, os.ModePerm); err != nil {
		log.Fatalln(err)
	}

	args := os.Args[1:]

	if len(args) == 0 {
		log.Fatalln("the available commands are: [add, sync]")
	}

	if len(args) == 1 {
		if args[0] == "add" {
			log.Println("re-adding all synced files")

			if err := commands.SyncToRemote(syncedDir, homeDir); err != nil {
				log.Fatalln(err)
			}

			return
		}

		if args[0] == "sync" {
			log.Println("syncing files from tracked directory")

			if err := commands.SyncFromRemote(syncedDir, homeDir); err != nil {
				log.Fatal(err)
			}

			return
		}

		log.Fatalln("unknown command, valid commands are: [add, sync]")
	}

	if len(args) == 2 {
		if args[0] == "add" {
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalln(err)
			}

			localDir := args[1]
			if !filepath.IsAbs(localDir) {
				localDir = filepath.Join(cwd, args[1])
			}

			if !strings.HasPrefix(localDir, homeDir) {
				log.Fatalln("adding non /home items is not supported")
			}

			if err := commands.AddToRemote(localDir, syncedDir, homeDir); err != nil {
				log.Fatalln(err)
			}

			return
		}

		log.Fatalln("unknown command")
	}
}
