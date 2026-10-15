# Composed Breadcrumb acceptance checkpoint

The combined Mac lifecycle fixture now includes the controlled Breadcrumb composite,
with its own row so each location has measured visible geometry. Each fresh window
owns its own path. The final two native selector requests reject navigation to
Workspace and accept navigation to Models. Exactly one owner callback is required
for each request. Completed native publication must show Models as static text and
remove the old Current leaf. Ordinary native objects may retain identity when their
stable managed ID changes role.

The tested runtime is base `b003d3759a9a121093f93f14656eec73b33daabd` plus the seven
source/spec changes recorded in the qualification, with 6,216 pinned prepared
leaves. The report was written after the run; its bytes are separately reviewed.
This is a source-delta checkpoint, not a claim that the native executable was built
from the future commit. The following long campaign must use a clean committed
endpoint and new evidence.

The complete adapter offscreen race suite passed in 90.425 seconds. Native-command
race tests passed, including acceptance with a reused object identity and refusal
of incomplete role/name/geometry publication. Model tests verify rejected requests,
accepted truncation, stale closure refusal and independent windows. Software layout
checks verify the breadcrumb items, 1,000 dense labels and selected color fields;
Extension content uses allocation placeholders in this software-only test.

Five actual Cocoa window cycles passed in 50.480 seconds: 26 steps per window,
130 attempts and admissions, no retry steps. Each window completed color editing,
tree selection/disclosure, text, tabs, modal open/close and breadcrumb requests.
All 1,000 dense labels were contained in the observed viewport each cycle. The
additional breadcrumb row changes available tree geometry: 15 completed native AX
rows were observed. This is distinct from the preceding fixture's 17-row geometry.
Queue admission is separate from accepted owner state and completed publication.

Tracked native cleanup ran on the main thread with zero outstanding tracked native
resources, accessibility windows and pending callbacks. A separate producer-bound
process observation found the recorded owned group absent. Menu restoration was
not qualified. These observations do not establish a native/GPU memory budget.

The first layout check failed because a breadcrumb nested in the existing toolbar
clipped Models. Moving it to a separate row fixed the measured layout. Internal
review also corrected an erroneous test expectation that role changes replace
native object identity, and tightened reviewed-harness stability across lock waits
and termination. The failed check and resolved review findings will be retained in the evidence
archive before commit.

The evidence archive will retain, before commit, source snapshots, exact prepared-input hashes, build
and test receipts, native action/result streams, producer scripts and reviews.
Executables are omitted with their digests retained. The spec's composed timeout
now matches the existing implementation: 10 seconds to three hours.

This checkpoint does not qualify final-source 1,000-cycle retention, overnight
soak, P4/P6/P11, sustained frame/input latency, native/GPU resource budgets,
physical input/display variants, human VoiceOver/IME, service interoperability or
Mac distribution. Those gates remain open.
