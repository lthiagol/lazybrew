verdict: changes-required

# reviewer lb29-footer-redesign — r1 (a918265..8eb3291)

Judged with the rulings Q-lb29-sample-width (rule-derived 140 row, `? more` at 140) and
Q-lb29-info-keylines (Tap/Services Info key lines in `·` style) applied.

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 renderer | ok (F-r1-03 low) | footer.go fitFooter/renderFooter; own sweep 60..160 step 2 on Status, Formulae, Services: never wider, no trailing `·`/`│`, drops are a prefix of `C R 1-7 Tab / q`, ≥3 panel actions while a droppable global remains, arrow hint always kept. Separators in real frames: dark `38;2;108;112;134`, light `38;2;156;160;176` (theme Muted) on all 514 `·` and 60 `│`; never accent. |
| AC-02 key tables | ok | Mutations: keyArrows removed from globalHints → TestFooterShowsTheArrowHintOnEveryPanel, TestPanelKeysAreOnFooterAndHelp fail; keyScroll help removed → TestKeymapMatchesDispatch, TestPanelKeysAreOnFooterAndHelp fail; keyTabs help removed → same. Restored. No dispatch code changed; keymap/dispatch tests pass. |
| AC-03 frames, help, README, flows | changes-required | F-r1-01, F-r1-02. Flows TestFooterFlow{Status,Formulae}At140 pass. README matches the drawn rows. |

## Gate (my runs)
- `GOFLAGS=-p=1 make ci` → exit 0, all floors met. (A first run with `GOTMPDIR` set failed the 4
  path-length Logs tests the receipt warns about; environment only, not the slice.)
- `GOOS=darwin go build ./...` → exit 0. deadcode: nothing. staticcheck: the same 5 as base 8af0ef1.
- No go.mod change; floors unchanged; one commit per AC in order, receipt separate.

## V4 frame diff (own harness, 17 scenes × 80x24/100x30/140x40/200x50 × dark/light, base 8af0ef1)
- Last row differs on every scene except help (`?/esc close`, unchanged).
- Body rows differ only in: help (key column lines), Tap Info key line, Services Info key line
  (ruled). Goldens changed in the diff are footer/key-line expectations only.

## Findings

F-r1-01 (medium) — internal/gui/keybindings.go:38,40 (with help.go:121-126 helpUnit) — the new
helpKey texts `j/k up/down` and `h/l left/right` widen the global section's key column from 9 to
14 cells, so at 100x30 the help screen now clips a word: base `Home/End   Top/bottom (g/G outside
Search)`, HEAD `Home/End        Top/bottom (g/G outside Sear…`. No `…` on any base help frame;
one on HEAD (100x30, both themes). The brief requires no clipped word. — Fix must achieve: no help
line clipped at 80x24, 100x30, 140x40, 200x50 that was whole on base, and a test that fails on
such a clip (e.g. no `…` in the help frame at 100x30), still listing j/k, h/l, Shift+Tab.

F-r1-02 (low) — internal/gui/keybindings.go:35 `keyNext` key `Tab` — the footer draws `Tab next`;
the brief's target and global group say `tab next`. flows/footer_test.go and the receipt pin
`Tab next`. — Fix must achieve: the footer shows `tab next` (the help may keep `Tab`), flows
updated, keymap tests still pass.

F-r1-03 (low) — internal/gui/modal/frame.go keyLine — the brief says "Modals and the help screen
use the same `·` separator" and AC-01 says "(and modal/help hint rows)"; the dialog's own key row
still reads `[y]es   [n]o   esc cancel` (spaces). The receipt names it untouched but raises no
question. — Fix must achieve: the modal key row uses the muted `·` separator, or a recorded
orchestrator ruling in work/questions.md that the dialog row is out of scope.

## Seen, not a finding
- Status (and an open overlay on Status) shows `R refresh` in both groups at 140/200 — pre-existing
  (base listed it twice too); runner noted it.
