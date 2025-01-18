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
// if nil queue, creates it
func (r Runner) addPath(p string, queue *taskQueue) *taskQueue {
	if queue == nil {
		queue = newTaskQueue()
	}
	p = r.normalize(p)

	stat, err := os.Stat(p)
	if err != nil {
		queue.err(err)
		return queue
	}

	syncedDir, err := r.pathSynced(p)
	if err != nil {
		queue.err(err)
		return queue
	}

	if path.Base(p) == ".git" {
		queue.add(func() error { return zipDir(os.DirFS(p), syncedDir+".zip") })
		r.logger.Println("zipping", p)
		return queue
	}

	if !stat.IsDir() {
		queue.add(func() error { return write(p, syncedDir) })
		r.logger.Println("adding", p)
		return queue
	}

	entries, err := os.ReadDir(p)
	if err != nil {
		queue.err(err)
		return queue
	}

	for _, v := range entries {
		r.addPath(path.Join(p, v.Name()), queue)
	}

	return queue
}

// synced <- real
// call with empty dir ("")
// if nil queue, creates it
func (r Runner) addSynced(dir string, queue *taskQueue) *taskQueue {
	if queue == nil {
		queue = newTaskQueue()
	}
	if dir == "" {
		dir = r.synced
	}

	syncedEntries, err := os.ReadDir(dir)
	if err != nil {
		queue.err(err)
		return queue
	}

	for _, v := range syncedEntries {
		syncedPath := path.Join(dir, v.Name())

		if path.Base(syncedPath) == ".git.zip" {
			continue
		}

		realPath, err := r.pathSyncedReverse(syncedPath)
		if err != nil {
			queue.err(err)
			continue
		}

		syncedStat, err := v.Info()
		if err != nil {
			queue.err(err)
			continue
		}

		// assume that error means realPath does not exist
		// delete our own synced path then
		if _, err := os.Stat(realPath); err != nil {
			queue.add(func() error { return os.RemoveAll(syncedPath) })
			r.logger.Println("removing", syncedPath)
			continue
		}

		if syncedStat.IsDir() {
			r.addSynced(syncedPath, queue)
		} else {
			queue.add(func() error { return write(realPath, syncedPath) })
			r.logger.Println("adding", realPath)
		}
	}

	return queue
}

// synced -> real
// call with empty dir ("")
// if nil queue, creates it
func (r Runner) sync(dir string, queue *taskQueue) *taskQueue {
	if queue == nil {
		queue = newTaskQueue()
	}
	if dir == "" {
		dir = r.synced
	}

	syncedEntries, err := os.ReadDir(dir)
	if err != nil {
		queue.err(err)
		return queue
	}

	for _, v := range syncedEntries {
		syncedPath := path.Join(dir, v.Name())

		realPath, err := r.pathSyncedReverse(syncedPath)
		if err != nil {
			queue.err(err)
			continue
		}

		// sync .git here, remove .git from real, then unzip into real .git
		if path.Base(syncedPath) == ".git.zip" {
			gitPath := path.Join(path.Dir(realPath), ".git")

			queue.add(func() error {
				if err := os.RemoveAll(gitPath); err != nil {
					return err
				}

				return unzipDir(syncedPath, gitPath)
			})

			r.logger.Println("unzipping", syncedPath)
			continue
		}

		stat, err := v.Info()
		if err != nil {
			queue.err(err)
			continue
		}

		if stat.IsDir() {
			r.sync(path.Join(dir, v.Name()), queue)
		} else {
			queue.add(func() error { return write(syncedPath, realPath) })
			r.logger.Println("syncing", syncedPath)
		}
	}

	return queue
}

func (r Runner) UserSync() []error {
	return r.sync("", nil).wait()
}

func (r Runner) UserAdd(p string) []error {
	if p == "" {
		return r.addSynced("", nil).wait()
	}

	return r.addPath(p, nil).wait()
}
