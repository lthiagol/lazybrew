package task

import (
	"context"
	"errors"
	"sync"

	tea "charm.land/bubbletea/v2"
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

// taskOutputBatch is the most output lines one TaskOutputMsg carries.
const taskOutputBatch = 64

type runningTask struct {
	task  *Task
	ctx   context.Context
	outCh <-chan string
	errCh <-chan error
	// err is the task's error, once errCh has delivered it.
	err error
	// started is set once Run has been called, by the first reader.
	started bool
	// reading is set while a readChunk Cmd for the task is out: RunNext issues
	// no second one, so the task has one reader at a time.
	reading bool
	// done is set when the reader returned the task's TaskCompletedMsg. The
	// task stays current until Update hands the completion back (RunNext).
	done bool
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
	// A done task whose completion Update has not handled yet still counts:
	// only the RunNext of that completion clears it (FC2).
	canStart := m.current == nil
	if !canStart {
		total++
	}
	if total >= m.max {
		return EnqueueQueued, ErrQueueFull
	}

	m.queue = append(m.queue, t)
	if canStart {
		return EnqueueStarted, nil
	}
	return EnqueueQueued, nil
}

// Idle reports that no task is running and none is queued.
func (m *Manager) Idle() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.current == nil && len(m.queue) == 0
}

func (m *Manager) CancelCurrent() {
	m.mu.Lock()
	rt := m.current
	m.mu.Unlock()

	if rt != nil && rt.task.Cancel != nil {
		rt.task.Cancel()
	}
}

// RunNext is called from Update after every task message. It clears a task
// whose completion has come back, issues the running task's next reader (at
// most one at a time), or starts the next queued task. Starting never calls
// Task.Run here: the task becomes current, with its Cancel set, and the first
// reader calls Run on the Cmd goroutine, so a Run that blocks never holds
// Update and Esc can cancel it at once (FC1).
func (m *Manager) RunNext() tea.Cmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.current != nil && m.current.done {
		m.current = nil
	}

	if m.current != nil {
		rt := m.current
		if rt.reading {
			return nil
		}
		rt.reading = true
		return m.readChunk(rt)
	}

	if len(m.queue) == 0 {
		return nil
	}

	next := m.queue[0]
	m.queue = m.queue[1:]

	ctx, cancel := context.WithCancel(context.Background())
	next.Cancel = cancel
	m.current = &runningTask{task: next, ctx: ctx}

	return func() tea.Msg {
		return TaskStartedMsg{ID: next.ID, Title: next.Title}
	}
}

// readChunk is the next message of a running task: a batch of its output, or
// its completion once the output is done and the error is in. The first one
// calls the task's Run. Only one is out per task at a time (rt.reading), so
// the fields it reads and writes have one owner.
func (m *Manager) readChunk(rt *runningTask) tea.Cmd {
	return func() tea.Msg {
		msg := rt.read()
		m.mu.Lock()
		rt.reading = false
		if _, ok := msg.(TaskCompletedMsg); ok {
			rt.done = true
		}
		m.mu.Unlock()
		return msg
	}
}

// read is the body of readChunk.
func (rt *runningTask) read() tea.Msg {
	if !rt.started {
		rt.started = true
		outCh, errCh, err := rt.task.Run(rt.ctx)
		if err != nil {
			return TaskCompletedMsg{ID: rt.task.ID, Title: rt.task.Title, Err: err}
		}
		rt.outCh, rt.errCh = outCh, errCh
	}
	for rt.outCh != nil {
		if rt.errCh != nil {
			// The task is still running: take whichever comes first, a line
			// or the error.
			select {
			case line, ok := <-rt.outCh:
				if ok {
					return TaskOutputMsg{ID: rt.task.ID, Lines: rt.batch(line)}
				}
				rt.outCh = nil
			case err := <-rt.errCh:
				rt.err = err
				rt.errCh = nil
			}
			continue
		}
		// The error is in, and brew's lines may still be buffered behind it.
		// They come before the completion: read up to the close of outCh,
		// or until the task is cancelled (C10).
		select {
		case line, ok := <-rt.outCh:
			if ok {
				return TaskOutputMsg{ID: rt.task.ID, Lines: rt.batch(line)}
			}
			rt.outCh = nil
		case <-rt.ctx.Done():
			rt.outCh = nil
		}
	}
	if rt.errCh != nil {
		rt.err = <-rt.errCh
		rt.errCh = nil
	}
	return TaskCompletedMsg{
		ID:    rt.task.ID,
		Title: rt.task.Title,
		Err:   rt.err,
	}
}

// batch is first, then the lines already buffered behind it, up to
// taskOutputBatch. It never blocks: a line not yet buffered waits for the next
// message. A close of outCh seen here is recorded, so the next read sees it.
func (rt *runningTask) batch(first string) []string {
	lines := []string{first}
	for len(lines) < taskOutputBatch {
		select {
		case line, ok := <-rt.outCh:
			if !ok {
				rt.outCh = nil
				return lines
			}
			lines = append(lines, line)
		default:
			return lines
		}
	}
	return lines
}
