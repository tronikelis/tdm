package runner

import "sync/atomic"

type asyncRunner struct {
	errChan chan error
	stack   *atomic.Int64
}

func newAsyncRunner() asyncRunner {
	return asyncRunner{
		errChan: make(chan error),
		stack:   &atomic.Int64{},
	}
}

func (a asyncRunner) Run(task func() error) {
	a.stack.Add(1)
	go func() {
		a.errChan <- task()
		a.stack.Add(-1)
		if a.stack.Load() == 0 {
			close(a.errChan)
		}
	}()
}

func (a asyncRunner) WaitErrors() []error {
	errors := []error{}

	for err := range a.errChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
