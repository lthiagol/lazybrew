# Releasing

Only when the user asks. Work stays on `wip`; `stable` moves only on request.

Tap changes go through the user's `/tap` skill (`~/.claude/skills/tap`). You cannot
run it: suggest the command (e.g. `/tap bump lazybrew v1.2.0`) and let the user run it.

## Dev build (`lazybrew-dev`)

1. Tag the `wip` tip: `git tag -a vX.Y.Z-rcN -m vX.Y.Z-rcN && git push origin vX.Y.Z-rcN`.
2. `/tap bump lazybrew-dev vX.Y.Z-rcN` (the dev tap `lthiagol/dev`).

## Stable

1. The human smoke gate passes (`SMOKE-CHECKLIST.md`; add a run note).
2. Fast-forward: `git checkout stable && git merge --ff-only wip && git push origin stable`.
3. Tag the `stable` tip `vX.Y.Z` (annotated), push the tag.
4. `/tap bump lazybrew vX.Y.Z` (`lthiagol/tap`).

## Gotchas

- Both formulae pass `-X main.version=#{version}`; keep it, or `--version` prints
  `lazybrew dev`.
- **Moving a tag to another commit changes its tarball's sha256** (the archive
  records the commit id). A tap formula pinned to that tag breaks until it is bumped.
- Tags are `v<semver>`, pre-releases `vX.Y.Z-rcN` (no dot before N).
