package task

import "context"

// Status represents the current state of a task.
type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusSuccess
	StatusFailed
	StatusCancelled
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusSuccess:
		return "success"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Task represents a single serialized operation.
type Task struct {
	ID     string
	Title  string
	Status Status
	Run    func(ctx context.Context) (<-chan string, <-chan error, error)
	Cancel context.CancelFunc
}

// TaskStartedMsg is emitted when a task begins execution.
type TaskStartedMsg struct {
	ID    string
	Title string
}

// TaskOutputMsg carries a single line of streaming output from a running task.
type TaskOutputMsg struct {
	ID   string
	Line string
}

// TaskCompletedMsg signals that a task has finished.
type TaskCompletedMsg struct {
	ID    string
	Title string
	Err   error
}

// TaskRejectedMsg is sent when Enqueue fails (e.g. queue full).
type TaskRejectedMsg struct {
	Reason string
}
