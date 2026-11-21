# Fix brief — lb25-search-polish, round r2 (orchestrator-verified)

Ruling `Q-lb25-ok-uses`. Set `style.Icons.OK` to `✔` in both icon sets (Nerd too).
Tests first: a frame assertion per place (dashboard "ready to brew", command-log
success, success toast, trusted tap row) showing `✔`, none showing `✓`; grep that
no `✓` remains in production code except where a test or comment names it. Name
each golden that changes. One commit, subject ends `[Q-lb25-ok-uses]`. Gate:
`make test-fast PKG=./internal/gui/...`, then `make ci`. Append "Fix r2" (≤ 4
lines) to `work/inbox/runner-lb25-search-polish.md`, commit, do not edit
afterwards; reply with one done line. Never launch the TUI.
