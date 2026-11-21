# Slice: lb28-themes — Catppuccin flavours, Dracula, Nord, chosen in the config

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb27-status-tabs` closed. User request and rulings (2026-11-21),
backlog LB-25: Catppuccin (all four flavours), Dracula, Nord; **config only, no
live switching, no new key, no env override**. Never launch the TUI against the
real brew; no brew write commands.

## Themes and names

`gui.theme` accepts (case-insensitive): `dark` (= `catppuccin-mocha`, the default),
`light` (= `catppuccin-latte`), `catppuccin-mocha`, `catppuccin-macchiato`,
`catppuccin-frappe`, `catppuccin-latte`, `dracula`, `nord`. An unknown name is a
config load error that lists the valid names (as an unknown `gui.icons` already
is). Mocha and Latte keep exactly today's colors: every existing golden/frame
stays byte-identical.

## Palettes (verify against upstream before you trust these)

Catppuccin Macchiato: rosewater #f4dbd6, flamingo #f0c6c6, pink #f5bde6, mauve
#c6a0f6, red #ed8796, maroon #ee99a0, peach #f5a97f, yellow #eed49f, green
#a6da95, teal #8bd5ca, sky #91d7e3, sapphire #7dc4e4, blue #8aadf4, lavender
#b7bdf8, text #cad3f5, subtext1 #b8c0e0, subtext0 #a5adcb, overlay2 #939ab7,
overlay1 #8087a2, overlay0 #6e738d, surface2 #5b6078, surface1 #494d64, surface0
#363a4f, base #24273a, mantle #1e2030, crust #181926.
Catppuccin Frappé: rosewater #f2d5cf, flamingo #eebebe, pink #f4b8e4, mauve
#ca9ee6, red #e78284, maroon #ea999c, peach #ef9f76, yellow #e5c890, green
#a6d189, teal #81c8be, sky #99d1db, sapphire #85c1dc, blue #8caaee, lavender
#babbf1, text #c6d0f5, subtext1 #b5bfe2, subtext0 #a5adce, overlay2 #949cbb,
overlay1 #838ba7, overlay0 #737994, surface2 #626880, surface1 #51576d, surface0
#414559, base #303446, mantle #292c3c, crust #232634.
Dracula (draculatheme.com): background #282a36, current line/selection #44475a,
foreground #f8f8f2, comment #6272a4, cyan #8be9fd, green #50fa7b, orange #ffb86c,
pink #ff79c6, purple #bd93f9, red #ff5555, yellow #f1fa8c (darker bg #21222c).
Nord (nordtheme.com): nord0 #2e3440, nord1 #3b4252, nord2 #434c5e, nord3 #4c566a,
nord4 #d8dee9, nord5 #e5e9f0, nord6 #eceff4, nord7 #8fbcbb, nord8 #88c0d0, nord9
#81a1c1, nord10 #5e81ac, nord11 #bf616a, nord12 #d08770, nord13 #ebcb8b, nord14
#a3be8c, nord15 #b48ead.
The values above were written from memory. If you can fetch the upstream palette
(`catppuccin/palette` palette.json, the Dracula and Nord spec pages) without a
`go run`/`go install`, check them; if not, say so in the receipt: the reviewer
verifies them against upstream.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — Macchiato and Frappé palettes; Dracula and Nord themes
- [ ] AC-02 — config names, validation, app wiring, sample, README
- [ ] AC-03 — frames and contrast per theme

### AC-01
`palette.go`: `Macchiato` and `Frappe` as `Palette`s with the hex test style the
file already has (`theme_test.go` `mochaHex`/`latteHex`: one table per flavour,
every token). Constructors (`MacchiatoTheme`, `FrappeTheme`,
`DraculaTheme`, `NordTheme`) returning `*Theme`. Dracula and Nord are not
Catppuccin-shaped: define their 13 roles directly from their official colors
(`Accent`, `Success`, `Warning`, `Error`, `Info`, `Pinned`, `Muted`, `Subtle`,
`Text`, `Border`, `Focus`, `Surface`, `Background`), documented in a comment table
(role → color → why). Where an official color is too dim for a text role, say how
you derived a lighter one. A registry (`style.ThemeByName(name) (*Theme, bool)` and
`style.ThemeNames()`) is the one place names live. Tests first: hex tables; every
role of every theme is non-nil; the registry covers every name above; `dark`/`light`
resolve to the exact Mocha/Latte `Theme` values (equality against `DarkTheme()`,
`LightTheme()`).

### AC-02
`internal/config`: validate `gui.theme` against `style.ThemeNames()` (check the
import direction: config must not import gui; if it would, keep the name list in
`config` and have a test assert it equals `style.ThemeNames()`). Error text lists
the valid names. `internal/app/app.go`: replace the `light`/default switch with the
registry. Sample (`config.Sample()`): the `theme` comment lists every name; the
completeness tests pass. README: a Themes paragraph with the names and one line
each. Tests: each name loads; case-insensitive; unknown errors with the list;
`dark`/`light` unchanged in behaviour.

### AC-03
Contrast test over every theme: `Text` on `Background` ≥ 7:1; `Subtle`, `Accent`,
`Success`, `Warning`, `Error`, `Info` on `Background` ≥ 4.5:1; `Muted` ≥ 3:1;
`Surface` distinct from `Background`; computed with the WCAG relative-luminance
formula in the test. If an official color fails a bound, derive a role color
(documented) rather than lower the bound; if you must lower one, name the theme
and role in the receipt with the numbers. Frame tests: for each theme render
Formulae (with a selected row), an Operation overlay and a modal at 80x24 and
check the SGR of the background and accent in the frame equals the theme's
colors; Mocha and Latte frames byte-identical to the base commit.

## Done means

`make ci` green (`GOFLAGS=-p=1` if /tmp is tight); `GOOS=darwin go build ./...`;
floors not lowered; deadcode/staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Disk is tight: no large scratch. Receipt
`work/inbox/runner-lb28-themes.md` (≤ 40 lines, committed, not edited afterwards),
one commit per AC, tests first, commit subjects end with `[LB-25]`.
