package task

import (
	"context"
	"errors"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

var ErrQueueFull = errors.New("task queue is full")

const DefaultMaxQueue = 10

type Manager struct {
	mu      sync.Mutex
	queue   []*Task
	current *Task
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
	return m.current != nil
}

func (m *Manager) Enqueue(t *Task) (started bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	total := len(m.queue)
	if m.current != nil {
		total++
	}
	if total >= m.max {
		return false, ErrQueueFull
	}

	t.Status = StatusPending
	m.queue = append(m.queue, t)
	return false, nil
}

func (m *Manager) CancelCurrent() {
	m.mu.Lock()
	current := m.current
	m.mu.Unlock()

	if current != nil && current.Cancel != nil {
		current.Cancel()
	}
}

func (m *Manager) RunNext() tea.Cmd {
	m.mu.Lock()
	if m.current != nil {
		m.mu.Unlock()
		return nil
	}
	if len(m.queue) == 0 {
		m.mu.Unlock()
		return nil
	}

	next := m.queue[0]
	m.queue = m.queue[1:]
	next.Status = StatusRunning
	m.current = next
	m.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	next.Cancel = cancel

	outCh, errCh, err := next.Run(ctx)
	if err != nil {
		next.Status = StatusFailed
		m.clearCurrent()
		return func() tea.Msg {
			return TaskCompletedMsg{
				ID:    next.ID,
				Title: next.Title,
				Err:   err,
			}
		}
	}

	return func() tea.Msg {
		var runErr error
		if errCh != nil {
			runErr = <-errCh
		}

		m.mu.Lock()
		if runErr != nil {
			next.Status = StatusFailed
		} else {
			next.Status = StatusSuccess
		}
		m.mu.Unlock()
		m.clearCurrent()

		_ = outCh

		return TaskCompletedMsg{
			ID:    next.ID,
			Title: next.Title,
			Err:   runErr,
		}
	}
}

func (m *Manager) clearCurrent() {
	m.mu.Lock()
	m.current = nil
	m.mu.Unlock()
}
