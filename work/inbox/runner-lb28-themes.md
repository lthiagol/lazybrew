# Receipt: runner lb28-themes (LB-25)

Commits (one per AC, in order):
- AC-01 `ca32bac` palettes; Dracula, Nord; registry (no Co-Authored-By trailer: missed, not amended)
- AC-02 `12413b2` gui.theme names, validation, app, sample, README
- AC-03 `0961c05` contrast floors; frame colours; Frappé Muted
- AC-03 fix `42efc88` drop unreachable `style.ThemeNames` (deadcode)

## AC-01
- Macchiato, Frappé: all 52 values match upstream `catppuccin/palette` palette.json (0 mismatches).
- Dracula: all 11 official values appear on draculatheme.com/spec. Nord: all 16 appear on nordtheme.com.
  Reviewer: that checks each value exists, not each role mapping. Roles are in the comment tables in `theme.go`.
  Derived: Dracula Subtle `#bababa`; Nord Error `#d19aa1`, Muted `#848c9d`.
- `TestPaletteTokensAreOfficialHex`, `TestDraculaRoles`, `TestNordRoles`, `TestThemeRegistryNames`: PASS.

## AC-02
- `gui.theme: solarized` → `gui.theme: unknown value "solarized" (want one of dark, light, ...)`: PASS.
- `Catppuccin-Macchiato` loads; app applies it (`TestNewAppliesTheConfiguredTheme`): PASS.
- Config does not import the GUI; `config.ThemeNames()` is checked against the registry by `TestRegistryMatchesConfig`.
- Sample comment and README list every name: PASS.

## AC-03
- WCAG contrast of each role on the background, `contrast_test.go`. Tight: Dracula Error 4.53, Dracula Muted 3.03, Nord Info 4.64.
- Frappé Muted was overlay0 at 2.87:1 (under 3:1); now overlay1 at 3.65:1.
- **Latte is under floors; its colours are pinned.** Muted 2.30, Success 2.96, Warning 2.31, Info 4.34, Subtle 4.37.
  Recorded in `latteBelow`: the test fails if one falls further. Filed `Q-lb28-latte-contrast`.
- Frames (`theme_frames_test.go`): 8 names × {formulae selected, operation, confirm modal} at 80x24.
  Canvas base/text, role SGR tokens present, no other theme's accent. PASS (24 subtests).
  Probe: Mocha formulae frame has 60× `38;2;203;166;247` (accent), 4× `48;2;49;50;68` (surface); absent token count 0.
- Mocha and Latte byte-identical: every `frameStates` entry, dark and light, 80x24 and 140x40,
  dumped at base `1566a41` and at `0961c05`: `cmp` identical (687358 bytes). Dump test not committed.

## Gate (on `42efc88`)
- `GOFLAGS=-p=1 make ci`: exit 0. Floors: app 71.0≥64, config 93.6≥85, gui 91.7≥83, style 91.7≥88.
- `GOOS=darwin go build ./...`: exit 0. Per-commit `make test-fast` on touched packages: PASS.
- deadcode `./cmd/...`: clean (was `unreachable func: ThemeNames` before the fix).
- staticcheck: 5 findings, none in this slice (`keymap_test.go` SA1019 ×2, `launch_check_test.go` SA4006 ×3; unchanged since `1566a41`).
- No TUI launch, no brew commands.

## Questions
- `Q-lb28-latte-contrast`: keep Latte with five roles pinned, or darken them.

## Seen, not done
- `DESIGN.md:93` still lists `gui.theme` as `dark` or `light`.
- `SMOKE-CHECKLIST.md` has no entry for the new themes.

## Fix r2 (`1a58d2a`, [LB-25])
- F-r1-01 `theme.go:122`, `theme_test.go:450`: Surface comment says selection; value unchanged.
- F-r1-02: README sample regenerated from `--config-sample` (30 lines, `cmp` identical); lead-in now "Here is the sample:".
- F-r1-03: `themes` comment names `config.themeNames` and `TestRegistryMatchesConfig`. Gate: `GOFLAGS=-p=1 make ci` exit 0 (floors met); `test-fast` style+config ok.
