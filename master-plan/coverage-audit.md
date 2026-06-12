# Lazybrew — Brew Command Coverage Audit

> Audit of all Homebrew commands against what's covered in the milestones.  
> **Last reviewed against:** Homebrew 6.0.0 (2026-06-11)

---

## ✅ Fully Covered

| Command | Milestone | Notes |
|---|---|---|
| `brew install` | M6 | Formulae + casks, from search results |
| `brew uninstall` | M6 | With confirmation + dependency warning |
| `brew uninstall --zap` | M6 | Cask deep uninstall (`X` key) |
| `brew reinstall` | — | **See gap below** |
| `brew upgrade` | M6 | Single + all + batch select |
| `brew update` | M6 | From Status panel |
| `brew info` | M4 | Main panel Info tab; 6.0.0 adds binaries, installed_dependents, list_versions fields |
| `brew list` | M4 | Formulae + Casks panels |
| `brew search` | M5 | Search modal + results panel (uses `--json=v2`) |
| `brew outdated` | M4 | Dedicated Outdated panel |
| `brew tap` / `brew untap` | M7 | Full lifecycle |
| `brew tap-info` | M4/M7 | Main panel Tap Info tab; 6.0.0 adds `trusted` field and lists formulae/casks |
| `brew tap --repair` | M7 | Repair action |
| `brew trust` / `brew untrust` | M3/M7 | Full granular trust UI; 6.0.0 adds `--json=v1` flag |
| `brew services list/start/stop/restart/run/cleanup` | M4/M8/M11 | Full lifecycle; run via `f`, cleanup via `c` |
| `brew pin` / `brew unpin` | M8 | Toggle in Formulae **and Casks** panels (6.0.0 added cask pinning) |
| `brew cleanup` | M8 | With dry-run preview |
| `brew doctor` | M8 | Diagnostics display |
| `brew leaves` | M8 | Filter toggle in Formulae |
| `brew autoremove` | M8 | With dry-run preview |
| `brew deps --tree` | M4 | Deps tab in main panel |
| `brew bundle dump/install/cleanup/check/list` | M8/M11 | Full Brewfile support via `B` key menu; 6.0.0 adds parallel installs, `trusted:` entries |
| `brew home` | M7 | `o` key opens browser |
| `brew config` | M4 | Config tab in Status panel |
| `brew --version` | M9 | Shown in Status panel |
| `brew vulns` | M8/M11 | Vulnerability check via `homebrew-brew-vulns` tap; `v` key |
| `brew missing` | M8/M11 | Missing deps check (6.0.0: +casks support); `m` key |

---

## ⚠️ Gaps — User-Facing Commands NOT in Milestones

These are commands that a regular Homebrew user might use, but are **not currently covered** in any milestone:

### Must-address (should add to milestones)

| Command | What It Does | Suggested Placement |
|---|---|---|
| `brew reinstall` | Uninstall + reinstall (fixes broken installs) | **M6** — add as action key `r` in Formulae/Casks panel |
| `brew uses --installed <name>` | "What depends on X?" (reverse deps) | **M4** — add as a tab in the main panel ("Dependents" or "Used By") |
| `brew fetch [--all-platforms]` | Pre-download without installing (6.0.0: `--all-platforms` flag) | **M6** — could be useful for offline prep |
| `brew desc` | Show description (standalone) | Covered indirectly via `info`, but could enhance search results |

### Added in Homebrew 6.0.0 (new gaps)

| Command | What It Does | Suggested Placement |
|---|---|---|
| `brew exec` | Run commands from a formula's environment (like `npx`) | **P2** (nice-to-have) — niche but useful |
| `brew vulns` | Check installed packages for known CVEs (via `homebrew-brew-vulns` tap) | **P1** — could add to Status panel (M8) as a security feature |
| `brew as-console-user` | Run as the right user under MDM/root | **Out of scope** — enterprise niche |

### Nice-to-have (P2, can defer)

| Command | What It Does | Why Defer |
|---|---|---|
| `brew link` / `brew unlink` | Manage symlinks for keg-only formulae | Niche; keg-only users know what they're doing |
| `brew edit` | Open formula in editor | Niche; developer/maintainer use |
| `brew log` | Git log of formula changes | Niche; informational only |
| `brew cat` | Show formula source | Developer tool |
| `brew options` | Install options for formula | Mostly deprecated; few formulae have options now |
| `brew analytics on/off` | Toggle analytics | One-time config, rarely changed |
| `brew migrate` | Migrate renamed packages | Rare, usually automatic |
| `brew install --adopt` | Adopt existing app as cask | Very niche |
| `brew postinstall` | Re-run post-install steps | Rare troubleshooting |
| `brew completions` | Manage shell completions | Not relevant in TUI context |
| `brew formulae` / `brew casks` | List ALL available (not just installed) | Massive lists; search covers discovery |
| `brew shellenv` | Print shell env setup | Not relevant in TUI context |

### Out of scope (developer/maintainer commands)

| Command | Why Out of Scope |
|---|---|
| `brew create`, `brew audit`, `brew style`, `brew test` | Formula development tools |
| `brew bottle`, `brew bump`, `brew bump-formula-pr`, `brew bump-cask-pr` | Maintainer CI/publishing tools |
| `brew extract`, `brew irb`, `brew ruby`, `brew sh` | Developer debugging tools |
| `brew livecheck` | Maintainer version tracking |
| `brew pr-*`, `brew release`, `brew dispatch-*`, `brew determine-*` | CI/CD pipeline tools |
| `brew generate-*`, `brew typecheck`, `brew prof` | Internal tooling |
| `brew vendor-install`, `brew setup-ruby`, `brew update-reset` | Internal maintenance |

---

## Summary

| Category | Count | Status |
|---|---|---|
| ✅ Fully covered | 31 commands | In milestones |
| ⚠️ Gaps (should add) | 4 commands | Need to add to milestones |
| 🆕 New in 6.0.0 (gaps) | 2 commands | 1 P1, 1 P2, 1 out-of-scope |
| 🔶 Nice-to-have (P2) | 14 commands | Can defer post-v1 |
| 🚫 Out of scope | 20+ commands | Developer/maintainer tools |

### Verdict

We're covering **~78% of user-facing commands**. The 8 remaining gaps are small — most are one-line additions to existing milestones. The biggest gaps are:

1. **`brew reinstall`** — important for fixing broken installs
2. **`brew uses --installed`** — "what depends on this?" is very useful info
3. **`brew fetch`** — less critical, but nice for offline/slow-network prep
4. **`brew bundle check/list`** — rounds out the Brewfile support
5. **`brew exec`** (new in 6.0.0) — niche but part of the brew user workflow now
6. **`brew vulns`** (new in 6.0.0) — security feature, now covered in M8
7. **`brew missing`** — now covered in M8

Homebrew 6.0.0 did not remove or break any commands in our plan. The trust system we planned for M7 is now mandatory (was optional in 5.x), which validates our decision to include it. The `brew pin` extending to casks is additive and easy to support. The main risk is the new default **ask mode** — lazybrew must ensure brew runs non-interactively when called from the TUI (set `HOMEBREW_NO_ASK` or pipe stdin).
