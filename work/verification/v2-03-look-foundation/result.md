# v2-03-look-foundation — frame result

`before/` = copy of `work/verification/v2-02-layout-bugs/after/` (the v2-02
frames, captured on the v2-02 HEAD with real Homebrew, 7.0.15).
`after/` = this slice's frames, captured on the AC-05 HEAD (`d58fb96`) with the
Unicode icon set (`LAZYBREW_ICONS=unicode`, set in `frame_dump_test.go.txt`),
real Homebrew. `after/140x40/*.ansi` is the colour evidence.

Every row below is explained by AC-01..AC-05. Nothing unexplained was found.

| screen | 140x40 | 80x24 |
|---|---|---|
| 01-status | Header row added (name, Homebrew version); sidebar titles now in the borders with icons; the Status box shows its line count in its border; the footer is one row ending `? more`. | Same, with the sidebar's 24-cell boxes cutting labels (`not loaded` → `no…`); the name always stays whole. |
| 02-formulae | Same chrome; the main breadcrumb `Formulae › List` is the border title; the list rows and columns are unchanged. | Same; the tab bar fits the box now, with `Files` whole. |
| 03-formulae-info | Breadcrumb `Formulae › Info` in the border; the info text is unchanged. | Same. |
| 04-formulae-deps | Breadcrumb `Formulae › Deps` in the border; the deps list is unchanged. | Same. |
| 05-outdated | Header shows no outdated badge: the machine has nothing outdated, so the count is known but zero and the badge is correctly omitted. Sidebar titles and the footer as 01. | Same. |
| 06-taps | Breadcrumb `Taps › Tap Info`; the sidebar count and the footer as 01. | Same. |
| 07-confirm-modal | The modal is in the same place and drawn the same; the footer shows the modal's keys: `y yes  n no  esc cancel`. | Same. |
| 08-help | The help box is unchanged; its last row is the footer, `?/esc close` (help's own keys). | Same; the help box still cuts at 24 rows (v2-05 scrolls it). |
| 09-search | Search title in the border; the search row and results are unchanged. | Same. |
| 10-toast | The toast is in the same place as in v2-02: it sits on the last body row, which is the Log box's bottom border at this size. Pre-existing; the toast restyle is v2-04. | Same. |
| 11-deps-bash | Breadcrumb `Formulae › Deps` in the border; the deps list of bash is unchanged. | Same. |

## Notes

- **The body is still `height - 2` rows.** The one-row header and the one-row
  footer replace the old two-row footer. Inside it, each sidebar box gained the
  row its title used to take, and the main content gained the breadcrumb row
  back, because both titles now sit in the borders.
- **Header counts are known only when the panels are loaded.** On the Status
  frame the header shows the version but not `125 formulae`, because the
  Formulae and Casks panels have not been fetched; the dashboard's counts are
  text, not model data. This is by design (`known` only), see the runner's
  questions in `work/questions.md`.
- **Sidebar Status count is the dashboard's line count** (16), because the
  Status panel's items are dashboard lines. This is the count the border
  shows; it is not a count of anything the user recognises.
- **80×24 narrow labels.** The sidebar is 24 cells at 80 columns, so
  `not loaded` is cut to `no…` in the border; the name is always whole.
