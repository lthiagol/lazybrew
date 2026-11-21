verdict: changes-required

# Review lb28-themes r1 (ca32bac..bfaac4d, base 1566a41)

Three low doc findings; the colours, wiring, contrast and frames are right.

| AC | Commit(s) | Verdict |
|---|---|---|
| AC-01 palettes, Dracula, Nord, registry | ca32bac | ok, except F-r1-01, F-r1-03 |
| AC-02 config, app, sample, README | 12413b2 | ok, except F-r1-02 |
| AC-03 contrast, frames | 0961c05, 42efc88 | ok (Latte per ruling Q-lb28-latte-contrast) |

## Evidence
- **V1:** `GOFLAGS=-p=1 make ci` exit 0, all floors met (app 71.0, config 93.6, gui 91.7, style 91.7). `GOOS=darwin go build ./...` exit 0.
  (A first run with GOTMPDIR under ~/.cache failed 3 `services_logs_test` cases: the long temp path got truncated in the header. That came from my setup, not the slice; it passes with the default /tmp.)
- **V2:** I fetched the upstream sources myself: `catppuccin/palette` palette.json, draculatheme.com/spec, and `nordtheme/nord` src/nord.css.
  - Catppuccin: Mocha, Latte, Macchiato and Frappé match upstream on all 26 tokens each (104 values, 0 mismatches).
  - Dracula: all 11 hex match the spec table by name. Nord: nord0–nord15 match nord.css by index.
  - Role mappings are sound: Accent/Focus purple or nord8, Error red, Warning orange or nord13, Info cyan or nord9, Pinned pink or nord15.
  - Derived colours, recomputed by hand: Dracula Subtle = 0.7·#f8f8f2 + 0.3·#282a36 = #bababa. Nord Error = 0.6·nord11 + 0.4·nord6 = #d19aa1. Nord Muted = 0.6·nord3 + 0.4·nord4 = #848c9d. All three exact.
- **V3:**
  - All 8 names load, in any case (tests).
  - `bin/lazybrew --config` with `theme: solarized` exits 1 and lists the 8 names.
  - `dark`/`light` equal `DarkTheme()`/`LightTheme()`. A missing or null `theme` still gives `dark`.
  - Explicit `theme: ""` now errors (it drew dark before), the same as `icons: ""`. That fits the brief.
  - `--config-sample` lists every name.
- **V4:**
  - Contrast, recomputed per role (WCAG): every theme meets the floors, except Latte's 5 ruled roles (2.30/2.96/2.31/4.34/4.37). Tightest: Dracula Error 4.53, Muted 3.03; Nord Info 4.64; Frappé Error 4.65, Muted (overlay1) 3.65.
  - Own harness: 8 themes × 25 states (frameStates, plus loaded Formulae with a selection, Outdated, Taps trusted/untrusted, the Trust tab, all 4 toasts) × {80x24, 140x40}, 400 frames.
  - I parsed the SGR (fg, bg) of every visible character. No character is drawn in its own background colour. The lowest pairs are box borders (1.69–1.80, like Mocha), then Muted. Each new theme draws only its own 12–13 role colours.
- **V5:** dark, light, catppuccin-mocha and catppuccin-latte frames at HEAD vs base 1566a41: 200 `cmp` comparisons, 0 differ.
- **V6:**
  - No change to go.mod, the Makefile, or the floors. No new key, env variable or live switch.
  - deadcode is clean. staticcheck shows the same 5 findings as the base (keymap_test SA1019 ×2, launch_check_test SA4006 ×3).

## Findings
- **F-r1-01 (low)** — `internal/gui/style/theme.go:122`, `theme_test.go:450`
  - What: Dracula Surface `#44475a` is labelled "current line".
  - Evidence: the spec names `#44475A` **Selection**; **Current Line** is `#6272A4`.
  - Fix must achieve: the comments call `#44475a` Selection. The value stays.
- **F-r1-02 (low)** — `README.md:122`
  - What: the block README introduces as "The start of the sample" shows `# one of the names under Themes below`.
  - Evidence: `--config-sample` prints `# dark, light, catppuccin-mocha, … dracula or nord`, so the two no longer match.
  - Fix must achieve: the README's sample block matches `--config-sample` output line for line, or the sentence no longer says it is the sample.
- **F-r1-03 (low)** — `internal/gui/style/theme.go:180`
  - What: the comment says `themes` "is the one place the theme names live" and then points to config's list.
  - Evidence: since 42efc88 the names live in two lists, `config.themeNames` and `style.themes`, kept in step by `TestRegistryMatchesConfig`.
  - Fix must achieve: the comment states the two lists and the test that ties them. Dropping `style.ThemeNames` is accepted: AC-02 allows a list in config plus a test.

## Recorded, no fix asked
- Process: AC-03 spans 2 commits (0961c05, 42efc88). ca32bac has no Co-Authored-By trailer. No rewrite is allowed, so these stay as they are.
- Seen, out of scope: the input modal placeholder uses unthemed 256-colour 240 (`#585858`, the bubbles default). It reads at 2.31:1 on Mocha, 1.73 on Frappé, 1.76 on Nord and 2.0 on Dracula. Fixing it changes Mocha frames, so it belongs in the backlog.
- Seen, out of scope: Dracula Accent on Surface (the selected row) is 3.79:1; the floors measure on Background only. `DESIGN.md:93` still says `dark` or `light` (the builder noted it too).
