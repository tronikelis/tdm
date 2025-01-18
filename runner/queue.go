package runner

// methods on this are not concurrent-safe,
// tasks are, but methods are not
type taskQueue struct {
	count   int
	errChan chan error
}

func newTaskQueue() *taskQueue {
	return &taskQueue{errChan: make(chan error)}
}

// schedules task in the bg
func (t *taskQueue) add(task func() error) {
	go func() { t.errChan <- task() }()
	t.count += 1
}

// wrapper around .add, adds task that instantly returns error
func (t *taskQueue) err(err error) {
	t.add(func() error { return err })
}

func (t *taskQueue) wait() []error {
	errors := []error{}

	for range t.count {
		err := <-t.errChan
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
