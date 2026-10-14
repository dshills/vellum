# Tree and retained-view geometry publication

Managed trees now refuse completed-frame inspection, hit coordinates and queued
accessibility actions during a full scene layout, even when the portable model
generation is unchanged. Completed geometry retains an atomic originating scene
reference and checks that scene's full-layout fence. Same-generation pending-index
metadata is also checked. The last completed geometry remains available privately
for read-only semantic retention.

Outlines now use the existing bounded virtual-view pending/completed transaction.
The legacy Grid-prefixed wire/inspection fields keep their names and apply to both
managed tables and outlines. Owner snapshots stamp every host/row/page with one
binding incarnation, target/completed generation, pending state and copied frame.
While layout is pending, the bridge retains last-completed data/frames but revokes
actions, enabled/focused state and owner handlers. Cocoa selectors independently
refuse pending actions. Completed-frame publication restores current semantics and
frames atomically while preserving compatible object identities.

The completion comparer includes all TreeInspection data/geometry except frame
ordinal. A full layout schedules owner publication even if final bounds/generation
are unchanged. Source stamps, viewport, scale, selection, pending metadata,
row/disclosure/page frames and retained cardinalities remain separate provenance.
A scoped host-role transition is a cache/native ownership boundary; unrelated
ordinary controls and independent view scopes retain their existing behavior.

Internal review found retained scene transfers that bypass each descendant's
reparent path. Recursive withdrawal now precedes modal preparation, native modal
close, programmatic close, keyed ancestor parking and workbench parking. It
revokes both Grid and Tree completion and captured scene references; Grid's local
layout stamp becomes invalid/pending. The scoped read-only AX cache is separate.
Retained widgets/models remain the same; only a destination frame restores geometry.

The original manual regression returns stale inspection/hit geometry and accepts
an old AX action during full layout. It fails before the fix. Additional inert
sentinel regressions fail before modal-transfer/workbench-parking withdrawal;
they exercise actual transfer/reconciliation and the real panel.View callback,
without claiming that a seeded sentinel is a native frame. Final race checks cover
restored publication, all data/geometry comparator fields, scope-kind changes,
retained ancestor moves and both modal close paths.

A Cocoa production transaction fixture checks same-generation pending/ready
identity, readable retained frames, independence from ordinary geometry callbacks,
selector defenses, binding replacement, real removal, scope-kind replacement and
terminal retirement. It creates actual Cocoa accessibility objects without OS
windows and does not certify human screen-reader behavior.

The final framework passes the complete Cogent adapter offscreen race suite
(88.65 seconds), adapter vet and the loaded-tree consumer race suite. Actual
Cocoa transaction fixtures run within that adapter suite. Prepared dependency
inputs remain byte-identical: 6204 files, no new upstream patch.

The stored 100k-tree native workflow completes all nine stages, including actual
filter settlement and independent surviving-window behavior. Its consumer does
not join native teardown, so that run supplies workflow evidence only. Wide and
deep tree preflights each complete 120 ordinary scroll samples with 71 clipped
rows and exactly one tree notification per sample. Their p95 total UI work is
8.953 and 9.036 ms respectively, above the 4 ms target. These are single-process
compatibility checks, without a matched speed comparison.

Three five-cycle composed runs pass on the fixed framework. The final acceptance
command also awaits an acknowledged owner inspection and matching completed
native outline transaction: host/row scope, incarnation, completed and target
generation, nonempty frames, label, selection, disclosure, full index range and
visible row count. Duplicate native DomainIDs refuse even when their native
IDs/object identities differ. Each of its five cycles observes 18 completed AX
rows, 1000 fully visible density labels, accepted native text, editor history,
Grid/Tree geometry and modal open/close. Terminal cleanup runs on the main thread
and observes zero native objects, AX windows, window titles and pending callbacks.
The last forced-GC sample is 42,339,912 Go heap bytes and 337,199,104 RSS bytes;
five cycles do not qualify steady-state retention, a 1000-cycle campaign or native
allocation/GPU-memory attribution.

Version boundaries are explicit: the full framework suite and initial native
preflights precede only the acceptance-command extension. The final command's
race test, build and composed run cover that extension. Framework source remains
unchanged across these boundaries. Internal review approves the framework and
final checker correction; the checkpoint receives one Prism Anthropic Claude
Sonnet 5.5 review before commit.

[Validated summary](records/final-summary.json) records the final receipts and
scope. [Qualification](records/qualification.json) and the
[evidence index](records/evidence-index.json) bind the raw archive to reviewed
source, producers, test failures, version boundaries, internal review and Prism
result. Native binaries are excluded; their hashes remain in process receipts.

Prism covered all 19 selected files in four provider calls and reported one medium
finding (raw exit 1): the checker might reject a native row for an empty-bounds
owner row. Production explicitly omits that row. Internal review confirms exact
set equality is intentional; accepting extra native rows would weaken the check.
Explicit positive omitted-clipped-row and negative extra-native-clipped-row
fixtures pass under race. The raw finding remains preserved, with its
[reviewed disposition](records/prism-disposition.json); no second provider review
was run. That follow-up changes tests and reporting only.

The first new test used the wrong snapshot helper name and failed compilation;
its corrected original-source failing regression supplies the oracle. An early
programmatic-close edit missed the direct path; the regression caught it, and
that failure is preserved alongside the fix. The first full offscreen suite
passes the driver but is denied the workbench test's owned Unix socket; an
explicitly admitted complete rerun passes. No failed run is promoted as evidence.

This checkpoint addresses current geometry and retained-view accessibility
ownership. It does not close full P6/P11 latency, native/GPU memory, physical
input/displays, overnight soak, human accessibility or Mac distribution gates.
