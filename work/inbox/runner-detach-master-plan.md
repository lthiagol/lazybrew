# Receipt — runner lane, slice `detach-master-plan`

## Commits (one per AC, in order)
| AC | commit | subject |
|---|---|---|
| AC-01 | `3a770a7` | docs: AGENTS.md stops driving work from mp |
| AC-02 | `b95af56` | docs: README drops the plan directory, points at work/ |
| AC-03 | `7428739` | docs: DESIGN.md records the mp/master-plan detach decision |
| AC-04 | `3431b9d` | docs: add tracked lane-flow structure and smoke checklist |
| AC-05 | `180d398` | chore: detach master-plan from git tracking, keep it locally frozen |

## Evidence per AC
- **AC-01** — `git show 3a770a7 --stat` → 1 file, 34+/11−: the `mp` Planning Rules section, the `master-plan/` map row, the `master-plan/AGENTS.md` reference and the "Planning context" line are gone; **Planning** + **Lane Protocol** and the `SMOKE-CHECKLIST.md` release-gate pointer are in; Bubble Tea/testing/git/verification rules unchanged in the same diff.
- **AC-02** — `git show b95af56 --stat` → 1 file; map line now `work/`; `grep -i -E 'master-plan|mp ' README.md` → no planning prose left.
- **AC-03** — `git show 7428739 -- DESIGN.md` → 2 lines: broken `master-plan/milestones/18-…` link → plain text + the required 2026-10-31 decision row. Link was broken: `ls master-plan/milestones/ | grep -i 18` → no match.
- **AC-04** — created `work/README.md`, `work/backlog.md`, `work/STATE.md`, `work/questions.md`, `SMOKE-CHECKLIST.md`; `grep -c '^\- \[x\]' SMOKE-CHECKLIST.md` → **28**; `diff <(grep -v Provenance SMOKE-CHECKLIST.md) master-plan/smoke-checklist.md` → identical apart from the provenance line (item text and `[x]` preserved).
- **AC-05** — `git rm -r --cached master-plan` (33 index removals in `180d398`); `.gitignore` has `/master-plan/`, no `.mp-project-write.lock`; `git ls-files master-plan | wc -l` → **0**; `ls master-plan | wc -l` → **18**; sha256 of all 33 on-disk files vs `3431b9d` → **0 changed**.

## Gate (all four, in order)
| Command | Result |
|---|---|
| `make build` | **PASS** (exit 0) |
| `make lint` (`go vet ./...`) | **PASS** (exit 0) |
| `go test ./... -race -count=1` | **PASS** (exit 0; 9 packages `ok`) |
| `make cover-check` | **PASS** (exit 0; brew 64.3/62, presentation 95.4/91, gui 55.9/36, modal 41.4/41) |

## For the next session
- No Go source/test/`Makefile`/workflow touched; `master-plan/` is still on disk, untracked and gitignored, with all 33 files unmodified. Remaining `master-plan` mentions: the `.gitignore` entry (with its "frozen local archive" comment), sentences calling it a frozen untracked local-only archive, and the brief itself (not editable) — no live `mp` instruction survives.
- `work/STATE.md` row updated this session (state `implemented`, verdict `awaiting reviewer`) per lane protocol rule 4; AC-04 created it as `implementing` / `—`.
- **Contradiction found:** AC-04 says the checklist is "tracked only inside `master-plan/`", but git also tracked a stale `master-plan/.mp-write.lock`. Handled as specified — checklist copied out to `SMOKE-CHECKLIST.md`, directory untracked with no on-disk file modified.
