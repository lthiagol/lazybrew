# runner — lb18-config-sample

**Commits (AC order)**
- AC-01 `ddbb3d0` config.Sample() from Default(); `fb8b195` aligns the header's env lines (fix to AC-01, separate commit, no amend)
- AC-02 `6717737` `--config-sample` / `--config-create`
- AC-03 `f11cea1` README

**AC-01** `go test ./internal/config/ -run Sample -v -count=1` → PASS: TestSampleLoadsToDefault, TestSampleHasEveryKeyOnce, TestSampleCommentsCoverEveryKey, TestSampleHeaderNamesEnvOverrides, TestSampleFollowsDefaults, TestSampleDuration. The default-change test uses the unexported `sample(cfg)` (`Sample()` = `sample(Default())`): casks_ttl 7m and auto_refresh_pause 0s show up in the output.

**AC-02** `go test ./cmd/lazybrew/ -v` → PASS: TestConfigSamplePrintsToStdout, TestConfigCreateWritesSampleAndMakesDir, TestConfigCreateRefusesExistingFile (bytes unchanged; stderr names the path and "not touched"), TestConfigCreateDefaultPath (HOME redirected). By hand, with `env PATH=/nonexistent` so no brew can run:
```
$ ./bin/lazybrew --config-sample | head -4      → header lines, exit 0
$ ./bin/lazybrew --config-create --config $S/lb18/c.yml
/…/scratchpad/lb18/c.yml                        exit 0
$ ./bin/lazybrew --config-create --config $S/lb18/c.yml
Error: /…/c.yml already exists; not touched     exit 1
$ ./bin/lazybrew --config-sample | cmp - c.yml  → identical
```
The flag branch in `main()` sits before `app.New`; `runConfigFlags` does not reference the app.

**AC-03** README: two flag rows, a "Create a config" paragraph, and the sample's first 22 lines in place of the hand-written block. `diff` of that block against `--config-sample | head -22` is empty.

**Gate**
- Each commit: `make test-fast PKG=./internal/config/...` and `./cmd/lazybrew/...` → ok.
- `make ci` → exit 0 (fmt-check, vet, `-race` tests, floors: config 87.4 ≥ 85, all others OK, build).
- `make fmt-check` → exit 0. `GOOS=darwin go build ./...` → exit 0.

**Deviations from the brief**
- No default-path function existed (`loadFile` built the path inline). Extracted `config.DefaultPath()`; load behaviour is unchanged.
- README shows `> file` rather than the brief's `>>`: `>>` would append a second copy of every key to an existing file.
- File modes 0755/0644 are subject to umask; not asserted.

**Questions filed:** none. **Seen, not done:** `internal/app/app.go` `enableDebugLog` builds the same `~/.config/lazybrew` path inline; left alone.

**For the next session**
- While checking, I launched the app by mistake (`go run ./cmd/lazybrew </dev/null`). No brew write happened: brew's FETCH_HEAD is still 06:14:31, and auto mode updates only after 6 h.
- A `make run` (pid 1238068, started 07:13, repo cwd) is not mine; left running.

## Fix r1

- Commits, one per finding: F-r1-01 `e61fd8d`, F-r1-02 `22a6e03`, F-r1-03 `327b323`, F-r1-04 `4201a0d`. README's sample block was kept in step with the new TTL prose (`diff` empty).
- F-r1-01: `TestSampleStatesTheRules` failed on 7 lines before the fix (six `*_ttl`, `update_stale_hours`); passes after.
- F-r1-02: with `DefaultCasksTTL` set to 7m, `TestSampleFollowsDefaults` passes; the mutation was reverted and `config.go` is clean.
- F-r1-03: `TestConfigSampleAndCreateTogetherAreRejected` failed (exit 0) before the fix. By hand: exit 1, stdout empty, one stderr line, no file.
- F-r1-04: passes on current code. A first draft built the path with `DefaultPath()` and survived a mutation, so it now pins the literal `~/.config/lazybrew/config.yml`; the mutation fails it.
- Gate: `make ci` exit 0 (fmt-check, vet, `-race`, floors; config 92.6 ≥ 85). `GOOS=darwin go build ./...` exit 0. Per commit: `make test-fast PKG="./internal/config/... ./cmd/lazybrew/..."` ok.
- Flags via `env PATH=/nonexistent ./bin/lazybrew`: sample exit 0; create gives 0755 dir and 0644 file; a second create exits 1 and leaves the file equal to `--config-sample`.
- No TUI launch, no brew write. No questions filed.
