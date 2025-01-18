package runner

import (
	"errors"
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
		r.logger.Println("adding", p)
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
// call with empty dir ("")
func (r Runner) addSynced(dir string, errChan chan error) int {
	syncedEntries, err := os.ReadDir(dir)
	if err != nil {
		go func() { errChan <- err }()
		return 1
	}

	c := 0

	for _, v := range syncedEntries {
		syncedPath := path.Join(dir, v.Name())

		if path.Base(syncedPath) == ".git.zip" {
			continue
		}

		realPath, err := r.pathSyncedReverse(syncedPath)
		if err != nil {
			go func() { errChan <- err }()
			c += 1
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
			r.logger.Println("adding", realPath)
			c += 1
		}
	}

	return c
}

// call with empty dir ("")
func (r Runner) sync(dir string, errChan chan error) int {
	if dir == "" {
		dir = r.synced
	}

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

		// sync .git here, remove .git from real, then unzip into real .git
		if path.Base(syncedPath) == ".git.zip" {
			gitPath := path.Join(path.Dir(realPath), ".git")
			go func() {
				if err := os.RemoveAll(gitPath); err != nil {
					errChan <- err
					return
				}

				errChan <- unzipDir(syncedPath, gitPath)
			}()
			r.logger.Println("unzipping", syncedPath)
			c += 1
			continue
		}

		stat, err := v.Info()
		if err != nil {
			go func() { errChan <- err }()
			c += 1
			continue
		}

		if stat.IsDir() {
			c += r.sync(path.Join(dir, v.Name()), errChan)
		} else {
			go func() { errChan <- write(syncedPath, realPath) }()
			r.logger.Println("syncing", syncedPath)
			c += 1
		}
	}

	return c
}

func waitForErrors(fn func(errChan chan error) int) []error {
	errChan := make(chan error)
	count := fn(errChan)

	errors := []error{}

	for range count {
		err := <-errChan
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

func (r Runner) UserSync() []error {
	return waitForErrors(func(errChan chan error) int {
		return r.sync("", errChan)
	})
}

func (r Runner) UserAdd(p string) []error {
	if p == "" {
		return waitForErrors(func(errChan chan error) int {
			return r.addSynced("", errChan)
		})
	}

	return waitForErrors(func(errChan chan error) int {
		return r.addPath(p, errChan)
	})
}
