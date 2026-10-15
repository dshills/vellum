# Managed workbench replay checkpoint

This checkpoint adds the concrete built-in Workbench to the deterministic
software adapter. It also fixes a desktop code path: a native tab button selected
its native page but did not emit the Tabs change event that Vellum expected, so
the owner workspace stayed on the old active panel. Each actual tab button now
queues a checked owner activation. Tab, split and drag requests capture a
revocable input scope, preventing requests queued before disable, hide, dialog
dismissal or modal exclusion from becoming valid again after reactivation.

Manual preflight evaluates every hosted panel once and validates the complete
graph before mounting. Namespaced inspection/input IDs separate repeated local
IDs; hidden panel records retain identities and drafts. Session windows use
independent software scenes and notify their group on ordinary close. Explicit
manual mode suppresses desktop bounds capture/timers and refuses persistence,
autosave, existing-window resizing and native context-menu input. Failed
construction releases the manual lease without closing application registrations.
The detailed contract is [managed replay](../../../specs/replay.md).

New regressions exercise actual routed tab clicks and field edits at batches 1,
7 and 64, comparing model state, measured semantics/geometry and completed
software pixels. Other checks cover tab reordering and split dragging with an
explicit clock advance, stale header tokens, retained uncommitted drafts,
disabled/hidden/modal input revocation, floating-window closure, construction
rollback and a stored 10,000-row, twenty-column grid with scoped cell input and
scrolling. The independent replay consumer tests the new public target helpers
without accessing driver internals. The offscreen fixture explicitly uses its
Control+A keymap; it does not qualify physical Mac shortcuts.

The retained diagnostic attempts include fixture compilation errors, incorrect
modifier encoding, a drag without advancing the explicit clock, and an invalid
direct Body.Close call outside native event ownership. The corrected closure
test uses SystemWindow.CloseReq. These failed attempts are preserved and are not
counted as successful checks. The initial actual tab route failed because of
the real owner-synchronization bug described above.

Before commit, final check receipts, source hashes, internal review and
injected-defect controls must be retained in
[qualification](records/qualification.json) and [archive index](records/archive-index.json).
The checkpoint also requires one Prism Anthropic Claude Sonnet 5.5 review.
The Mac native fixture is compiled only. No GUI, GPU, OS input-source, VoiceOver,
clipboard or appearance test runs while the human uses the machine.

This is an enabling checkpoint, not complete P9 acceptance. The independently
versioned large professional workspace record, worker-built views/cancellation,
canvas/GPU observations, current native lifecycle, sustained frame/memory targets,
P4/P6/P9/P11, idle/overnight, physical input/displays, human accessibility,
interoperability and distribution qualification remain open. SDK version
v0.1.0-alpha.1 and experimental Mac status are unchanged.
