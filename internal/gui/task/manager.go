package task

import (
	"context"
	"errors"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var ErrQueueFull = errors.New("task queue is full")

// EnqueueResult reports what happened to a task handed to Enqueue.
type EnqueueResult int

const (
	// EnqueueStarted means the queue was empty and no task is running:
	// the caller must invoke RunNext to actually start the task.
	EnqueueStarted EnqueueResult = iota
	// EnqueueQueued means another task is already running: the task was
	// parked in the queue and will start when the running one completes.
	EnqueueQueued
)

const DefaultMaxQueue = 10

type runningTask struct {
	task  *Task
	outCh <-chan string
	errCh <-chan error
	done  bool
}

type Manager struct {
	mu      sync.Mutex
	queue   []*Task
	current *runningTask
	max     int
}

func NewManager(maxQueue int) *Manager {
	if maxQueue <= 0 {
		maxQueue = DefaultMaxQueue
	}
	return &Manager{
		queue: make([]*Task, 0, maxQueue),
		max:   maxQueue,
	}
}

func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.current != nil && !m.current.done
}

// Enqueue parks the task in the queue. It reports EnqueueStarted when the
// queue is empty and nothing is running (the caller must then invoke
// RunNext to start the task), and EnqueueQueued when another task is
// already running — the task is parked and will run when the current one
// completes, so callers can tell the user "queued" instead of claiming
// the operation was refused (r1 finding 6).
func (m *Manager) Enqueue(t *Task) (EnqueueResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	total := len(m.queue)
	canStart := m.current == nil || m.current.done
	if !canStart {
		total++
	}
	if total >= m.max {
		return EnqueueQueued, ErrQueueFull
	}

	t.Status = StatusPending
	m.queue = append(m.queue, t)
	if canStart {
		return EnqueueStarted, nil
	}
	return EnqueueQueued, nil
}

func (m *Manager) CancelCurrent() {
	m.mu.Lock()
	rt := m.current
	m.mu.Unlock()

	if rt != nil && rt.task.Cancel != nil {
		rt.task.Cancel()
	}
}

func (m *Manager) RunNext() tea.Cmd {
	m.mu.Lock()

	if m.current != nil && m.current.done {
		m.current = nil
	}

	if m.current != nil {
		rt := m.current
		m.mu.Unlock()
		return m.readChunk(rt)
	}

	if len(m.queue) == 0 {
		m.mu.Unlock()
		return nil
	}

	next := m.queue[0]
	m.queue = m.queue[1:]
	m.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	next.Cancel = cancel
	next.Status = StatusRunning

	outCh, errCh, err := next.Run(ctx)
	if err != nil {
		next.Status = StatusFailed
		return func() tea.Msg {
			return TaskCompletedMsg{
				ID:    next.ID,
				Title: next.Title,
				Err:   err,
			}
		}
	}

	rt := &runningTask{
		task:  next,
		outCh: outCh,
		errCh: errCh,
	}
	m.mu.Lock()
	m.current = rt
	m.mu.Unlock()

	return func() tea.Msg {
		return TaskStartedMsg{ID: next.ID, Title: next.Title}
	}
}

func (m *Manager) readChunk(rt *runningTask) tea.Cmd {
	return func() tea.Msg {
		if rt.outCh == nil {
			var err error
			if rt.errCh != nil {
				err = <-rt.errCh
				rt.errCh = nil
			}
			m.mu.Lock()
			rt.done = true
			rt.task.Status = statusFromErr(err)
			m.mu.Unlock()
			return TaskCompletedMsg{
				ID:    rt.task.ID,
				Title: rt.task.Title,
				Err:   err,
			}
		}

		// Prefer draining buffered output lines before handling the
		// error channel: the error may already be buffered while up to a
		// full channel of output lines is still pending, and taking the
		// error first would drop them (r1 finding 15).
		select {
		case line, ok := <-rt.outCh:
			if ok {
				return TaskOutputMsg{ID: rt.task.ID, Line: line}
			}
			rt.outCh = nil
			return m.readChunk(rt)()
		default:
		}

		select {
		case line, ok := <-rt.outCh:
			if ok {
				return TaskOutputMsg{ID: rt.task.ID, Line: line}
			}
			rt.outCh = nil
			return m.readChunk(rt)()
		case err := <-rt.errCh:
			rt.outCh = nil
			rt.errCh = nil
			m.mu.Lock()
			rt.done = true
			rt.task.Status = statusFromErr(err)
			m.mu.Unlock()
			return TaskCompletedMsg{
				ID:    rt.task.ID,
				Title: rt.task.Title,
				Err:   err,
			}
		}
	}
}

func statusFromErr(err error) Status {
	if err == nil {
		return StatusSuccess
	}
	if errors.Is(err, context.Canceled) {
		return StatusCancelled
	}
	return StatusFailed
}
