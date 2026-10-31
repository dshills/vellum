# Multi-day Documents evaluation

Status on 2026-10-31: **pending — zero qualifying operator days recorded**.
Three short automated sessions and a later physical focus check are useful scoped
acceptance; they are not ordinary multi-day use. Keep build25-ed4f71f49bbf frozen
for this evaluation and use the versioned profile from the daily-use launcher.
Do not migrate to a different candidate mid-evaluation or count a different app.

Use Documents for an ordinary small notebook on at least three separate local
calendar dates. On each qualifying day, edit a note, Find/Replace, Undo/Redo,
save, Quit normally, relaunch and confirm exact content. Use the app's normal
save/close decisions; do not interrupt the everyday profile for recovery testing.
Keep existing data and backups. Reproducible crashes, lost edits, incorrect history
or save failures are blockers and belong in ISSUE-LOG.md before optional polish.

Record only metadata in sessions.json after actual use; never copy notebook text
or confidential details. Each entry has this shape (replace the date with the
actual day; this example is **not** a recorded session):

```json
{
  "date": "2026-10-31",
  "kind": "operator",
  "checks": ["edit", "find-replace", "undo-redo", "save", "quit-restart"],
  "open_blockers": []
}
```

Only include checks actually performed. Use kind `automated` for tool-driven
sessions. Add issue IDs to open_blockers for unresolved blockers, and update those
IDs after a separately recorded fix verification; preserve the dated issue history
in ISSUE-LOG.md. A different candidate requires a fresh ledger and qualification.

Run `python3 -B scripts/check-daily-evaluation.py` to report coverage, or add
`--require-recorded` for a nonzero exit until three distinct operator days contain
all checks and no listed open blockers. Future dates are refused; multiple entries
on one date count once. This checks declared coverage, not whether operator reports
are true, and does not independently certify correctness or production readiness.
The empty ledger is intentional until real operator sessions are reported.
