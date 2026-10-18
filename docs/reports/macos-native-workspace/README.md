# Native professional workspace checkpoint

Both native CPU-scene cases pass the full immutable professional workspace
record: 216 operations and 27 captures at 10,000 rows, then the same counts at
one million rows (432 operations and 54 captures total). This closes the full
native workspace-record subgate. The native control catalogue and full M2 remain
open; M3–M7 and full Mac completion remain open.

## Implementation

The optional public `cogent.DesktopSession` supplies real originating FIFO input,
actual composition barriers and detached completed semantic/focus/grid/tree/
canvas inspection. Original WindowRefs, current frame capabilities and exact
qualified IDs protect admission. One request and at most 64 live windows are
retained. Undispatched input cancellation, actual terminal frame cancellation,
real window retirement and driver shutdown have explicit cleanup paths.

The dedicated tagged [native consumer](../../../examples/replay/cmd/workspace-native/README.md)
uses the unchanged records/goldens, ordinary pointer hit testing/capture, keyboard,
commands/history, worker results, docking and lifecycle. It verifies open/edit,
sort/filter/cancel, stable selection, tree changes, Unicode notes, tab reorder,
accepted placement preview and exactly one object, Undo/Redo, canvas navigation,
modal blocking, close-main and surviving-inspector state. It never invokes
application command callbacks as a substitute for routed input.

Native allocation differs from software. Named targets use completed bounds;
untargeted points use the original author's frozen-cell/disclosure/tab/canvas
anchors against each actual pane. Unknown/clipped anchors refuse. Only authored
Select All intents 17/122 translate from software Control+A to live-bound Mac
Command+A. No global keymap, OS input source, VoiceOver or preference changes.

An independent sequential grid reference constructs the stored row order from
immutable generated data and stable sort/filter rules, preserves the old row key
and intra-row offset on replacement, applies the authored logical wheel, then
clamps each viewport independently. It first reproduces software goldens and
never reads a candidate offset to derive its expectation. At 10k, native sorted/
scrolled/filtered Y values are 239546/239258/82058; at 1M they are
23980824/23980536/8247266. The 1M sorted state is above bottom and is not normalized
to bottom. Both native grid viewports are 623×482 versus software 624×480.
Other domain assertions remain exact; viewport sizes are independently validated
and camera centers retain the predeclared one-logical-pixel allowance. Actual
painted objects and active accepted placement preview are checked separately
from annotation counts. Workbench header focus reads actual ContainsFocus only
within the admitted visible/modal scope, independently of tab selection.

## Actual native results

| Stored rows | Applied / captures | App wall seconds | Session requests | Worker accepted/completed | Result |
| --- | --- | --- | --- | --- | --- |
| 10,000 | 216 / 27 |6.956295 | 431 | 7 / 7 | Pass |
| 1,000,000 | 216 / 27 |8.466642 | 431 | 7 / 7 | Pass |

Each process joins its scenario and worker, closes the executor/session with zero
pending/live references, retires tracked services/callbacks/AX/file panels/menu/
tray, and exits with its owned process group absent. Each emits 18 exact known
inherited editor Monospace diagnostic lines; unknown output remains a failure.
Request counts are not frame or presentation counts. Wall seconds are fixture
runtime, not a latency benchmark.

The bounded scope ran before its original launch cutoff under the shared native
lease, with IP networking denied, 240-second owned watchdogs and 16 MiB stream
limits. Whole checkout/prepared/retained-SDK boundaries remain unchanged during
execution. Both scope cases and all prior failed scopes are sealed.

## Source and checks

Qualification uses base `594bd37`, private stage0081 Core origin-Custom routing and
current Go overlays over the unchanged retained stage0079 SDK from `abe67c6`.
The compiler is Go 1.25.6 Darwin/ARM64; controlled `GOMAXPROCS=2` is explicit.
The SDK's source-built native archive is
`786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.
Native binary SHA256: `eefaf0babeb21781ff91b2fed49c6eab061f7604d201f539094c54cf6b700d56`.
The source worktree's default native library is a separate configuration;
compilation there does not inherit the SDK library's runtime qualification.
No SDK is refreshed or public release made.

Passing checks include the complete driver-root package (`.`) race 275.961s; focused desktop race
12.684s/10 methods; complete workspace race 167.740s, including all six software
record/batch combinations (1,296 operations/162 observations); seven native-oracle
methods 29.244s; driver and actual consumer vet; native consumer compilation; and
21 synthetic parser/result controls and all 11 affected preparation tests after
including stage81 in the pinned patch-order regression. Unchanged Core modal-routing race coverage
is reused by exact prepared endpoint correspondence. Actual stage81 preparation
passes in a private root and the adoption worktree: 6,263 prepared files, identical
whole vectors. Installed Go source correspondence and the separate source-worktree
native build are recorded in the bundle. Synthetic controls and software FIFO
routing tests are not native execution evidence.

The original native v2 case fails capture 41 because software Control+A is Mac
Home (`HarborNorth Star` versus `Harbor`). The corrected v4 case passes 41 then
fails 46 because diagnostic tab focus is missing. Both 10k failures remain failed,
with their 1M cases skipped. Unexecuted v3 preflight refuses an old binary hash.
Initial prototype compilation/fixture/vet failures are retained. A later focused
test used an encoded native ID in a Manual fixture and failed; correcting that
fixture produces focused-v12. Adoption correspondence initially included SDK-
excluded prototype modules, assumed a root go.sum and literal module equality;
final checks explicitly exclude those child modules, check absent sums, and
permit only the two exact SDK native-library replacement trailers.

## Limits and next gate

Native Go event routing has autonomous native scheduling, extra readiness frames
and a native wall-time wait. It is not deterministic Replay.Runner, OS/physical
pointer input, IME, VoiceOver, clipboard/services, presentation, performance,
full catalogue or cross-window palette drag evidence. Manual incarnation integers
and software image hashes are not repurposed as native capabilities or images.

Every terminal Cocoa snapshot still includes one inactive untitled
`NSKVONotifying_TUINSWindow`. Zero tracked resources and process absence do not
establish Cocoa deallocation or leak closure. M4 ownership classification remains
open. Native catalogue acceptance is next for M2; materially different measured
performance work and M3–M7 remain in the [delivery plan](../../../specs/macos-delivery.md).

Internal source/execution/results and staged checkpoint reviews, raw streams,
exact inputs, controls and Prism disposition are retained in the
[evidence archive](records/checkpoint.tar.gz), indexed by its adjacent manifest.

One Prism CLI invocation uses Anthropic Claude Sonnet 5.5 over all 27 staged files
and 131,736 bytes: five chunks/five LLM calls, complete coverage, no excluded or
skipped files, truncation or suppression. Its initial exit 1 reflects four medium
findings. Independent assessment leaves no runtime blocker:
input cancellation permits already dispatched effects; real driver retirement
releases pending requests; fatal unjoined shutdown intentionally emits incomplete diagnostics and avoids
potentially blocking lock-taking reads; readiness waits on actual FIFO/composition rather than an
immediate-success loop. All five provider-discarded findings are also assessed;
the M2 label receives an editorial clarification to “Partial.” No runtime Go
change or second provider invocation follows. The archive retains full rationale
and the unmodified provider response; this is not a claim Prism returned zero
findings. Internal review also restores the executable check-script mode, fixes
the preparation test's stage endpoint, and corrects test/evidence wording.
