# Item 2 — external Save/Open integrity

Build24 was exercised with two independent seven-note sample notebooks in a new
private acceptance workspace. Real native Save As panels selected distinct
Acceptance-A.vellumnotes and Acceptance-B.vellumnotes outputs. Each file was
independently decoded after Save and checked against its literal title/body,
seven-note count and selected n10 identity; [results](item-2-result.json) retain
exact file hashes and UTF-8 byte lengths for synthetic acceptance content.

A then used the actual Open chooser to select its saved file. The window reported
Opened notebook, seven total notes, Acceptance A selected and its exact Unicode
body. B was independently saved and decoded; its native re-open is not claimed.
The existing external-file regression separately covers Save As path admission,
clean history, subsequent Save without reopening a chooser, default-path isolation,
explicit reopen and detached restored history. Its race execution is recorded
before commit. It is controlled service/model evidence, not physical input.

No production change was justified by this integrity check. Native chooser return
left synthetic key-window focus unreliable; this remains an unresolved targeting/
focus observation rather than a claim of whole-window restoration. Quit cancellation,
restart/recovery and other consumers are separate later items. Frozen build24
bytes and existing personal notebooks are unchanged.

Focused external Save As/reopen regression passed under Go 1.25.6, race and
offscreen. Prism default staged review completed with exit 0 and full coverage.
Two low findings questioned host-version chronology and n10 among seven notes.
Both were checked against observed data: sw_vers reports 27.2, and persisted IDs
are n1–n6 plus n10 with next_id 11 (IDs are not row counts). The result now retains
that direct command output and actual ID vectors. No file hash/body/count was
changed to satisfy the reviewer. [Original review](item-2-prism.json).
