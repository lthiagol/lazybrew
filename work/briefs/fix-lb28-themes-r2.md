# Fix brief — lb28-themes, round r2 (low findings, orchestrator-verified)

Findings: `work/inbox/reviewer-lb28-themes-r1.md`. One commit, subject ends `[LB-25]`,
no behaviour change.

- F-r1-01: the Dracula comments (`theme.go` ~:122, `theme_test.go` ~:450) call
  `#44475a` Selection (the spec's name), not "current line". Value unchanged.
- F-r1-02: the README sample block matches `--config-sample` output line for line
  (regenerate it from the binary, do not hand-edit), or stop calling it the sample.
- F-r1-03: the `style.themes` comment states the two name lists (`config.themeNames`
  and `style.themes`) and that `TestRegistryMatchesConfig` ties them.

Gate: `GOFLAGS=-p=1 make test-fast PKG=./internal/gui/style/... ./internal/config/...`,
then `GOFLAGS=-p=1 make ci`. Never launch the TUI. Append "Fix r2" (≤ 4 lines) to
`work/inbox/runner-lb28-themes.md`, commit, do not edit afterwards; reply with one
done line.
