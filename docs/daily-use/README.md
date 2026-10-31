# Documents daily evaluation

Use Documents for a small real notebook during this evaluation. The frozen candidate is **build25-ed4f71f49bbf**, source ed4f71f49bbf623dd8e205571669539212068c62, minimum macOS 27.0. It is an internal ad-hoc signed build. Source-side tests and acceptance records added later do not change its installed bytes.

From the checkout root, run:

```sh
python3 -B scripts/daily-documents.py --check-only
python3 -B scripts/daily-documents.py
```

Alternatively double-click `scripts/Launch-Daily-Documents.command`. Python 3 and the existing candidate installation/archive are required. The launcher verifies candidate receipts, installed bytes/signatures, host compatibility, and that this candidate is not already running. Verification failures stop launch. It requests launch through macOS; visible startup is still confirmed in the app.

Daily data lives at `~/Library/Application Support/Vellum Internal Evaluation/Documents-build25-ed4f71f49bbf`, separately from the disposable acceptance fixture. First launch creates it with mode 0700; A and B are separate notebooks. The launcher refuses symlink paths or an existing evaluation directory with broad permissions. It never imports personal files or repairs permissions. Use local nonsynced files for evaluation; avoid confidential or irreplaceable material while qualification is ongoing. Save As may deliberately select another location.

Scope remains frozen: edit, Find/Replace, Undo/Redo, Save/Open, two-window Quit, restart, recovery. Fix reproducible blockers within these workflows before adding features. Do not change renderer, public API, packaging policy, or candidate identity as part of daily evaluation.

Each session:

1. Edit a short note; try Find/Replace and Undo/Redo after focus settles.
2. Save, Quit normally, relaunch and check exact content and selection.
3. Record any defect in ISSUE-LOG.md with exact steps and candidate identity. Keep observations separate from guesses.

Use the item 3 and item 4 acceptance records for coordinated Quit and isolated interruption results. Do not intentionally interrupt an everyday notebook to repeat the disposable recovery test. Before changing candidate versions, save and quit both windows and copy the evaluation directory to a separate local backup. A multi-day evaluation has **not** yet been completed; this kit makes it ready to start. Native acceptance so far covers the recorded document workflows, not every control or physical pointer geometry.

Build25 qualification on 2026-10-31 covered three native sessions with two normal save/quit/relaunch cycles, Unicode persistence, Find/Replace, Undo/Redo and exact saved-content checks. The synthetic fixture is separate from daily data. See [item 5](../reviews/next-five-stability/item-5.md). Earlier build24 daily data remains in the unversioned Documents directory; the launcher starts a fresh build25 profile and performs no automatic migration. Multi-day human evaluation remains outstanding.
