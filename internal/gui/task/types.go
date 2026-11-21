package task

import "context"

// Task represents a single serialized operation.
type Task struct {
	ID     string
	Title  string
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
