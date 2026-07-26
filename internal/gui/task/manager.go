package task

import (
	"context"
	"errors"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var ErrQueueFull = errors.New("task queue is full")

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

func (m *Manager) Enqueue(t *Task) (started bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	total := len(m.queue)
	canStart := m.current == nil || m.current.done
	if canStart {
		started = true
	} else {
		total++
	}
	if total >= m.max {
		return false, ErrQueueFull
	}

	t.Status = StatusPending
	m.queue = append(m.queue, t)
	return started, nil
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
			rt.done = true
			rt.task.Status = statusFromErr(err)
			return TaskCompletedMsg{
				ID:    rt.task.ID,
				Title: rt.task.Title,
				Err:   err,
			}
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
			rt.done = true
			rt.task.Status = statusFromErr(err)
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
	return StatusFailed
}
