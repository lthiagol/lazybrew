# Milestone 19 — Bubble Tea Concurrency & Task Manager

> **Status:** 🔜 Planned  
> **Size estimate:** L (4–6 days)  
> **Depends on:** M13 ✅, M18.5 (DESIGN concurrency ADR)  
> **Enables:** M20 phase C, M21 T1, M17  
> **Parallel track:** B (Engineering) — starts after M18.5  
> **Gate criteria:** Zero `program.Send` in commands; TaskManager owns all writes; race tests pass

See [planning-challenge-2026-06-13.md](../planning-challenge-2026-06-13.md) for challenged TaskManager design.

---

## Goal

Replace ad-hoc goroutines and `program.Send` with a **single TaskManager** that serializes write operations, streams output through Bubble Tea messages, and never leaves the UI in a stuck "busy" state.

---

## Out of Scope

- Batch upgrade logic (M20.3 — uses TaskManager but defined there)
- Controller package split (backlog B-01)
- Runner SIGKILL on cancel (future; M19 uses context cancel as today)

---

## Architecture Decisions (ADRs)

| ID | Decision | Alternatives rejected |
|---|---|---|
| D19-1 | Package `internal/gui/task/` | Top-level package — too scattered |
| D19-2 | Zero `program.Send` in handlers | Keep pump — violates Bubble Tea purity |
| D19-3 | Reads outside TaskManager | Queue reads — unnecessary blocking |
| D19-4 | Progress modal on Model, not Manager | Manager owns UI — coupling |
| D19-5 | Queue max 10 tasks | Unbounded — memory risk |

---

## Step Index

| Step | Title | Size | Depends | Deliverable |
|---|---|---|---|---|
| 19.0 | TypedCache safe get | S | — | No panic on wrong type |
| 19.1 | Task types + messages | S | 18.5 | `task/types.go`, msg types |
| 19.2 | TaskManager core (sequential) | M | 19.1 | Queue + one-at-a-time |
| 19.3 | TaskManager streaming cmd | M | 19.2 | Output lines as messages |
| 19.4 | TaskManager cancel | S | 19.3 | CancelFunc wired |
| 19.5 | Wire Manager into Model.Update | M | 19.4 | Message handlers |
| 19.6 | Migrate doMutation | M | 19.5 | No goroutine in doMutation |
| 19.7 | Migrate service + tap writes | M | 19.6 | serviceAction, trust, untap, repair |
| 19.8 | Migrate streaming diagnostics | M | 19.5 | doctor, missing, vulns, cleanup, brewfile |
| 19.9 | Remove isBusy + audit Send | S | 19.6–19.8 | Clean Model |
| 19.10 | Milestone verification | S | 19.9 | Audit + race tests |

---

## Steps

### 19.0 — TypedCache Safe Get

**Size:** S · **Moved from M21** — do before concurrency work

**Files:** `internal/brew/cache.go`, `cache_test.go`

**Implementation:**
```go
func (tc *TypedCache[T]) Get() (T, bool) {
    val, ok := tc.cache.Get(tc.key)
    if !ok { var zero T; return zero, false }
    typed, ok := val.(T)
    if !ok {
        tc.cache.Invalidate(tc.key)
        var zero T
        return zero, false
    }
    return typed, true
}
```

**Acceptance criteria:**
- [ ] Wrong type stored → miss + key invalidated, no panic

**Tests:** `TestTypedCacheWrongTypeReturnsMiss`

---

### 19.1 — Task Types + Messages

**Size:** S · **Depends on:** 18.5

**Files to create:**
- `internal/gui/task/types.go`
- `internal/gui/task/doc.go` (package comment)
- Extend `internal/gui/messages.go` OR keep task messages in task package with re-exports

**Types:**
```go
type Task struct {
    ID       string
    Title    string
    Run      func(ctx context.Context) (<-chan string, <-chan error, error)
    Cancel   context.CancelFunc
}

type Status int // Pending, Running, Success, Failed, Cancelled
```

**Messages (add to gui/messages.go):**
```go
type TaskStartedMsg struct { ID, Title string }
type TaskOutputMsg struct { ID, Line string }
type TaskCompletedMsg struct { ID, Title string; Err error }
type TaskRejectedMsg struct { Reason string } // queue full or already running
```

**Acceptance criteria:**
- [ ] Compiles; documented in DESIGN.md
- [ ] Message names don't collide with existing ProgressLineMsg (decide mapping in 19.5)

**Tests:** `TestTaskStatusString` if exported

---

### 19.2 — TaskManager Core

**Size:** M · **Depends on:** 19.1

**File:** `internal/gui/task/manager.go`

**API:**
```go
type Manager struct { /* queue, current, mu */ }

func NewManager(maxQueue int) *Manager
func (m *Manager) IsRunning() bool
func (m *Manager) Enqueue(t Task) (started bool, err error) // err if queue full
func (m *Manager) CancelCurrent()
func (m *Manager) RunNext() tea.Cmd // returns nil if idle; starts next task
```

**Behavior:**
- Only one task Running at a time
- Pending tasks FIFO (max 10)
- `Enqueue` while running → append to queue, return `started=false`
- `Enqueue` while idle → set current, return `started=true` + caller invokes `RunNext()`

**Acceptance criteria:**
- [ ] Second task waits until first completes
- [ ] 11th enqueue returns error / TaskRejectedMsg

**Tests:**
- [ ] `TestManager_Sequential`
- [ ] `TestManager_QueueFull`

---

### 19.3 — TaskManager Streaming as tea.Cmd

**Size:** M · **Depends on:** 19.2

**Design:** One `tea.Cmd` per running task that:
1. Reads lines from channel
2. Returns **batched** `TaskOutputMsg` via custom batch message OR returns single multi-line message

**Recommended pattern (Bubble Tea idiomatic):**
```go
// Cmd returns first output or completion; Update re-arms with tea.Batch
func (m *Manager) runCurrentCmd() tea.Cmd {
    return func() tea.Msg {
        // read one line or wait for completion
        select {
        case line, ok := <-ch:
            if ok { return TaskOutputMsg{Line: line} }
            // channel closed, wait err
        case err := <-errCh:
            return TaskCompletedMsg{Err: err}
        }
    }
}
```

In `Update`, on `TaskOutputMsg` → re-queue `runCurrentCmd()` until `TaskCompletedMsg`.

**Acceptance criteria:**
- [ ] No `program.Send` anywhere in this path
- [ ] Output order preserved

**Tests:** `TestManager_OutputStreamingOrder`

---

### 19.4 — TaskManager Cancel

**Size:** S · **Depends on:** 19.3

**Implementation:**
1. Progress modal cancel button calls `manager.CancelCurrent()`
2. Cancel invokes task's `context.CancelFunc`
3. Emit `TaskCompletedMsg{Err: context.Canceled}`

**Acceptance criteria:**
- [ ] Cancel mid-stream stops task; queue advances to next

**Tests:** `TestManager_Cancel`

---

### 19.5 — Wire Manager into Model.Update

**Size:** M · **Depends on:** 19.4 · **Blocks:** 19.6

**Files:** `gui.go`, possibly `task.go` (move batchState only; manager on Model)

**Model changes:**
```go
type Model struct {
    tasks *task.Manager
    // remove: isBusy
}
```

**Update cases:**
| Message | Action |
|---|---|
| `TaskStartedMsg` | Open ProgressModal with title |
| `TaskOutputMsg` | Append to modal OR map to ProgressLineMsg |
| `TaskCompletedMsg` | Close modal, toast, invalidate refresh, `RunNext()` |
| `TaskRejectedMsg` | Toast warning |

**Mapping decision:** Keep `ProgressLineMsg` internally — TaskOutputMsg handler converts to avoid rewriting modal.

**Acceptance criteria:**
- [ ] Manager initialized in `New()`
- [ ] Idle manager has no modal

**Tests:** `TestModelTaskStartedOpensModal`

---

### 19.6 — Migrate doMutation

**Size:** M · **Depends on:** 19.5

**File:** `commands.go` — `doMutation`

**Implementation checklist:**
1. **Validate first** (selection, name, panel) — before Enqueue
2. Build `task.Task` with Run closure calling appropriate `FormulaeWrite`/`CasksWrite`/Runner stream
3. `manager.Enqueue(task)` — if rejected, toast and return
4. Delete entire `go func()` block (lines ~269–312)
5. Delete `m.program == nil` early return or handle gracefully without stuck state

**Acceptance criteria:**
- [ ] Install/uninstall/reinstall/upgrade/fetch all use manager
- [ ] Invalid selection → toast or silent return, **manager not running**
- [ ] `go test -race` clean for mutation tests

**Tests:**
- [ ] `TestDoMutationInvalidSelectionNotStuck`
- [ ] `TestDoMutationRejectedWhenRunning`
- [ ] Update existing `TestDoMutationWhenBusy` for TaskManager

---

### 19.7 — Migrate Service + Tap Writes

**Size:** M · **Depends on:** 19.6

**Functions to migrate:**
| Function | Task type | Notes |
|---|---|---|
| `serviceAction` | Short write | start/stop/restart/run |
| `executeTrustAction` | Write | trust/untrust tap |
| `trustItemCmd` | Write | formula/cask trust |
| `executeUntap` | Write | after confirm |
| `executeRepair` | Write | after confirm |
| `togglePin` | Write | fast; still serialized |

**Pattern:** Replace `return m, func() tea.Msg { ... MutationResultMsg }` with Task + completion handler that sends refresh.

**Acceptance criteria:**
- [ ] Rapid `s` + `i` doesn't run concurrent brew writes
- [ ] Trust/untrust shows progress modal

**Tests:** `TestServiceActionQueuedWhenBusy`

---

### 19.8 — Migrate Streaming Diagnostics

**Size:** M · **Depends on:** 19.5

**Functions currently using `program.Send`:**
- `runDoctor`
- `runMissing`
- `runVulns`
- `brewCleanup` (preview + execute paths)
- `executeBrewfileAction`

**Implementation:** Each becomes a Task with stream reader; output via TaskOutputMsg chain.

**Special case — `runMissing`/`runVulns`:** Collect formatted lines in Run closure; stream to modal.

**Acceptance criteria:**
- [ ] `grep -r 'program\.Send' internal/gui/` → **zero matches** outside tests (if any)

**Tests:**
- [ ] `TestRunDoctorStreamsOutput` (mock runner)
- [ ] `TestRunVulnsNoProgramSend` (static)

---

### 19.9 — Remove isBusy + Dead Code

**Size:** S · **Depends on:** 19.6–19.8

**Implementation:**
1. Remove `isBusy` field from Model
2. Remove `ProgressCompleteMsg` paths that only existed for goroutine pattern (consolidate to TaskCompletedMsg OR keep both with clear ownership — document in DESIGN)
3. Remove `confirmCallback` pattern for serviceCleanup if superseded by pendingAction (audit — keep one pattern)

**Acceptance criteria:**
- [ ] `isBusy` grep → zero
- [ ] `program.Send` grep → zero in production gui code

---

### 19.10 — Milestone Verification

**Size:** S · **Depends on:** 19.9 · **Requires:** M18.8 AGENTS.md merged

**Audit script (manual or Makefile target):**
```bash
! rg 'program\.Send' internal/gui --glob '!*_test.go'
! rg 'isBusy' internal/gui
go test -race ./...
```

**Manual smoke:**
1. Install package (mock or real)
2. Cancel mid-install
3. Queue two operations — second waits
4. Doctor output appears in modal

**Acceptance criteria:**
- [ ] Audit passes
- [ ] Manual smoke signed off in PR description

---

## Test Plan (milestone-level)

| Test | Tier | Step |
|---|---|---|
| TypedCache wrong type | unit | 19.0 |
| Manager sequential/queue/cancel/stream | unit | 19.2–19.4 |
| doMutation not stuck | unit | 19.6 |
| Concurrent mutation rejected | unit | 19.6 |
| No program.Send | static | 19.9 |
| Race detector | all | 19.10 |

---

## Definition of Done

- [ ] Steps 19.0–19.10 complete
- [ ] TaskManager per D19-1–D19-5
- [ ] All write paths through manager
- [ ] `program.Send` eliminated from handlers
- [ ] `isBusy` removed
- [ ] Unit tests + race pass
- [ ] DESIGN.md + AGENTS.md reflect final message flow
- [ ] status.md updated

---

## Post-Milestone Gate

- [ ] M20.3 batch upgrade may start
- [ ] M22.1 minimal CI recommended
- [ ] M21.1 teatest helper may start

---

## Rollback Plan

If TaskManager integration fails mid-milestone:

1. Keep 19.0 TypedCache (safe independent)
2. Revert 19.5+ Model wiring
3. Restore `isBusy` temporarily with 19.6 fix only (validate-before-set) as hotfix

Do not ship hotfix without 19.6 validation fix.
