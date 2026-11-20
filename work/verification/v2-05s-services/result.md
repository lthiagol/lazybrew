# v2-05s-services — frame result

`before/` is `work/verification/v2-05-content-polish/after/` (frames 01–18).
`after/` is this slice, dumped by `frame_dump_test.go.txt`: the v2-05 harness
with two new screens, 19 and 20. `140x40` has `.txt` and `.ansi`; `80x24` has
`.txt` only, as before.

Captured with the real Homebrew on this machine (three services: dbus stopped,
moshi-hook running with pid 800, pulseaudio stopped), the Unicode icon set.
Reads only, plus one `S` that opened the stop confirm for moshi-hook and was
cancelled with `esc`. No brew services write ran: moshi-hook is still running
with the same pid, and its `ActiveEnterTimestamp` is still
`Wed 2026-11-18 11:37:48 -03`.

## Per screen

| screen | 140x40 | 80x24 | what changed |
|---|---|---|---|
| 01–18 | same | same | Byte-identical to `before/`, at both sizes (`diff -rq`). |
| 19-services-info | new | new | Services → moshi-hook (second row) → Info. Status `▶ running`, PID 800, User agent, Loaded and Registered yes, Unit file, Command, Working dir, Log (one row: log and error log are the same file). The key line reads `S stop  r restart  c cleanup`, and the footer leads with `S stop  r restart`. At 80 wide the long paths are cut with `…`. |
| 20-services-stop-confirm | new | new | `S` on moshi-hook opens the stop confirm: danger chip `STOP`, `Stop moshi-hook?`, the consequence `The service stops until you start it again.`, `[y]es [n]o esc cancel`. Nothing was answered; the harness sent `esc`. |

Frames 10 and 13 (the toast frames) are unchanged, so the toast wrap leaves
one-row toasts as they were.

## Checks

- `diff -rq before/ after/`: only `19-*` and `20-*` differ, and they are new.
- Every frame (40 `.txt`: 18 + 2 at each size) is exactly its W×H in cells,
  counted with east-asian widths.

## What the frames do not show

- The needs-sudo toast: no system service on this machine. Covered by
  `TestServicesWriteNeedsSudoIsNamed` and `TestServiceNeedsSudoToastNamesTheCommand`.
- A scheduled service and a cron schedule: none here. Covered by
  `TestServiceInfoShowsSchedule` and `TestServicesServiceInfoParsesSchedule`.
- The stopped footer (`s start  f run`): frame 19 is a running service. Covered
  by `TestServicesFooterLeadsWithStateActions`.
- The long sudo toast wrapped over rows: covered by `TestToastWrapsALongMessageWhole`.
