package main

import (
	"bufio"
	"log"
	"os"
	"path"

	"github.com/tronikelis/tdm/runner"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatalln("provide command")
	}

	cmd := os.Args[1]
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}
	synced := path.Join(home, ".tdm/synced")

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}

	stdoutBuf := bufio.NewWriter(os.Stdout)
	defer stdoutBuf.Flush()
	runner := runner.NewRunner(synced, home, wd, log.New(stdoutBuf, "", log.Default().Flags()))

	switch cmd {
	case "add":
		target := ""
		if len(os.Args) >= 3 {
			target = os.Args[2]
		}
		if errors := runner.Add(target); len(errors) != 0 {
			log.Fatalln(errors)
		}
	case "sync":
		panic("not implemented")
	}
}
