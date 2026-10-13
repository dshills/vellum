# Mac native file-panel lifecycle and Finder attempt

The opt-in native consumer passes 200 real Open/Save caller-cancellation cycles
across two final-source race-instrumented processes, then normal owner shutdown
with an active Open panel and an active Save panel. The production chooser,
cancellation and main shutdown function bodies are byte-identical to the parent.
Only explicitly opted-in diagnostics and the acceptance consumer are added.

Each final run alternates 50 Open and 50 Save panels. Before caller cancellation,
the main-thread inventory observes exactly one visible tracked panel of the
requested class and a separate Go sample observes one retained operation. An
empty cancellation result is followed by zero tracked/visible/class counts and
zero retained operations before the next admission. The final caller starts with
Background context; application lifetime cancels it, the caller joins, new use
returns ErrClosed, and native main-loop cleanup has no panels or pending callbacks.
Both final processes exit 0 with empty stderr, in approximately 54.9 seconds each.
These durations include race instrumentation and concurrent compilation and are
not performance acceptance measurements.

The initial ten-cycle process failed because the fixture incorrectly expected
ErrClosed from an already-admitted call. It actually returned context.Canceled;
all ten cancellations and tracked native shutdown retirement had completed.
Internal review corrected this distinction and added a regression that rejects
nil errors, nonempty paths and other cancellation/refusal classes. The initial
failure remains raw exit 1. An intermediate 100-cycle Open-shutdown pass remains
historical V2 evidence; it does not replace either final-source process.

The final driver default/opt-in race suites and affected consumer race x3/vet
pass. The initial offscreen race suite passes. Compiled V2 isolation proves six
diagnostic symbols and the native-check macro absent in default, offscreen and
offscreen-plus-opt-in builds; its four driver probe source files exactly match
the final version. Final-only changes affect the consumer regression and scope
documentation. Diagnostic sources still exist in the source SDK payload.

Internal source review approves eight exact source files. One Prism review uses
Anthropic Claude Sonnet 5.5, covers all eight files in two chunks, returns exit 0
with no findings, and uses 16,028 input / 1,679 output tokens. The archive retains
the source groups, raw process receipts/logs, provider result, source certificate,
selected function-body identities and all 6,174 prepared dependency identities.
This is selected-source/prepared evidence, not a complete compiler/cache/SDK
dependency closure. Binaries, app bundles, caches and desktop snapshots are excluded.

## Physical Finder attempt remains unqualified

A separate unchanged-parent file-drop build and private ad hoc bundle were ready
for a real Finder drag. No drag was issued: native app selection through the UI
tool stalled for 17,293.7148 seconds despite a requested 30-second timeout, then
returned a stale window snapshot. The actual native fixture stopped after
240.807 seconds with exit 1, a READY line and context deadline exceeded. Its
bounded process lifetime worked; this supplies no receiving/routing acceptance
and does not demonstrate a framework file-drop failure. The session-level UI
attempt record is distinguished from the native terminal receipt.

## Scope

Counts cover precisely production-tracked objects; they do not independently
discover a visible panel removed by an exceptional cleanup path. Main-thread
Cocoa and Go callback observations are separate, non-atomic samples. This isolated
sequential fixture qualifies normal lifecycle retirement, not concurrent-panel
behavior, exceptional AppKit cleanup, late-callback ACKs, native heap reclamation
or actual early native-loop failure. Partial worker/loop progress is retained by
the consumer; pre-loop startup failures still rely on the parent process receipt.

This phase selects and writes no files, accesses no clipboard and changes no
preferences. Actual editor selection/persistence is independently qualified in
the native-chooser report. Physical Cancel clicks, Finder drag, human IME and
VoiceOver, cross-display behavior, broader native/GPU memory, sustained performance
and public distribution remain open. Full Mac completion is not established.
