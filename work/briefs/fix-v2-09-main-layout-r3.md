# Fix brief — v2-09-main-layout, round r3 (orchestrator ruling)

Source: ruling on `Q-v2-09-a` in `work/questions.md` and the reviewer's r3
note. One commit per item (subject names the id), receipt
`work/inbox/runner-v2-09-main-layout-fix-r3.md` (≤ 30 lines, committed once).
Gate per ruling `Q-v2-builder-speed`.

- **Q-v2-09-a** — the Info split threshold becomes main box inner width
  **≥ 84** (one named constant), so 140x40 splits. Check every split Info view
  at 140x40 reads well at that width (fields column not cut to nonsense; prose
  wraps in its column). Update tests; regenerate frames (140x40 now shows
  the split; 80x24 unchanged); `result.md` names the changed screens.
- **Log padding** — the Log box gets the same one column of padding left and
  right as the main box. Regenerate frames; name changes.

Reply with exactly one line: `done: work/inbox/runner-v2-09-main-layout-fix-r3.md`.
