# Releasing

Only when the user asks. Work stays on `wip`; `stable` moves only on request.

## Dev build (`lazybrew-dev`)

1. Tag the `wip` tip: `git tag -a vX.Y.Z-rcN -m vX.Y.Z-rcN && git push origin vX.Y.Z-rcN`.
2. In `~/coding/homebrew-tap`, `Formula/lazybrew-dev.rb`: set `url` to
   `https://github.com/lthiagol/lazybrew/archive/refs/tags/vX.Y.Z-rcN.tar.gz` and
   `sha256` to `curl -sL <url> | sha256sum`. Commit only that file (that clone's
   git config has the work email: commit with
   `-c user.name="Luiz Thiago" -c user.email=lthiagol@gmail.com`), push.

## Stable

1. The human smoke gate passes (`SMOKE-CHECKLIST.md`; add a run note).
2. Fast-forward: `git checkout stable && git merge --ff-only wip && git push origin stable`.
3. Tag the `stable` tip `vX.Y.Z` (annotated), push the tag.
4. Bump `Formula/lazybrew.rb` in the tap the same way as the dev formula (url,
   sha256).

## Gotchas

- **`Formula/lazybrew.rb` does not pass the version.** Its build line has
  `ldflags: "-s -w"` without `-X main.version=#{version}` (the dev formula has it),
  so a stable install prints `lazybrew dev` for `--version`. Add it on the next bump.
- **Moving a tag to another commit changes its tarball's sha256** (the archive
  records the commit id). A tap formula pinned to that tag breaks until it is bumped.
- Tags are `v<semver>`, pre-releases `vX.Y.Z-rcN` (no dot before N).
- `.github/workflows/stable-ci.yml` still has a `tap-update` job that runs on every
  `v*` tag and fails (it never checks out the repo). The manual bump above is the
  process; the job is a known leftover (LB-41).
