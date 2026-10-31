# Documents evaluation issue log

Current candidate: build25-ed4f71f49bbf. Initial setup below used build24-b9212489afc9.

Initial setup observation (2026-10-31): ready to begin daily evaluation; launch smoke check passed. No multi-day sessions had been recorded at setup.

No open reproducible document-workflow defect was found in the scoped acceptance run. The earlier keyboard Undo failure was resolved by settling focus between actions; see [the focus and Undo acceptance report](../reviews/document-priorities/item-1.md). This is not a claim that the application is defect-free.

For each new issue, copy this template:

- ID / date / candidate / macOS:
- Workflow and affected notebook (avoid private content):
- Exact steps, including focus and save state:
- Expected:
- Actual:
- Reproduces: attempts / failures:
- Impact: data loss, blocked work, or inconvenience:
- Evidence location (redacted):
- Status / fix commit / verification:

## Recorded checks

- ID / date / candidate / macOS: QA-25-1 / 2026-10-31 / build25-ed4f71f49bbf / 27.2
- Workflow and affected notebook: synthetic B notebook; three short sessions.
- Exact steps: create Unicode note, save, Quit, reopen; create second note, insertion Undo/Redo, Find/Replace and replacement Undo/Redo, save, Quit, reopen; verify eight notes and selected n11; Quit.
- Expected: exact saved content and selection persist; Undo/Redo restores the recorded text states.
- Actual: expected content and states observed. Both normal exits before relaunch were independently confirmed.
- Reproduces: one three-session sequence; no mismatch in the recorded checks. This is not three repetitions of every action.
- Impact: no data loss or workflow blocker observed within this scope.
- Evidence location: [session report](../reviews/next-five-stability/item-5.md), [structured result](../reviews/next-five-stability/item-5-result.json).
- Status: bounded checks passed; multi-day human evaluation outstanding.

## Tool observation

- ID / date / candidate / macOS: CUA-25-1 / 2026-10-31 / build25-ed4f71f49bbf / 27.2
- Workflow and affected notebook: session-two editor focus in synthetic B.
- Exact steps: AX click on body then attempted select-all/paste; inspect unchanged text; coordinate click in visible body; inspect focus. New Note followed by Tab then established explicit body focus.
- Expected: editor focus after click.
- Actual: neither click path established observable editor focus through the tool; attempted input left saved note unchanged.
- Reproduces: one AX-click attempt and one coordinate-click attempt in this session; physical mouse reproduction not performed.
- Impact: limited tool automation; no confirmed product data loss or blocker.
- Evidence location: [session report](../reviews/next-five-stability/item-5.md).
- Status: not reproduced with physical input on the later v1.1.2 suite build4, source 121dde859369b3ce091fe66cdfa2d34367e23be1 (2026-10-31); historical build25 was not physically retested: both body and title accept text after clicking. The earlier tool delivery observation is not a confirmed product defect. See [focus qualification](../reviews/application-readiness/item-1.md).
