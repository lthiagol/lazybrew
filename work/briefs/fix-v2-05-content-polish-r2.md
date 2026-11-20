# Fix round r2 — slice `v2-05-content-polish`

> The slice contract and the r1 fix brief still apply. Findings from
> `work/inbox/reviewer-v2-05-content-polish-r2.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r2-01 | major | Upgrade-all progress handles everything plain `brew upgrade` prints: a formulae `Upgrading N outdated package(s):` header, then a separate casks header, and `Upgrading N dependent(s) of upgraded formula(e):` followed by per-dependent `Upgrading x` steps. **Total = sum of every `outdated package(s):` header's N + every dependents header's N**; headers are never package steps. Read the strings from Homebrew's own source on this machine (`$(brew --repository)/Library/Homebrew/cmd/upgrade.rb`, `.../cask/upgrade.rb` — read only) and build test transcripts from them: (a) 2 formulae then 1 cask, mid-run on the cask → `2/3`, end → `3/3`; (b) 1 formula + 1 dependent, mid-run on the dependent → `1/2`, end → `2/2`; (c) the r1 cases still pass. The bar is never full while a step is still running. |
| F-r2-02 | minor | At least one dashboard test asserts the singular `1 package` as a whole word (regexp `1 package\b` or the full row), and the receipt shows the reviewer's mutation M4 (`package` → `packages` for 1) now fails it. |
| F-r2-03 | minor | `result.md` lines 17 and 42 describe the current card order and the 80×24 frame. |

## Rules
- One commit per finding, subject naming the id. Stage by name. Gate (CGO_ENABLED=1 for -race).
- Receipt `work/inbox/runner-v2-05-content-polish-fix-r2.md` (≤ 25 lines) — check the line count **before** committing; commit it **exactly once**. Worktree clean. No push. Never run brew write commands.
