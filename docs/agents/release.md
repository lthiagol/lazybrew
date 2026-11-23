# Releasing

Only when the user asks. Work stays on `wip`; `stable` moves only on request.

## Dev build (`lazybrew-dev`)

1. Tag the `wip` tip: `git tag -a vX.Y.Z-rcN -m vX.Y.Z-rcN && git push origin vX.Y.Z-rcN`.
2. In the dev tap `lthiagol/dev` (`~/coding/homebrew-dev`, repo
   `lthiagol/homebrew-dev`), `Formula/lazybrew-dev.rb`: set `url` to
   `https://github.com/lthiagol/lazybrew/archive/refs/tags/vX.Y.Z-rcN.tar.gz` and
   `sha256` to `curl -sL <url> | sha256sum`. Commit only that file, push.

## Stable

1. The human smoke gate passes (`SMOKE-CHECKLIST.md`; add a run note).
2. Fast-forward: `git checkout stable && git merge --ff-only wip && git push origin stable`.
3. Tag the `stable` tip `vX.Y.Z` (annotated), push the tag.
4. Bump `Formula/lazybrew.rb` in `lthiagol/tap` (`~/coding/homebrew-tap`) the same
   way (url, sha256).

## Gotchas

- Both formulae pass `-X main.version=#{version}`; keep it, or `--version` prints
  `lazybrew dev`.
- **Moving a tag to another commit changes its tarball's sha256** (the archive
  records the commit id). A tap formula pinned to that tag breaks until it is bumped.
- Tags are `v<semver>`, pre-releases `vX.Y.Z-rcN` (no dot before N).
