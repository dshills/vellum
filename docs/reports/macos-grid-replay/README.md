# Stored managed-grid software replay

This checkpoint adds a regression and updates its contract. It changes no runtime
implementation or public API. The regression composes the actual managed grid,
inline text editor, application-owned history and two independent windows using
the explicit offscreen clock/router/software renderer.

The two datasets retain 10,000 and 1,000,000 rows with twenty uint32 fields per
row. A pilot authors one sixteen-event trace per dataset; fresh instances consume
that unchanged physical-coordinate trace at dispatch batches 1, 7 and 64. Ten
checkpoints compare domain revisions, active/selected cells, draft/editor state,
viewport, measured cells/headers, control semantics/focus and both software image
identities, including bounds and stride. The trace covers range selection,
two-axis wheel input, immutable index reversal, F2 editing, typed `42`, Enter
commit, Undo and Redo. Captured scrolled-out input and closed-window input refuse;
closing the grid preserves its inspector's completed image.

Independent oracles check row keys and values against the normal/reversed row
formula, require a deep post-reversal viewport, derive the undo baseline from
the domain record key and verify exact `42`/original/`42`, increasing cell/source
versions and journal counts. Internal review identified the original pilot-only
reverse/undo gap; these assertions resolve it.

The final two-source endpoint passes the complete adapter offscreen race suite
in 99.398 seconds of wrapper time (Go reports 91.106 seconds) and offscreen vet
in 3.272 seconds. An isolated Go overlay that drops wheel motion exits 1 in
6.984 seconds at the required-transition oracle, as expected. Source/prepared
hashes remain stable; all 6,216 prepared leaves are unchanged. Earlier successful
preliminary runs retain their narrower source scope; one preliminary race receipt
records documentation drift and is not the final gate evidence. This report is
written after the final code checks and reviewed separately.

Checkpoint record paths: [qualification record](records/qualification.json),
[raw evidence](records/evidence.tar.gz) and
[archive index](records/archive-index.json). All three are required before commit.
No desktop
window, native GPU, Cocoa input/service, clipboard or OS preference is exercised.
Live testing remains disabled while the human is using the computer.

The trace is authored once per dataset in each test process, rather than loaded
from an independently versioned large-workspace recording. Index reversal is
constant-time reordering, not a worker sort/filter workload. The full P5/P6
workspace with canvas/workers, P9 acceptance, target frame and memory budgets,
the failed native endurance campaign on base `2c87a3d` and overnight/P11 checks
remain open. This software regression does not diagnose the native failure or
qualify the complete Mac framework.
