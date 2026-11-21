# runner: cr4-gui-dedup fix r1

Commits, one per finding:
- fe72f96 cr4 fix F-r1-01: cancelled linesTask always reports ctx.Err()
- 187a8cb cr4 fix F-r1-02: keymap test holds each panel to its footer and help

## F-r1-01 — linesTask cancel hang

Red before the fix (tests written first; `go test ./internal/gui/ -run TestLinesTask`):
- TestLinesTaskCancelMidStream: `errCh delivered nothing within 2s`
- TestLinesTaskCancelledBeforeStart: `errCh delivered nothing within 2s`
- TestLinesTaskCancelThroughManager/doctor, /missing, /vulns: `no TaskCompletedMsg within 2s`

Green after: all TestLinesTask*, TestErrTask*, TestStreamTask* pass under -race;
TestLinesTaskCancelThroughManager ×5 under -race passes (300 runs per shape).

Exit paths in task_build.go; each now sends exactly one value on errCh:
- errTask: run's error sent once on a buffered channel, synchronously.
- linesTask: fetch error sent once. Otherwise sendLines returns ctx.Err(), which is
  nil only when every line went out and the task was not cancelled. Sent once; out closed by defer.
- streamTask: passes brew's errCh through. Every brew stream delivers one value or
  closes errCh, so readChunk cannot block: DefaultRunner.ExecuteStream (defer close,
  sendErr once), sudoRunner, LoggingRunner and streamWrite (one receive, one send, close).
  A nil errCh is skipped by readChunk.

## F-r1-02 — keymap test per panel

The test now reads each panel's dispatch from the msg.String() cases at the top of
handleKey (m.activePanel ==/!= decides the panels), each panel's footer from
footerHints on a real model over its states, and each panel's help from Global
plus its own page. Rule: every key a panel answers is covered by a table entry on
its footer (if it has a footer word) and on its help (if it has a help line); every
footer and help line names a key the panel answers.

Mutations, each run then restored (git status clean afterwards):
1. drop stDoctor from the Status footer:
   `keymap_test.go:355: Status: the dispatch answers "d", which its footer and help do not both show`
2. drop fFetch from the Formulae footer:
   `keymap_test.go:355: Formulae: the dispatch answers "F", which its footer and help do not both show`
3. drop the n hint from ConfirmModal.Hints() (TestModalHintsMatchUpdate):
   `keymap_test.go:490: ConfirmModal: Update answers "n", which Hints does not list`
4. drop stDoctor from the Status help page:
   `keymap_test.go:355: Status: the dispatch answers "d", which its footer and help do not both show`

Production change: modal.Hint gains Keys (the msg spellings the modal answers).
Footer and help frames are unchanged. Modal exemptions: confirm enter/h/l/left/right
(shown as the Yes/No buttons), menu digits (drawn on each row). Stale exemptions fail.

## Gate

- `make ci`: exit 0 (gofmt, vet, go test ./... -race -count=1, every coverage floor
  OK, build). Output: /tmp/claude-1000/-home-agent-coding-lazybrew/12a5785a-8597-4303-a72d-bba11610c761/scratchpad/ci.txt
- `GOOS=darwin go build ./...`: exit 0.
- `make test-fast PKG=./internal/gui/...`: pass after each commit.
- deadcode and staticcheck: not run, as instructed (not installed; fetch denied).

## Next session must know

- deadcode and staticcheck are still unrun for this slice.
- Esc on panel footers appears only when an operation is done or help is open; the
  footer check counts those states, so no footer frame changed.
- Seen, not done: none new.
