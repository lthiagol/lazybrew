# Fix round r1 — slice `v2-01-charm-v2`

> The slice contract `work/briefs/v2-01-charm-v2.md` still applies in full
> (invariants, gate, no visual change). This brief lists the findings from
> `work/inbox/reviewer-v2-01-charm-v2-r1.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r1-01 | major | Both `setup-go` steps (`.github/workflows/wip-ci.yml`, `.github/workflows/stable-ci.yml`) use `go-version-file: go.mod`; no pinned `go-version` remains; every other step and `cache: true` unchanged. |
| F-r1-02 | minor | Unit tests in `internal/gui` proving a `tea.PasteMsg{Content: ...}` (a) is inserted into the focused search input, (b) on the Search panel with the input unfocused focuses it and inserts the text, (c) reaches an open `InputModal`'s text field. Assert on the resulting value, not just "no panic". |

## Rules

- One commit per finding, in order, subject naming the id, e.g.
  `fix(ci): build with the Go version go.mod declares (F-r1-01)`. Stage by name.
- If a test in F-r1-02 reveals the paste path is actually broken, fix the code
  in the same commit and say so in the receipt.
- Gate: `make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`,
  `make cover-check` — all pass, real output in the receipt.
- AC-03 frame equivalence must still hold if you touch any production file:
  re-run the harness and record the empty diff in the receipt.
- Receipt `work/inbox/runner-v2-01-charm-v2-fix-r1.md` (≤ 30 lines), committed
  **once** — get it under the limit before committing. Worktree clean. Do not push.
