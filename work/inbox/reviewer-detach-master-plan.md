# Receipt — reviewer lane, slice `detach-master-plan`

Verdict: changes-needed

Reviewed `3a770a7`, `b95af56`, `7428739`, `3431b9d`, `180d398` (+ receipt `2dff233`). Gate run by me, not taken from the runner's claim.

## Findings

F-01 | major | The runner found two defects in the brief but recorded neither in `work/questions.md`, which still asserts "_None. No questions are open._" The contract's Questions section and AGENTS.md rule 5 require a brief-wrong or ambiguous item to be written there. (a) AC-04 says the checklist is "tracked only inside `master-plan/`", but git also tracked `master-plan/.mp-write.lock`. (b) AC-04 is self-contradictory for LB-06: it must be "closed-by-verification with LB-01" yet "every row's status starts as active except the dropped row".
| `cat work/questions.md` → "_None. No questions are open._"; `git show 180d398 --name-status` → `D master-plan/.mp-write.lock`; `grep -n 'LB-06' work/backlog.md` → status `active` | Append both to `work/questions.md` (append, never rewrite); the orchestrator owns the answers.

F-02 | minor | LB-06's `status` cell reads `active`; the reviewer brief expects a closed value (`closed-by-LB-01`). The closure is stated in the `work` column instead, so the status column alone hides that the M13 review is closed.
| `grep -n 'LB-06' work/backlog.md` → `| LB-06 | … | M13 | medium | active |` | Set LB-06 status to `closed-by-LB-01` (the more specific clause wins).

F-03 | minor | LB-DROP-01 calls the non-existent `master-plan-bkp` directory "part of the frozen, untracked, local-only archive". It exists nowhere on this machine, so it cannot be part of the archive. The required reason ("the target directory does not exist on this machine") is present and correct.
| `ls -d master-plan-bkp ../master-plan-bkp` → "No such file or directory" | Drop the archive parenthetical from the LB-DROP-01 row and its reason paragraph.

F-04 | minor | AC-01.7 asked for unrelated content to stay byte-for-byte. The trailing double-space markdown break was stripped from the untouched "Architectural context" header line. Rendering is identical (it is now the last blockquote line), so this is cosmetic.
| `git diff 74554ad..3a770a7 -- AGENTS.md` → `-… (DESIGN.md).  ` / `+… (DESIGN.md).` | Restore the two trailing spaces, or accept and note it.

## Evidence (command → result)

- **V1** `grep -n -iE '\bmp (status|milestone|backlog|decisions|…)\b' AGENTS.md` → no hits (exit 1). `grep -n -i master-plan AGENTS.md` → only lines 89/91, both inside the frozen-untracked-archive sentence. `git show 3a770a7 -- AGENTS.md | grep '^-' | grep -iE 'program.Send|task.Manager|zero goroutine|teatest|cover-check|Do not commit unless'` → no hits: no unrelated engineering rule deleted.
- **V2** `grep -n -iE 'master-plan|\bmp\b' README.md` → no hits (exit 1). `git show b95af56 --stat` → `README.md | 2 +-`, 1 file only; map row now `work/  Lane flow: briefs, receipts (inbox/), state, backlog, questions`.
- **V3** `git show 7428739 -- DESIGN.md` → exactly 2 changes: broken link → plain text, plus the 2026-10-31 decision row (matches AC-03 verbatim). `grep -n master-plan DESIGN.md` → only that row, no link. `ls master-plan/milestones/ | grep -i 18` → no match, so the old link was indeed broken.
- **V4** All five files exist. `grep -c '^\- \[x\]' SMOKE-CHECKLIST.md` → 28. `diff <(grep -v -i provenance SMOKE-CHECKLIST.md) master-plan/smoke-checklist.md` → identical (exit 0); item text and `[x]` states preserved. Backlog carries LB-01…LB-06 + LB-DROP-01 with the required sources/priorities (see F-02/F-03). `work/STATE.md` row: state `implemented` (not `implementing`), verdict `awaiting reviewer`, commits column names all five ids.
- **V5** `git ls-files master-plan | wc -l` → 0. `.gitignore` has `/master-plan/` (line 22) and no `mp-project-write.lock`. `git show 180d398 --name-status` → 33 `D master-plan/…` plus `.gitignore`. `ls master-plan | wc -l` → 18 entries, 33 files on disk, none missing. Blob comparison of all 33 on-disk files against `3431b9d` → **no `CHANGED` lines**: the archive is byte-identical and untouched.
- **V6** `git diff --stat 74554ad..180d398` → only `.gitignore`, `AGENTS.md`, `DESIGN.md`, `README.md`, `SMOKE-CHECKLIST.md`, `work/*`, and `master-plan/` removals. `git diff --name-only 74554ad..180d398 -- internal/ cmd/ testdata/ scripts/ Makefile .github/` → empty.
- **V7** `git show 86d34c1 --stat` → the mouse-support commit (`internal/gui/*`), untouched. `git diff 74554ad..180d398 --stat -- internal/ cmd/` → empty: neither modified nor reverted.
- **V8** Gate run by me: `make build` → exit 0; `make lint` (`go vet ./...`) → exit 0; `go test ./... -race -count=1` → exit 0, 9 packages `ok`; `make cover-check` → "All coverage floors met" (brew 64.3/62, presentation 95.4/91, gui 55.9/36, modal 41.4/41). All four pass, matching the runner's claim.
- **Repo-wide** `git grep -n -i -E 'master-plan|mp status|mp milestone|mp backlog' -- . ':(exclude)master-plan'` → every hit is the frozen-archive sentence, the `.gitignore` entry, the DESIGN decision row, the checklist provenance line, or the briefs/receipts. No live `mp` instruction survives.

## Mutations (all reverted; `git status` clean after each)

- **M1** Appended `Source of truth: run mp status` to `AGENTS.md` → V1's grep fired (hit at line 197). Reverted.
- **M2** Removed `/master-plan/` from `.gitignore`, `git add -f master-plan/config.json` → `git ls-files master-plan | wc -l` printed **1** (was 0). Reverted via `git rm --cached` + `.gitignore` restore; count back to 0.
- **M3** Deleted one `- [x]` item from `SMOKE-CHECKLIST.md` → count dropped to **27** (was 28). Reverted; count back to 28. All three checks bite — none is vacuous.

## Questions

`work/questions.md` claims none are open. That is inaccurate: the `.mp-write.lock` contradiction was reported only in the runner's receipt, and LB-06's contract ambiguity was never raised (F-01). Two questions are **missed**.

`work/STATE.md` still reads `implemented` / `awaiting reviewer`; updating it to the final verdict is the orchestrator's row, and this lane commits only its receipt.

Verdict: changes-needed
