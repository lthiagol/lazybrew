# lazybrew — agent guide

A Bubble Tea v2 TUI (`charm.land/bubbletea/v2`) for Homebrew (Go). Architecture: [DESIGN.md](DESIGN.md).
Keep this file short; the details live in the files under **Read when needed**.

## Commands

```bash
make build      # bin/lazybrew
make test-fast  # unit tests, no -race; PKG=./internal/gui/... narrows it
make fmt        # gofmt
make ci         # the gate, five steps: gofmt check, vet, go test -race, coverage floors, build
```

`make ci` passing is "done". When /tmp is tight use `GOFLAGS=-p=1 make ci`; never
delete files in /tmp you did not create (another project shares it).

## Map

| Path | What |
|---|---|
| `cmd/lazybrew/` | `main.go`, flags |
| `internal/app/` | bootstrap: config, runner, `gui.New` |
| `internal/brew/` | brew runner, services, cache, parsers; `brewtest/` = fake runner for tests |
| `internal/config/` | YAML config, `Sample()` (`--config-sample`) |
| `internal/gui/` | Bubble Tea model, panels, keys, render; `flows/` = teatest E2E, `testutil/` = their model helpers |
| `internal/gui/{modal,presentation,style,task,anim,textutil}` | modals, formatters, themes/icons, TaskManager, motion, text helpers |
| `internal/logtail/` | bounded read of a log file's tail |
| `work/` | plan: briefs, `STATE.md` (resume record), `backlog.md`, `questions.md` |

## Hard rules

1. **This machine runs a real Homebrew with live services (`moshi-hook`).** Never run a
   brew write: install, uninstall, reinstall, upgrade, update, pin/unpin, tap/untap,
   trust/untrust, cleanup, autoremove, `services start|stop|restart|run|cleanup|kill`.
   Never run the TUI over the real brew: not `make run`, and not a test, harness or frame
   dump that builds `gui.New` over `brew.NewDefaultRunner()` (its `Init` runs `brew update`).
   Tests and harnesses use `brewtest.NewMockRunner`. Read-only brew commands are fine.
2. **Branches.** Work on `wip`. Never commit, merge or push to `stable` unless the
   user asks for it in this session.
3. **Git.** Commit only when asked or when running the lane/crew flow. Stage by
   name, never `git add -A` / `git add .`. Never push, force-push, rebase or amend
   unless asked, or as [work/README.md](work/README.md) authorises for a closed slice
   on `wip`. Commits are `Luiz Thiago <lthiagol@gmail.com>` (repo config).
   Conventional, imperative subjects.
4. **Bubble Tea.** State changes only in `Update`; no `program.Send`; no raw
   goroutines in handlers; every brew write is a task started with `m.enqueue`.
   Call a mutating method before the return (`cmd := m.f(); return m, cmd`), never
   `return m, m.f()`. Details: [docs/agents/gui.md](docs/agents/gui.md).
5. **Tests ship with the change**, and E2E flows assert on `View()`. Never lower a
   coverage floor without a ruling in `work/questions.md`.
6. **Scope.** One slice per change set; no drive-by refactors.
7. **`master-plan/` is a frozen, untracked archive.** Do not read it, edit it, or
   run `mp`.

## Read when needed

| When you are… | Read |
|---|---|
| writing or fixing tests, chasing a flake | [docs/agents/testing.md](docs/agents/testing.md) |
| adding, parsing or caching a brew command (`internal/brew/`) | [docs/agents/brew.md](docs/agents/brew.md) |
| changing anything under `internal/gui/` | [docs/agents/gui.md](docs/agents/gui.md) |
| tagging, releasing, bumping the tap | [docs/agents/release.md](docs/agents/release.md) |
| running a planned slice (brief → build → review → STATE row) | [work/README.md](work/README.md) |
| resuming work | `work/STATE.md` ("▶ Resume here" at the top; slice rows below); the "Open questions" section of `work/questions.md`; grep a Q-id or LB-id when one is cited |
