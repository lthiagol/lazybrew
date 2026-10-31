# Receipt — reviewer: lb01-m13-deps-fix

Verdict: ok

Gate run by me, not taken from the runner's receipt. Every mutation the record quotes reproduces byte-for-byte, and the Conclusion claims only what is true. Two minor defects remain in orchestrator-owned briefs (F-06, F-07); neither blocks `LB-01`'s closure.

## Evidence (command → result)

**S1 ✔** Presence section intact; all five presence commands re-run and every recorded output reproduces byte-for-byte (`commands.go` 18/22/833, `gui.go:455-456`, `commands.go` 779/788/794 + `:810`, `formulae.go:318` + field `:64`, `gui.go:627-634`). The mutation section carries, per invariant I-1..I-5, the exact diff, the command and the verbatim failing output. `git show b7dec88 --stat` → that file only.

**S2 ✔** The old "the four regression tests that cover it pass" sentence is gone; the new Conclusion claims only what holds (fix present in all five parts; each invariant covered by a test that fails when broken; gate green). Its table names nine tests and all nine failed under their invariant's mutation — no overclaim. The record's related claim also holds: under the I-1 and under the I-2 mutation the four contract-named tests stayed `ok`, exit 0.

**S3 ✔** `git show 544e0ed -- work/questions.md` → append-only (0 removed lines); Q-03 names the raiser, owner, observation and the corrected mapping; "_None. No questions are open._" still at line 8. `git show fa32989 -- work/briefs/review-lb01-m13-deps-verify.md` → the I-2 row now names the two `DataLoaded` tests, the I-5 row names `TestDepsSelectionChangeUpdatesContent`, and the "> **Correction (after the review).**" note is present.

**S4 ✔** `grep -n -A6 'Next slice' work/STATE.md` → names `LB-02` and no longer claims `LB-01` is next; the `lb01-m13-deps-verify` row is byte-identical in `fa32989`.

**S5 ✔** `git diff --name-only 923928c..HEAD` → only `work/` files; the same diff restricted to `internal/ cmd/ testdata/ scripts/ Makefile .github/ go.mod go.sum` printed nothing.

**S6 ✔** Gate run by me: `make build` → 0; `make lint` → 0; `go test ./... -race -count=1` → 0, every package `ok`; `make cover-check` → "All coverage floors met" (brew 64.3/62, presentation 95.4/91, gui 55.9/36, modal 41.4/41).

## Mutations I ran, and reverted (clean `git status` after each)

| inv | mutation | failure I saw | matches record |
|---|---|---|---|
| I-2 | `gui.go:455-456` re-kick deleted | `gui_test.go:1187: expected Deps tabContent populated after DataLoadedMsg, got cmd=nil content="" ok=false` + `gui_test.go:1220: expected loadTabContent cmd re-issued for Used By (no fast path) after DataLoadedMsg` | exact |
| I-5 | `gui.go:625-630` fetch cmd dropped | `gui_test.go:1403: expected zlib deps after j, ok=false content=""` | exact |
| I-3 | `commands.go:795` fast path never applies | `gui_test.go:1329` / `gui_test.go:1433` / `gui_test.go:1489: Deps body missing pcre2` | exact |
| I-1 | `commands.go:833` deadline removed | `gui_test.go:1310: fetchTabContentCmd hung past fetchTabContentTimeout (infinite-Loading reproducer)` | exact |
| I-4 | `formulae.go:318` build deps from `dependencies` | `formulae_test.go:155/169/195/199` | exact |

## Findings

F-06 | minor | The re-review brief miscounts: S2 enumerates nine test names but calls them "eight" (the fix brief's F-01 prose says "six"). The record's table names all nine correctly. | `awk '/^  \| I-1 \|/,/^  \| I-5 \|/' work/verification/lb01-m13-deps.md \| grep -o '`Test[A-Za-z]*`' \| sort -u \| wc -l` → 9 | correct the count in both briefs.

F-07 | minor | Q-03's closing sentence is stale: it calls the review brief's rows "stale … left untouched by the fix lane", but the orchestrator corrected them in `fa32989` minutes later. | `grep -n 'stale rows' work/questions.md` | append (never rewrite) one line to `work/questions.md` recording the `fa32989` correction.

## Questions

Q-01, Q-02 answered/resolved (pre-existing). Q-03 recorded and resolved; only its closing sentence is stale (F-07). Nothing newly raised by this cycle.
