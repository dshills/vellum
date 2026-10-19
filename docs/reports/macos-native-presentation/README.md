# Optional native presentation recording

This checkpoint adds `cogent.NewPresentationMetrics` and
`Driver.PresentationMetrics` for actual Metal display feedback. It implements the
observation boundary for the owner's measured 60 Hz policy; it does not declare a
passing responsiveness result. See the [API/measurement contract](../../../specs/native-presentation.md).

The recorder shares a bounded receipt/drawable budget across window incarnations,
with explicit pending, presented, skipped, unavailable, invalid and retired states.
Native handlers retain only scalar subscription/sequence IDs. Retirement revokes
late callbacks before surface/view teardown. A Vellum-owned CAMetalLayer subclass
preserves super.nextDrawable, with observation disabled by default and no extra
acquire/present/readback/wait. macOS 10.15.4 availability is checked before enabling
observation. Initial Mac qualification remains limited to tested configurations.

Input receipts currently cover admitted grid/tree routed scrolls before FIFO
send, in the same native host clock as actual presentation. They need no render
lock or Scene/Stage traversal. Prefix/armed generation identity is captured at
composition start. Ordinary, measured and cached-resize compositions are marked.
A successful final FrameMetrics generation must still be joined for input
qualification; an armed tag by itself is insufficient.

## Checks and corrections

Checks used Go1.25.6, GOMAXPROCS=2, actual candidate Objective-C sources and the
retained source-built native archive `786aaad8…2ecec4883`. Build/test children had
IP networking denied, bounded output/timeouts and observed process-group cleanup.
They did not launch a native window, input operator or benchmark.

- Intermediate integration v1: all 607 top-level driver/recorder methods pass with
  race detection (309.367s). This is explicitly an intermediate-source result.
- Final affected revision: ten focused driver methods, six recorder methods and
  sixteen Core/registry methods pass with race detection. Default driver vet and
  actual Darwin native test-binary compilation pass. Core tests disable the
  existing package's unrelated unkeyed-literal vet diagnostics; no clean default
  Core-vet claim is made.
- The complete stage88 patch reproduces byte-for-byte in temporary dependencies:
  5,993 upstream and 292 typesetting files. No module-cache source is edited.
- The reusable native presentation fixture compiles the actual override with
  `-Wall -Wextra -Werror`; controlled drawable stand-ins check six original acquire
  calls, two registered handlers, one nil acquisition, refusal/disabled paths,
  immutable identity, zero drawable ID/time and late scalar callbacks.
- The three ownership fixtures pass: surface64 creations/retirements and local
  pool drainage, input100 cycles/901 factory calls, and AX100 cycles with1,000
  nodes each retired before outer pool drainage. They create no NSWindow.

Internal review caught independent input/drawable quotas that admitted twice the
promised shared limit. Both admission paths now test the combined count; both
ordering directions and mixed capacity are covered. The failed review and exact
pre-fix snapshot are retained privately. Review also led to explicit runtime API
availability and removal of receipt-path scene traversal.

The first pool invocation failed because its stale AX fixture lacked the current
visibility callback symbol. Adding that stub exposed a second stale assumption:
its NSObject window lacked the observer's initial `occlusionState` selector. The
fixture now supplies an inert window protocol and links the actual current
implementation. Both original failures remain recorded; the final three-fixture
run passes. No product accessibility behavior was relaxed. Existing AX deprecation
and duplicate objc linker warnings remain separate from the warning-free new
native presentation fixture.

[Evidence summaries](evidence.json) retain exact receipt/log hashes, tested method
names, command outcomes and process retirement observations. Full raw logs,
source-pinned receipts and pre-fix snapshots remain in the development checkpoint
`macos-native-presentation-cd6599e`; summaries do not reconstruct those raw files.

## Remaining qualification

No real drawable was acquired/presented by these fixtures. Live interception,
actual callback/host-clock correlation, pending/drop behavior, affected native
render/teardown acceptance and the five-run60Hz campaign remain open. The optional
recorder has no pacing pass/fail analyzer and does not certify physical latency.
Earlier native workspace/text/catalogue records remain bound to their original
source; this backend change requires affected renewal. The private SDK is not
rebuilt by this checkpoint. M2–M7 and full Mac completion remain open.
