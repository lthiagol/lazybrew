package task

import "context"

// Task represents a single serialized operation.
type Task struct {
	ID    string
	Title string
	// Run starts the work and returns its output lines and its one error. The
	// manager calls it once, on a Cmd goroutine (never in Update), so it may
	// block; it must stop when ctx is cancelled. A non-nil third result is a
	// task that failed to start.
	Run    func(ctx context.Context) (<-chan string, <-chan error, error)
	Cancel context.CancelFunc
}

// TaskStartedMsg is emitted when a task begins execution.
type TaskStartedMsg struct {
	ID    string
	Title string
}

// TaskOutputMsg carries the next lines of streaming output from a running
// task, in order: up to taskOutputBatch of them in one message.
type TaskOutputMsg struct {
	ID    string
	Lines []string
}

// TaskCompletedMsg signals that a task has finished.
type TaskCompletedMsg struct {
	ID    string
	Title string
	Err   error
}
