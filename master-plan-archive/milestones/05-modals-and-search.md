# Milestone 5 — Modals, Input & Search

> **Status:** ✅ Complete  
> **Depends on:** Milestone 2 (TUI Shell), Milestone 3 (Brew Data Layer)  
> **Enables:** Milestone 6 (Package Mutations), Milestone 7 (Taps & Trust)

---

## Goal

Build the modal system (confirmations, text input, menus, progress indicators) and the search flow. These are the interactive building blocks that all mutation operations will use. After this milestone, you can search for packages and see results — but not yet install them.

---

## Why Modals Come Before Mutations

Every mutation (install, uninstall, upgrade, tap, trust) needs at least one modal:
- **Install** needs search input → result selection
- **Uninstall** needs confirmation
- **Upgrade** needs progress display
- **Tap** needs URL input
- **Trust** needs a menu of options

By building and testing the modal system first, all mutation milestones are pure business logic — they just assemble the modals.

---

## Steps

### 5.1 — Modal Framework

**What:** A modal overlay system that captures input and renders on top of the main UI.

**File:** `internal/gui/modal/modal.go`

**Architecture:**
```go
// ModalType determines the kind of modal
type ModalType int
const (
    ModalConfirm ModalType = iota
    ModalInput
    ModalMenu
    ModalProgress
    ModalError
)

// Modal is the interface all modals implement
type Modal interface {
    // Bubble Tea
    Init() tea.Cmd
    Update(msg tea.Msg) (Modal, tea.Cmd)
    View() string
    
    // Modal lifecycle
    IsActive() bool
    IsDone() bool         // true when modal has a result or was cancelled
    WasCancelled() bool   // true if user pressed Esc
}
```

> **Design decision — Type-safe results:** Instead of a generic `Result() interface{}`, each modal type returns its own typed result via a type-specific method. The root model uses a `ModalResult` message type that carries the result:
> ```go
> type ModalResult struct {
>     Confirm  *bool     // nil if not a confirm modal
>     Input    *string   // nil if not an input modal
>     MenuIdx  *int      // nil if not a menu modal
> }
> ```
> This avoids type assertions and makes the code self-documenting.

**Root model integration:**
- When a modal is active, the root model routes all input to the modal
- `Esc` closes the modal (cancel)
- Modal renders as a centered overlay on top of the existing UI
- Background is dimmed (Lip Gloss opacity trick: render BG in darker style)

**Visual:**
```
┌───────────────────────────────────────────────┐
│  (dimmed background)                          │
│                                               │
│     ┌── Confirm ──────────────────────┐       │
│     │                                 │       │
│     │  Uninstall ripgrep 14.1.1?      │       │
│     │                                 │       │
│     │  This cannot be undone.         │       │
│     │                                 │       │
│     │     [Yes]      [No]             │       │
│     │                                 │       │
│     └─────────────────────────────────┘       │
│                                               │
└───────────────────────────────────────────────┘
```

**Acceptance criteria:**
- [ ] Modal renders centered over background
- [ ] Background is visually dimmed
- [ ] All input routes to modal when active
- [ ] `Esc` cancels the modal
- [ ] Modal result returned to the caller
- [ ] Multiple modal types share the framework
- [ ] Works at various terminal sizes

---

### 5.2 — Confirmation Modal

**What:** "Yes/No" confirmation dialog for destructive actions.

**File:** `internal/gui/modal/confirm.go`

**Interface:**
```go
type ConfirmModal struct {
    title    string
    message  string
    selected bool   // true = Yes, false = No
    done     bool
    result   bool   // final answer
}

func NewConfirmModal(title, message string) *ConfirmModal
```

**Keybindings:**
| Key | Action |
|---|---|
| `y` / `Enter` (when Yes selected) | Confirm |
| `n` / `Enter` (when No selected) | Cancel |
| `h` / `l` or `←` / `→` | Toggle Yes/No |
| `Esc` | Cancel |

**Acceptance criteria:**
- [ ] Shows title and message
- [ ] Yes/No buttons with highlight on selected
- [ ] `y` is a shortcut for confirm
- [ ] `n` is a shortcut for cancel
- [ ] Returns boolean result
- [ ] Defaults to "No" (safe default for destructive actions)

---

### 5.3 — Text Input Modal

**What:** Single-line text input for search queries, tap URLs, etc.

**File:** `internal/gui/modal/input.go`

**Implementation:**
- Uses `bubbles/textinput` component internally
- Shows a prompt label (e.g., "Search:", "Tap URL:")
- Supports placeholder text
- Optional validation function

**Interface:**
```go
type InputModal struct {
    prompt     string
    input      textinput.Model
    done       bool
    cancelled  bool
    validator  func(string) error
}

func NewInputModal(prompt string, opts ...InputOption) *InputModal
```

**Options:**
```go
type InputOption func(*InputModal)
func WithPlaceholder(s string) InputOption
func WithValidator(fn func(string) error) InputOption
func WithInitialValue(s string) InputOption
```

**Keybindings:**
| Key | Action |
|---|---|
| `Enter` | Submit |
| `Esc` | Cancel |
| Standard text editing | Type, backspace, cursor movement |

**Acceptance criteria:**
- [ ] Text input renders with prompt
- [ ] Placeholder text shown when empty
- [ ] `Enter` submits, `Esc` cancels
- [ ] Validation error shown inline (if validator fails)
- [ ] Returns string result or cancellation

---

### 5.4 — Menu Modal

**What:** A scrollable list of options for action selection (e.g., trust config).

**File:** `internal/gui/modal/menu.go`

**Implementation:**
- Shows a titled list of options
- Uses `j/k` or `↑/↓` to navigate
- `Enter` selects
- Optional: icon/badge per option

**Interface:**
```go
type MenuItem struct {
    Label string
    Icon  string    // optional
    Key   string    // shortcut key (e.g., "1", "2")
}

type MenuModal struct {
    title    string
    items    []MenuItem
    selected int
    done     bool
}

func NewMenuModal(title string, items []MenuItem) *MenuModal
```

**Acceptance criteria:**
- [ ] Options listed with highlight on selected
- [ ] `j/k` navigation
- [ ] `Enter` selects
- [ ] Shortcut keys work (e.g., pressing `1` selects first option)
- [ ] `Esc` cancels
- [ ] Returns selected index or cancellation

---

### 5.5 — Progress Modal

**What:** A streaming output display for long-running operations.

**File:** `internal/gui/modal/progress.go`

**Implementation:**
- Shows a title and a scrollable output area
- New lines appear at the bottom (auto-scroll)
- A spinner indicates the operation is still running
- When complete, shows success/failure status
- Optional "Cancel" action (sends SIGINT to the underlying process)

**Interface:**
```go
type ProgressModal struct {
    title    string
    lines    []string
    spinner  spinner.Model
    done     bool
    err      error
    viewport viewport.Model
}

func NewProgressModal(title string) *ProgressModal
func (m *ProgressModal) AppendLine(line string)
func (m *ProgressModal) SetDone(err error)
```

**Visual:**
```
┌── Installing ripgrep ──────────────────────┐
│                                            │
│  ==> Downloading ripgrep-14.1.2.tar.gz     │
│  ==> Pouring ripgrep--14.1.2.arm64.bottle  │
│  ==> Summary                               │
│  /opt/homebrew/Cellar/ripgrep/14.1.2       │
│  ⣾ Installing...                           │
│                                            │
│  Press Esc to cancel                       │
└────────────────────────────────────────────┘
```

**After completion:**
```
┌── Installing ripgrep ──────────────────────┐
│                                            │
│  ==> Downloading ripgrep-14.1.2.tar.gz     │
│  ==> Pouring ripgrep--14.1.2.arm64.bottle  │
│  ==> Summary                               │
│  /opt/homebrew/Cellar/ripgrep/14.1.2       │
│                                            │
│  ✓ Installed successfully                  │
│                                            │
│  Press Enter or Esc to close               │
└────────────────────────────────────────────┘
```

**Acceptance criteria:**
- [ ] Spinner animates during operation
- [ ] Output lines stream in real-time
- [ ] Auto-scrolls to bottom
- [ ] Scrollable (PgUp/PgDn) for reviewing past output
- [ ] Shows success with green checkmark on completion
- [ ] Shows error with red X on failure
- [ ] `Esc` during operation cancels: sends SIGINT to the underlying brew process via context cancellation (`exec.CommandContext`), then waits up to 5s for graceful shutdown before SIGKILL
- [ ] `Esc` or `Enter` after completion closes the modal
- [ ] Cancel shows a "Cancelled" status (not "Error") in the modal

> **Cancel mechanism detail:** The progress modal holds a `context.CancelFunc`. When `Esc` is pressed during an active operation, it calls `cancel()`, which propagates to `exec.CommandContext` and sends SIGINT to the brew subprocess. The task manager (M6) handles the SIGKILL timeout. The modal transitions to a "Cancelled" state (amber, not red) and shows the output captured so far.

---

### 5.6 — Error Toast / Notification

**What:** A non-modal notification that appears briefly at the bottom.

**File:** `internal/gui/toast.go`

**Implementation:**
- Shows a styled message at the bottom of the screen (above the hint bar)
- Auto-dismisses after 3 seconds
- Types: success (green), error (red), info (blue), warning (amber)
- Queue multiple toasts (show one at a time)

**Interface:**
```go
type Toast struct {
    message string
    style   ToastStyle
    timer   time.Time
}

type ToastStyle int
const (
    ToastSuccess ToastStyle = iota
    ToastError
    ToastInfo
    ToastWarning
)
```

**Acceptance criteria:**
- [ ] Toast renders above the bottom bar
- [ ] Auto-dismisses after timeout
- [ ] Color-coded by type
- [ ] Multiple toasts queued

---

### 5.7 — Search Flow

**What:** End-to-end search: input → brew search → results in panel → select → view info.

**Implementation flow:**
1. User presses `/` (global) → InputModal with prompt "Search:"
2. User types query, presses Enter
3. lazybrew calls `brew search <query>`
4. Results populate the Search panel
5. Sidebar auto-switches to Search panel
6. User navigates results with `j/k`
7. Selecting a result shows its info in the main panel (calls `brew info` on demand)
8. `i` on a result will install it (wired in Milestone 6)

**Search panel list format:**
```
  Search: "neovim"  (5 results)
  ─────────────────────────────────
  ✓ neovim           📦  Vim-fork focused on extensibility
    neovim-qt        📦  Neovim client library and GUI
    neovim-remote    📦  Control nvim processes
    page             📦  Use neovim as pager
    neovide          🖥   Neovim client in Rust
```

- `✓` = already installed
- `📦` = formula, `🖥` = cask

**Acceptance criteria:**
- [ ] `/` opens search input modal from any panel
- [ ] Query submitted → brew search called (uses `--json=v2`)
- [ ] Results shown in search panel
- [ ] Panel auto-focuses to Search
- [ ] Already-installed packages marked
- [ ] Selecting a result shows info in main panel
- [ ] Empty search results: "No results for 'xyz'"
- [ ] Loading spinner during search
- [ ] Pressing `/` again while in Search panel pre-fills the previous query for editing
- [ ] `Esc` in Search panel clears results and returns to previous panel

> **Search UX notes:**
> - No incremental search for v1 — the user types a full query and presses Enter. `brew search` is fast enough that the delay is acceptable.
> - Search history is not persisted across sessions for v1. The "re-invoke `/` to edit previous query" covers the most common case.
> - Results are sorted by: installed first, then alphabetical. This matches user expectations (installed packages are more relevant).

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestConfirmModal_Yes` | Unit | `internal/gui/modal/confirm_test.go` | `y` key returns true |
| `TestConfirmModal_No` | Unit | `internal/gui/modal/confirm_test.go` | `n` key returns false |
| `TestConfirmModal_Esc` | Unit | `internal/gui/modal/confirm_test.go` | Escape cancels |
| `TestConfirmModal_DefaultNo` | Unit | `internal/gui/modal/confirm_test.go` | Default selection is No |
| `TestInputModal_Submit` | Unit | `internal/gui/modal/input_test.go` | Enter returns input text |
| `TestInputModal_Cancel` | Unit | `internal/gui/modal/input_test.go` | Esc returns empty + cancelled |
| `TestInputModal_Validation` | Unit | `internal/gui/modal/input_test.go` | Invalid input shows error, blocks submit |
| `TestMenuModal_Selection` | Unit | `internal/gui/modal/menu_test.go` | Enter returns selected index |
| `TestMenuModal_Shortcuts` | Unit | `internal/gui/modal/menu_test.go` | Number keys select option |
| `TestProgressModal_Stream` | Unit | `internal/gui/modal/progress_test.go` | Lines append and auto-scroll |
| `TestProgressModal_Complete` | Unit | `internal/gui/modal/progress_test.go` | Done state shows success/error |
| `TestToastDismissal` | Unit | `internal/gui/toast_test.go` | Toast disappears after timeout |
| `TestSearchFlow` | E2E (teatest) | `internal/gui/search_flow_test.go` | Full `/` → type → enter → results flow |
| `TestSearchEmpty` | E2E (teatest) | `internal/gui/search_flow_test.go` | Empty results message shown |
| `TestModalCapture` | E2E (teatest) | `internal/gui/gui_test.go` | Input routed to modal, not background |
| `TestModalEscClose` | E2E (teatest) | `internal/gui/gui_test.go` | Esc closes any modal |

---

## Definition of Done

- [ ] All 5 modal types implemented and tested
- [ ] Toast notification system working
- [ ] Search flow works end-to-end (input → results → info)
- [ ] Modals capture input (background doesn't receive keys)
- [ ] Modals render centered with dimmed background
- [ ] Esc closes all modals consistently
- [ ] All tests pass
