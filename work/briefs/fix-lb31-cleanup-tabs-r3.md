# Fix brief — lb31-cleanup-tabs, round r3 (orchestrator-verified)

Ruling `Q-lb31-cleanup-path-cut`. One commit, tests first, subject ends `[LB-29] F-r1-01`.
When the last segment of a path is wider than the Path column, show that segment cut in the
middle (start … end, so a download-cache name keeps its hash and formula start and its
`.bottle.tar.gz` end) and drop the directories; paths whose last segment fits keep today's
cut. Test at 80x24 content width: the two `~/Library/Caches/Homebrew/downloads/<hash>--<name>…
.arm64_sonoma.bottle.tar.gz` files render differently and each row keeps its identifying start;
`llvm`/`python` rows unchanged from fix r2. Red on HEAD. Gate: `GOFLAGS=-p=1 make ci`. Never launch
the TUI. Append "Fix r3" (≤ 5 lines) to `work/inbox/runner-lb31-cleanup-tabs.md`, commit, do not edit
afterwards; reply with one done line.
