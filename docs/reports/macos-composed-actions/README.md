# Composed native action acceptance

The all-component Mac lifecycle fixture now exercises actual in-process Cocoa
selectors across checkbox, switch, toggle, slider, number field, radio group,
segmented control, selector, button, disclosure heading, tabs, text field, modal
buttons and managed tree rows. Each window executes 20 steps. Direct owner model
setup and editor undo/redo remain separate from these native action observations.

A native queue admission cannot qualify a step. The runner posts an acknowledged
owner check behind each admitted action before retrying; exact callback counts
and accepted controlled values must match. Accepted non-idempotent presses wait
for native publication without resubmission. A callback request accepted by an
asynchronous tree worker is distinct from completed index/geometry publication.
The tree completion predicate additionally requires idle worker credit, a sealed
nonpending index, the expected visible count/disclosure state and matching native
scope/incarnation/generation/row metadata.

The tree retains 64 stored nodes, with four children under n0. Native selection
changes the active row to n0; collapse reduces visible rows to 60 and expansion
restores 64. Disclosure reserves one window-scoped task and its owner completion
credit before Prepare. The immutable builder runs off-owner in 64-operation steps.
Capacity refusal preserves the earlier request; window retirement cancels the
worker and suppresses its later commit. Normal successful steps require the
worker and owner completion to release their credit before closing the window.

The corrected native runs pass five windows/100 steps in 43.76 seconds and
20 windows/400 steps in 166.42 seconds. All 500 steps have exactly one attempt and
one admission, accepted owner state and completed native publication. Every
cycle has 18 completed AX rows and 1000 contained density labels. Both processes exit
zero with source/prepared endpoints identical and final main-thread cleanup:
no window titles, native outstanding operations, accessibility windows or pending
callbacks, and callbacks stopped. The 20-cycle final post-GC sample is 50,154,984
Go heap bytes and 354,140,160 RSS bytes; this is not a steady-state memory claim.

Focused native-command/fixture race tests, native vet and the offscreen acceptance
race suite pass. The latter includes real manual software layout of the density
controls; its native Extension allocation remains a software placeholder, so it
does not certify native Grid/Tree/canvas composition. Native runs provide that
separate evidence. The worker manual test verifies no builder runs before selected
worker execution, capacity refusal before Prepare, completed collapse/expansion,
credit release and suppressed late commit after window retirement.

[Validated summary](records/final-summary.json),
[qualification](records/qualification.json) and
[evidence index](records/evidence-index.json) bind final source, reviewed producers,
raw failed/corrected checks, partial action records, internal review and the
single Prism checkpoint. Native binaries are excluded; hashes remain in receipts.

Internal review caught an async retry hazard; owner acceptance now latches the
request while native completion awaits the worker and geometry. An injected held
completion regression requires exactly one perform. The first live run targets
the containing disclosure AXGroup and fails before completing any cycle, with
27 attempts and zero admissions. The corrected runner selects AXDisclosureTriangle
and tests same-domain group/heading ambiguity. A first unit fixture omitted
required window geometry and failed; the corrected fixture passes. Raw failures
are preserved. Every later action emits a bounded scalar record even on error.

These checks exercise native selector→owner→native-publication paths. They do not
synthesize physical pointer/key events, certify human VoiceOver/IME, validate menu
popup/tray/notifications/Finder workflows, move data views through a native modal,
or meet frame/pacing/input budgets. The earlier 1000-cycle retention result is on
committed bf3b1b5 with its earlier fixture. This changed hierarchical/interacting
workload needs its own long campaign; five/twenty-cycle results do not transfer
that qualification or complete P6/P11. No OS preferences or clipboard are changed.

Prism covered all ten selected files in three provider calls, reporting three
medium findings (raw exit 1). The run summary now contains direct exit/count/hash
traceability. Failed Prepare no longer cancels an unprepared request or replaces
its original error; fixture failures retain the first cause through later success.
The actual refusal/cancellation/manual-worker regressions pass under race. Internal
review approves these exact corrections. The final five-window preflight adds
100 completed steps in 42.81 seconds with clean teardown. The 30-window/600-step total
comprises 25 windows/500 steps before those error-path fixes and 5 windows/100 steps on
final source. [Resolutions](records/prism-resolutions.json) preserve the findings,
focused checks and version boundaries; no second provider review was run.
