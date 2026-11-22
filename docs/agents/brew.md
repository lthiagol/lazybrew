# Brew behaviour lazybrew depends on

Check brew's real output before you parse it: run the read-only command, or read
brew's source (`git clone --depth 1 https://github.com/Homebrew/brew`). Every item
below was a bug once.

## Adding a command

- Reads go on a `*Reader` in `internal/brew/<area>.go`, cached with `cached(...)`
  (`cache.go`). Writes go on a `*Writer` and end with
  `defer s.cache.InvalidateFor("<verb>")`, with a row for `<verb>` in
  `invalidateGroups` (`cache.go`). `brew.Client` (`client.go`) wires both; the GUI
  calls `client.<Area>Write.<Method>` from a task ([gui.md](gui.md) rule 2).
- Ship: the method, its cache invalidation (`writes_invalidate_test.go` covers it), a
  unit test on `brewtest.NewMockRunner`, and a fixture in `internal/brew/testdata/`
  captured from real read-only brew output — never an invented shape.

## Exit codes

- Exit 1 is a normal result for: `doctor` (warnings), `missing` (deps missing),
  `vulns` (a JSON report; exit 1 with an `Error:` line on stderr is a failure),
  `search` (no results), `cleanup -n` / `autoremove -n` with a listing, `outdated`
  (read as none). Each reader handles its own.
- The Log (`internal/gui/command_log.go`) maps exit 1 through `brew.ExitOneReport`,
  which knows only doctor, missing, vulns (Warn) and search (OK): a `cleanup -n` /
  `autoremove -n` exit 1 still shows as failed there (LB-39).
- `Execute` returns stdout and, on failure, a `*BrewExitError` whose `Stderr`
  holds stderr.

## Where the text is

- `brew doctor` prints its warnings on **stderr**; stdout is one line.
- `brew missing` prints the missing deps on **stdout** (`puts`); only the install
  hint goes to stderr.
- Without a TTY, `brew search` prints no section headers: run `--formula` and
  `--cask` separately to type the rows.

## JSON shapes

- `brew tap-info --json`: `name, user, repo, repository, path, installed,
  official, trusted, formula_names, cask_tokens`, plus (installed) `formula_files,
  cask_files, command_files, remote, custom_remote, private, HEAD, last_commit,
  branch`. There are no count keys and no `cask_names`: counts are list lengths.
- `brew trust --json=v1` lists trusted taps, formulae, casks, commands.

## Trust

- `brew install user/tap/formula` trusts that formula, not the tap.
- A tap is trusted only by `brew trust user/tap`; official taps are trusted
  implicitly.
- `brew tap user/repo/` (trailing slash) is an invalid name; lazybrew trims it.
