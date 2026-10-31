# Receipt — reviewer lane, fix cycle for slice `detach-master-plan`

Verdict: ok

Re-derived the fix from the tree. Reviewed `17b1b8d`, `dce1468`, `26f44cf`, `88dbed9` (+ receipt `0bcff0b`). Gate run by me, not taken from the runner's claim. All four findings are genuinely fixed; no earlier verified property regressed. One minor documentation defect in the fix receipt is recorded below (F-05) — it affects no acceptance criterion.

## Findings

F-05 | minor | The fix receipt's F-04 evidence bullet states "The brief's literal command prints `1`". It does not — the brief's literal command is the very command quoted one line above, and it prints `0`: the pattern `^[+-]` cannot match a diff *context* line, which begins with a space. The receipt's recorded output (`0`) is right; its explanation of the count is wrong. The re-review brief repeats the same error ("prints `1` because the line appears as a diff context line"). The substantive F-04 conclusion (byte-identical) is correct and confirmed below by a stronger check. | `git diff 74554ad..HEAD -- AGENTS.md \| grep -c '^[+-].*Architectural context'` → `0` (not `1`); `git diff 74554ad..HEAD -- AGENTS.md \| grep -n 'Architectural context'` → `8: > **Architectural context:** …` (context line, leading space) | Correct the sentence in `work/inbox/runner-fix-detach-master-plan.md` (and the note in `work/briefs/review-fix-detach-master-plan.md`) to say the literal command prints `0` because `^[+-]` excludes context lines. No implementation change.

## Evidence (command → result)

- **R1** `grep -n -i -E 'raised late|\.mp-write\.lock|LB-06' work/questions.md` → hits at 19/26/29/39/41/47. `git diff 3431b9d..HEAD -- work/questions.md` → one hunk, **31 added lines, 0 removed**: appended, not rewritten; the "_None. No questions are open._" line (8) stays. Q-01 and Q-02 each carry **Raised by:** runner lane, **Owner:** the orchestrator, **Resolution:** resolved.
- **R2** `grep -n 'LB-06' work/backlog.md` → `| LB-06 | … | M13 | medium | closed-by-LB-01 |`.
- **R3** `grep -n 'LB-DROP-01' work/backlog.md` → `| LB-DROP-01 | Dead header fix in the \`master-plan-bkp\` directory | B-01 | n/a | dropped |` — archive parenthetical gone from row and reason; reason keeps "the target directory does not exist on this machine". `ls -d master-plan-bkp ../master-plan-bkp` → "No such file or directory" (still absent).
- **R4** `diff <(git show 74554ad:AGENTS.md \| sed -n '5p') <(git show HEAD:AGENTS.md \| sed -n '5p')` → prints nothing (exit 0); `od -c` of both → identical, trailing two spaces present. Full `git diff 74554ad..HEAD -- AGENTS.md` removes 9 lines, all planning-related (header "Planning context", the `master-plan/` map row, "Planning Rules" + its 6 items, the References entry) — no unrelated content touched.
- **R5** Per-commit scope: `17b1b8d`→`work/questions.md`; `dce1468`→`work/backlog.md`; `26f44cf`→`work/backlog.md`; `88dbed9`→`AGENTS.md`. `git diff --stat 8d6fa5d..88dbed9` → exactly those three files. The four commits are in F-01→F-04 order.
- **R6** `git ls-files master-plan | wc -l` → 0. `ls master-plan | wc -l` → 18 entries, 33 files on disk. Blob compare of all 33 against `3431b9d` → **no `CHANGED` lines**. `.gitignore` has `/master-plan/` (22) and no `mp-project-write.lock`.
- **R7** V1 `grep -n -iE '\bmp (status|milestone|…)' AGENTS.md` → no hits; `master-plan` hits only at 89/91 (frozen-archive sentences); deleted-rule grep → no hits. V2 `grep -n -iE 'master-plan|\bmp\b' README.md` → no hits. V3 `DESIGN.md` → broken link replaced, 2026-10-31 decision row at 120, no link into `master-plan/`; `ls master-plan/milestones/ \| grep -i 18` → no match. V4 all five files exist; `grep -c '^\- \[x\]' SMOKE-CHECKLIST.md` → **28**; `diff <(grep -v -i provenance SMOKE-CHECKLIST.md) master-plan/smoke-checklist.md` → identical (exit 0). V5 as R6. `git diff --name-only 74554ad..180d398 -- internal/ cmd/ testdata/ scripts/ Makefile .github/` → empty. Repo-wide `git grep -n -i -E 'master-plan|mp status|mp milestone|mp backlog' -- . ':(exclude)master-plan'` → every hit is the .gitignore entry, a frozen-archive sentence, the DESIGN decision row, the checklist provenance line, the backlog's dropped-row/provenance lines, or a brief/receipt. **No live `mp` instruction survives**, including in the appended questions section.
- **R8** Gate run by me: `make build` → exit 0; `make lint` (`go vet ./...`) → exit 0; `go test ./... -race -count=1` → exit 0, 9 packages `ok`, 2 with no test files; `make cover-check` → "All coverage floors met" (brew 64.3/62, presentation 95.4/91, gui 55.9/36, modal 41.4/41). All four pass, matching the runner's numbers.

## Mutations (all reverted; `git status` clean after each)

- **N1** LB-06 status → `active` → R2's grep for `closed-by-LB-01` returned **0** hits. Reverted.
- **N2** Deleted the appended section (`sed -i '' '18,$d'`) → R1's grep exit 1, no hits. Reverted.
- **N3** Stripped trailing spaces from `AGENTS.md` line 5 → R4's diff reported the difference (exit 1). Reverted. All three checks bite — none is vacuous.

## Questions

`work/questions.md` now records both. **Q-01** (`.mp-write.lock` contradiction) — **answered**: resolution verified, `git show 180d398 --name-status` shows `D master-plan/.mp-write.lock`, `SMOKE-CHECKLIST.md` exists on disk, `git ls-files master-plan` = 0, archive blobs unchanged. **Q-02** (LB-06 clause ambiguity) — **answered**: `closed-by-LB-01` is the status cell. Both name the runner lane as raiser and the orchestrator as owner, as the fix brief required. **No new question** is raised by this fix cycle, and none is left unanswered. The "_None. No questions are open._" line is consistent: both entries are resolved.

## Notes (not findings)

- **R5's literal range is over-wide.** `git diff --name-only 8d6fa5d..HEAD -- … work/briefs work/inbox …` prints two paths: `work/inbox/runner-fix-detach-master-plan.md` (commit `0bcff0b` — the receipt the fix brief mandates) and `work/briefs/review-fix-detach-master-plan.md` (commit `6ae6fb8` — the orchestrator's own re-review brief). Neither is the fix lane's doing; scoped to the four fix commits the range shows only the three allowed files. The fix receipt's own invariant claim ("only AGENTS.md, work/backlog.md, work/questions.md") was true when written (HEAD was `88dbed9`) and is stale only because two later commits landed. Not a finding against the fix.
- `work/STATE.md` still reads `implemented` / `awaiting reviewer`; that row is the orchestrator's to close.

Verdict: ok
