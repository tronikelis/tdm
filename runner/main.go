package runner

import (
	"archive/zip"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"strings"
)

type Runner struct {
	synced string
	home   string
	wd     string
	logger *log.Logger
}

func NewRunner(synced string, home string, wd string, logger *log.Logger) Runner {
	return Runner{
		synced: synced,
		home:   home,
		wd:     wd,
		logger: logger,
	}
}

// takes p, and returns p in synced directory
func (r Runner) pathSynced(p string) (string, error) {
	p = r.normalize(p)

	withoutHome, found := strings.CutPrefix(p, r.home)
	if !found {
		return "", errors.New("can't convert real path to synced path")
	}

	return path.Join(r.synced, withoutHome), nil
}

// takes s, returns real (source) path
func (r Runner) pathSyncedReverse(s string) (string, error) {
	withoutSynced, found := strings.CutPrefix(s, r.synced)
	if !found {
		return "", errors.New("can't convert synced path to real path")
	}

	return path.Join(r.home, withoutSynced), nil
}

// also creates all directories up to `to`
func write(from string, to string) error {
	f, err := os.Open(from)
	if err != nil {
		return err
	}
	defer f.Close()

	fStat, err := f.Stat()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path.Dir(to), 0o770); err != nil {
		return err
	}

	t, err := os.OpenFile(to, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, fStat.Mode())
	if err != nil {
		return err
	}
	defer t.Close()

	_, err = io.Copy(t, f)
	if err != nil {
		return err
	}

	return nil
}

// also creates all directories up to `to`
func zipDir(from fs.FS, to string) error {
	if err := os.MkdirAll(path.Dir(to), 0o770); err != nil {
		return err
	}

	toFile, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o660)
	if err != nil {
		return err
	}
	defer toFile.Close()

	writer := zip.NewWriter(toFile)
	defer writer.Close()

	return writer.AddFS(from)
}

func (r Runner) normalize(p string) string {
	if path.IsAbs(p) {
		return p
	}

	return path.Join(r.wd, p)
}

// returns count of errChan writes
func (r Runner) addPath(p string, errChan chan error) int {
	p = r.normalize(p)

	stat, err := os.Stat(p)
	if err != nil {
		go func() { errChan <- err }()
		return 1
	}

	syncedDir, err := r.pathSynced(p)
	if err != nil {
		go func() { errChan <- err }()
		return 1
	}

	if path.Base(p) == ".git" {
		go func() { errChan <- zipDir(os.DirFS(p), syncedDir+".zip") }()
		r.logger.Println("zipping", p)
		return 1
	}

	if !stat.IsDir() {
		go func() { errChan <- write(p, syncedDir) }()
		r.logger.Println("syncing", p)
		return 1
	}

	entries, err := os.ReadDir(p)
	if err != nil {
		go func() { errChan <- err }()
		return 1
	}

	c := 0
	for _, v := range entries {
		c += r.addPath(path.Join(p, v.Name()), errChan)
	}
	return c
}

// returns count of errChan writes
func (r Runner) addSynced(dir string, errChan chan error) int {
	syncedEntries, err := os.ReadDir(dir)
	if err != nil {
		go func() { errChan <- err }()
		return 1
	}

	c := 0

	for _, v := range syncedEntries {
		syncedPath := path.Join(dir, v.Name())

		realPath, err := r.pathSyncedReverse(syncedPath)
		if err != nil {
			go func() { errChan <- err }()
			c += 1
			continue
		}

		if path.Base(syncedPath) == ".git.zip" {
			continue
		}

		syncedStat, err := v.Info()
		if err != nil {
			go func() { errChan <- err }()
			c += 1
			continue
		}

		// assume that error means realPath does not exist
		// delete our own synced path then
		if _, err := os.Stat(realPath); err != nil {
			go func() { errChan <- os.RemoveAll(syncedPath) }()
			r.logger.Println("removing", syncedPath)
			c += 1
			continue
		}

		if syncedStat.IsDir() {
			c += r.addSynced(syncedPath, errChan)
		} else {
			go func() { errChan <- write(realPath, syncedPath) }()
			r.logger.Println("syncing", realPath)
			c += 1
		}
	}

	return c
}

func (r Runner) Sync(errChan chan error) int {
}

func (r Runner) Add(p string) []error {
	errChan := make(chan error)

	var count int
	if p == "" {
		count = r.addSynced(r.synced, errChan)
	} else {
		count = r.addPath(p, errChan)
	}

	errors := []error{}

	for range count {
		err := <-errChan
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
