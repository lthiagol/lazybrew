verdict: ok

# reviewer — lb18-config-sample r1

Reviewed `wip` after 656de50 → runner commits ddbb3d0, fb8b195 (AC-01), 6717737 (AC-02), f11cea1 (AC-03); receipt 2a8b464.

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 `config.Sample()` | met | `sample(Default())` walks Config by reflection; 6 tests pass. Mutations: dropping `brew.taps_ttl` from `sampleComments` fails TestSampleCommentsCoverEveryKey. Writing `mouse: false` in place of the default fails TestSampleLoadsToDefault. Setting the default theme to light and `DefaultCasksTTL` to 7m makes the sample print `theme: light` and `casks_ttl: 7m`, and TestSampleLoadsToDefault still passes. |
| AC-02 flags | met | The branch sits before `app.New` in `main.go`. With `env PATH=/nonexistent`, `--config-sample` exits 0. `--config-create --config <c>/a/b/c.yml` creates dirs 755 and the file 644, prints the path, exits 0. A second run prints `Error: … already exists; not touched`, exits 1, and the sha1 is unchanged (e7bb0191…). A read-only dir gives `Error: open …: permission denied` on stderr (stdout empty), exit 1; a missing subdir under it gives `Error: mkdir …: permission denied`, exit 1. Mutation O_EXCL→O_TRUNC fails TestConfigCreateRefusesExistingFile. `--help` lists both flags. |
| AC-03 README | met | Both flags are in the flag table, plus a "Create a config" paragraph. The hand-written block is replaced by the first lines of the sample, which match `--config-sample` output, and the TTLs are pointed to. `> file` instead of `>>` is a correct deviation (`>>` would duplicate keys). |

**V1 gate:** `make ci` exit 0 (fmt, vet, `-race`, all floors OK, config 87.4 ≥ 85, build). `GOOS=darwin go build ./...` exit 0.
**V4 sample:** read in full. It loads to `Default()`. Durations print as `500ms`/`5m`/`30m`/`1h`/`30s`. The `sidebar_width 15-50` comment matches the clamp in `gui.sidebarWidth`, and the env header matches `Load`. No stale wording against README or `config.Config`. Gaps are in F-r1-01.
**V5 code:** the `DefaultPath()` extraction is equivalent. A throwaway reviewer test (not committed) passed for `Load("")` with HOME holding a config, `Load(path)`, and HOME="" falling back to defaults. No new dependency (go.mod/go.sum untouched), no goroutine, no global state added (`durationType` is a constant-like var). Errors go to stderr.
**V6:** one commit per AC (plus the AC-01 alignment commit), floors unchanged, no golden/testdata changes.

## Findings

F-r1-01 (minor) — internal/config/sample.go `sampleComments` — the TTL lines don't say that 0 or less keeps the default (`loadFile` resets it), so a user may write `0` expecting no caching. `update_stale_hours` doesn't say values below 1 are an error. — Evidence: `loadFile` TTL loop and the `UpdateStaleHours < 1` check. — Fix: the prose names both rules ("0 keeps the default", "1 or more").

F-r1-02 (minor) — internal/config/sample_test.go:120 — TestSampleFollowsDefaults hard-codes the real default `5m`. When `DefaultCasksTTL` is set to 7m, this test fails even though the sample follows the change. — Fix: compare against `sampleDuration(DefaultCasksTTL)`.

F-r1-03 (minor) — cmd/lazybrew/configfile.go `runConfigFlags` — when `--config-sample --config-create` are given together, it prints the sample and skips the create without saying so. — Fix: reject the combination with a stderr line and non-zero exit, or create and say so; add a test.

F-r1-04 (minor) — internal/config — no committed test that `Load("")` reads the file at `DefaultPath()` (the `~` path). The reviewer's check passed, but nothing pins it. — Fix: add a HOME-redirected `Load("")` test.

Note: the receipt reports one accidental TUI launch by the runner (`go run ./cmd/lazybrew`). It was not caused by either flag, and the runner reports no brew write.
