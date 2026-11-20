# Fix round r1 — slice `v2-05s-services`

> The slice contract (Safety section first) still applies. Findings from
> `work/inbox/reviewer-v2-05s-services-r1.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r1-01 | major | The state guard follows brew: a service is **loaded** (stop/restart apply) when its status is `started`, `error` or `scheduled` (anything but `none`/`stopped`/unknown-not-loaded — read brew's `formula_wrapper.rb` on this machine, read only, and cite it in a comment). `S`/`r` on `error` and `scheduled` open their confirms and run; the Info key line and footer offer stop/restart for them (and `s start` for `error` too, since brew allows re-starting a failed unit — state what you chose). Tests for `error` and `scheduled` covering guard, hints and confirm. |
| F-r1-02 | minor | After a **failed or cancelled** services task, reload Services (list + info) **without** the "Data refreshed" success toast; the error / needs-sudo toast stays the last word. After a successful one, keep today's behaviour. Test. |
| F-r1-03 | minor | The harness clears/expires toasts before frame 19 so the 80×24 Info key line is visible; re-capture 19 (both sizes, `.ansi` 140x40) and add a `result.md` line. |

## Rules
- Safety: never run `brew services start|stop|restart|run|cleanup|kill`. The harness still escapes the stop confirm.
- One commit per finding, subject naming the id. Stage by name. Gate (CGO_ENABLED=1 for -race).
- Receipt `work/inbox/runner-v2-05s-services-fix-r1.md` (≤ 25 lines), committed exactly once. Worktree clean. No push.
