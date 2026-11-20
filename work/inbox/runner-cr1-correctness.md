# Receipt — runner / cr1-correctness

**Commits:** AC-01 `02bd452` · AC-02 `044b1c5` · AC-03 `05f0f2f` · AC-04
`227b678` · AC-05 `5479014` · AC-06 `0a8547e`.

**AC-01 [C1]** `panelData.nameAt(i)`/`selectedName()` read the typed slice;
16 call sites moved. Pre-fix wrong names sent: `some-very-long-formula`,
`font-jetbrains-mono-ne`, `someuser/homebrew-a-ve`. Post: all pass. Services
stop and Search install pass pre-fix too (guards). 9 test files gained typed
rows; one tab test moved Deps → Used By (Deps is model-drawn).

**AC-02 [C2]** Pre-fix: `DATA RACE`, `OnExec got <nil>`. Post: pass under
`-race`, 200 iterations.

**AC-03 [C6]** `brew.ParseInfo` (new) feeds `parsePackageInfo`. Formula lost
JSON tags and `Installed`; Cask keeps tags (`UnmarshalJSON` reads them).
Pre-fix: formula Version `""`, cask decode error. Post: `1.8.2 (bottled)`.

**AC-04 [C5]** Pre-fix: nil-pointer panic on the key route. Post: pass.

**AC-05 [C3, C4]** Pre-fix: 6 ticks (want 1) after Init + 5 refreshes; 3 after
Init + 3 R presses. Post: 1 and 1, `-race` clean. Init arms auto-refresh
in both branches (Q-cr1-init-autorefresh).

**AC-06 [C12, C13]** Pre-fix: `error cut split a rune`; debug line on stderr.
Post: valid UTF-8; stderr quiet after one warning line.

**Gate:** `make ci` exit 0 (coverage app 70.0/64, brew 81.0/78, gui 86.4/83).
`GOOS=darwin go build ./...` ok. `make test-fast` per commit ok. Dead-code
and staticcheck not run (not a dead-code slice).

**Questions:** `Q-cr1-cask-name`, `Q-cr1-init-autorefresh`.

**Seen, not done**
- **Cask names are display names** (`Q-cr1-cask-name`): breaks cask actions.
- `selectedItem()`, `Cask.UnmarshalJSON`: unused in production → cr5.
- `panelData.results` duplicates `m.searchResults` → cr4.
- `Formula.InstallPath` never set; brew v2 has none.
- Unused JSON tags: Tap, TrustEntry, SearchResult, DoctorWarning, BrewConfig.
- gui.go ~1129 generic modal path keeps a non-modal result. No panic.
