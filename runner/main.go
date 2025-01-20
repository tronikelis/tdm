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

// dir -> synced
func (r Runner) addPath(p string, async asyncRunner) {
	p = r.normalize(p)

	async.Run(func() error {
		stat, err := os.Stat(p)
		if err != nil {
			return err
		}

		syncedPath, err := r.pathSynced(p)
		if err != nil {
			return err
		}

		if path.Base(p) == ".git" {
			syncedPath = syncedPath + ".zip"
			if _, err := os.Stat(syncedPath); err == nil { // do not override git dir
				return nil
			}

			r.logger.Println("zipping", p)
			return zipDir(os.DirFS(p), syncedPath)
		}

		if !stat.IsDir() {
			r.logger.Println("adding", p)
			return write(p, syncedPath)
		}

		entries, err := os.ReadDir(p)
		if err != nil {
			return err
		}

		for _, v := range entries {
			r.addPath(path.Join(p, v.Name()), async)
		}

		return nil
	})
}

// synced <- real
// call with empty dir ("")
func (r Runner) addSynced(dir string, async asyncRunner) {
	if dir == "" {
		dir = r.synced
	}

	async.Run(func() error {
		syncedEntries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		for _, v := range syncedEntries {
			async.Run(func() error {
				syncedPath := path.Join(dir, v.Name())

				if path.Base(syncedPath) == ".git.zip" {
					return nil
				}

				realPath, err := r.pathSyncedReverse(syncedPath)
				if err != nil {
					return err
				}

				syncedStat, err := v.Info()
				if err != nil {
					return err
				}

				// assume that error means realPath does not exist
				// delete our own synced path then
				if _, err := os.Stat(realPath); err != nil {
					r.logger.Println("removing", syncedPath)
					return os.RemoveAll(syncedPath)
				}

				if syncedStat.IsDir() {
					r.addSynced(syncedPath, async)
				} else {
					r.logger.Println("adding", realPath)
					return write(realPath, syncedPath)
				}

				return nil
			})
		}

		return nil
	})
}

// synced -> real
// call with empty dir ("")
func (r Runner) sync(dir string, async asyncRunner) {
	if dir == "" {
		dir = r.synced
	}

	async.Run(func() error {
		syncedEntries, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		for _, v := range syncedEntries {
			async.Run(func() error {
				syncedPath := path.Join(dir, v.Name())

				realPath, err := r.pathSyncedReverse(syncedPath)
				if err != nil {
					return err
				}

				// sync .git here, remove .git from real, then unzip into real .git
				if path.Base(syncedPath) == ".git.zip" {
					gitPath := path.Join(path.Dir(realPath), ".git")
					if _, err := os.Stat(gitPath); err == nil { // do not override existing git dir
						return err
					}

					r.logger.Println("unzipping", syncedPath)
					return unzipDir(syncedPath, gitPath)
				}

				stat, err := v.Info()
				if err != nil {
					return err
				}

				if stat.IsDir() {
					r.sync(path.Join(dir, v.Name()), async)
				} else {
					r.logger.Println("syncing", syncedPath)
					return write(syncedPath, realPath)
				}

				return nil
			})
		}
		return nil
	})
}

func (r Runner) UserSync() []error {
	async := newAsyncRunner()
	r.sync("", async)
	return async.WaitErrors()
}

func (r Runner) UserAdd(p string) []error {
	async := newAsyncRunner()

	if p == "" {
		r.addSynced("", async)
		return async.WaitErrors()
	}

	r.addPath(p, async)
	return async.WaitErrors()
}
