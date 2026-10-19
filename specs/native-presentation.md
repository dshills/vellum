# Native presentation recording

This optional macOS diagnostic implements the observation boundary needed by the
[60 Hz release policy](macos-responsiveness.md). It does not itself qualify that
policy. It is disabled unless an application supplies `Driver.PresentationMetrics`.

## Public use

```go
frames, err := cogent.NewFrameMetrics(1024)
if err != nil { return err }
presentations, err := cogent.NewPresentationMetrics(8192)
if err != nil { return err }
driver := cogent.Driver{FrameMetrics: frames, PresentationMetrics: presentations}
// Run the existing application with driver. Retain detached snapshots while it
// runs and after shutdown; do not interpret pending/retired entries as displayed.
```

`PresentationMetrics.Snapshot` returns detached scopes, input receipts and
acquisition/presentation records. It holds no application/native handles and
cannot be used to inject successful feedback. Copying the metrics wrapper makes
the copy invalid. Storage admits at most 64 total window incarnations and one
shared capacity of 1..65536 receipts plus drawable records. Exhaustion fails the
diagnostic; it does not evict earlier evidence. These are count bounds, not exact
allocator/process-memory limits. Plan capacity before a run. Observation is not
implicitly enabled for a long endurance campaign.

Each scope distinguishes window incarnations even if their application IDs are
the same. Acquisition sequences and backend subscription tokens are not reused.
A sample contains its scope, sequence, input tag, captured owner generation,
composition ordinal, armed flag, native acquisition time, drawable ID, actual
presentation time and state. Drawable ID zero is valid; do not key a cross-window
join by drawable ID alone.

## Timing and correlation

The Vellum-owned layer preserves super.nextDrawable and registers one Metal
presented handler only after reserving a sample. Its callback receives the
actual drawable's `presentedTime`; the time of a Present call is not substituted.
Receipt and acquisition use `CACurrentMediaTime`, in the same host-time domain.
Installation refuses unsupported systems before enabling observation (macOS
10.15.4 minimum). Other desktop platforms refuse installation.

For now, receipts cover admitted, initialized `EnqueueGridScroll` and
`EnqueueTreeScroll` events after validation/reservation and immediately before
FIFO enqueue. They include later queue wait, dispatch, owner work, reconciliation,
composition/acquisition and presentation. The receipt clock requires no render
lock or Scene/Stage traversal. This boundary does not include earlier application
request preparation and is not physical NSEvent/input-to-photon latency. Declare
the method for every campaign. Other workflows need an explicitly defined input
boundary before they can qualify input latency.

Generation and armed state are captured when tracing starts a composition. A
prefix frame cannot inherit a later owner transition. A backend ordinal brackets
ordinary, measured and cached-resize composition under the existing window lock.
A matching tag alone is insufficient: input qualification must join a successful,
completed FrameMetrics sample with the same scope/tag/generation and an armed,
actually presented composition. Compare receipt <= acquisition <= presentation.
Retain and inspect intermediate compositions instead of picking the first tagged
drawable. A successful armed stamp does not replace submission/error validation.

The later measurement campaign must retain raw samples, sample counts and all
intervals in its declared continuously changing phase. Pending, missing, retired,
unavailable or invalid feedback cannot disappear from the denominator. This API
has no built-in pacing pass/fail decision or physical latency certification.

## Loss and lifetime

Records distinguish pending, presented, skipped (zero presented time), unavailable
(nil drawable), invalid and retired. Nonfinite/backward clocks fail the recorder;
nonfinite values do not break diagnostic JSON. A drawable that never calls back
remains pending until its scope retires. Closing a window revokes its subscription
before native surface/view teardown. Already admitted callbacks independently
recheck recorder scope authority. Pending entries retire without manufacturing a
presentation. Snapshots remain available after close.

Native handlers capture only scalar token/sequence IDs, with no layer, window,
Go pointer or application reference. Registry/native locks are released before
sink callbacks. The bounded sink performs no window/main-thread work. The layer's
original borrowed-pointer/view ownership contract and autoreleased drawable
return lifetime are preserved.

## Verification status

[Checkpoint evidence](../docs/reports/macos-native-presentation/README.md) covers
headless accounting, late callbacks, actual bridge compilation and controlled
native drawable stand-ins. Those tests create no native window or real drawable.
Live interception, actual Metal delivery/clock correlation, pacing, affected
native lifecycle acceptance and the full Mac release gate remain open.

## Consumer preflight

The [serial capture checkpoint](../docs/reports/macos-presentation-capture/README.md)
adds opt-in stored grid/tree journals and a strict offline validator. Race tests,
vet, native builds and synthetic journal controls pass. No real native feedback
was captured here; serial scroll intervals do not qualify sustained 60 Hz pacing.
Callback/clock/lifecycle proof and a continuously changing campaign remain open.

See the [capture contract](presentation-capture.md) for flags, bounds and provenance.

## Fixed-deadline workload

The [paced consumer contract](paced-presentation.md) adds a separate opt-in 60 Hz
grid/tree workload and version 2 journal. It retains fixed deadlines, all backlog
and a final-state-only interval denominator. Headless checks do not establish
live callback delivery or close the measured Mac gate.

## Opt-in desktop routed-input transactions

`NewMeasuredDesktopSession(capacity)` adds bounded routed-Go-input measurement
without changing `DesktopSession.Input` or `Frame`. Install the session together
with `Driver.PresentationMetrics`; `FrameMetrics` remains mutually exclusive.
`MeasureInput(ctx, ref, input, expectation)` validates the current originating
frame/target, reserves the original initialized event group and records the
native host-time receipt immediately before publishing it to the native FIFO.
The preliminary session request/target-validation queue is outside that receipt
boundary. Go monotonic `ReceiptStartedAt`/`ReceiptFinishedAt` bracket the native
clock read; capture consumers must subtract their monotonic origin before JSON
serialization. Wall-clock or `time.Time` JSON conversion loses that evidence.
It does not call `Frame` or force an artificial response paint.

The event group remains unarmed through actual dispatch and owner work. The
originating owner drain and changed-view reconciliation must finish before the
expectation runs at most once on the UI owner (earlier failures skip it). That observational callback checks an
independently expected change against an application baseline, rejects no-ops,
refused or unrelated changes, and returns a positive application/model revision.
The framework cannot prove an arbitrary callback's honesty. Callbacks must not
mutate, wait, enter blocking native/session APIs or retain native objects.
An incomplete owner drain, invalid lifetime or missing changed reconciliation
fails the measurement without arming a composition.

The request tag, actual reconciled window-view generation and accepted
application revision are distinct. `DesktopInputMeasurement` retains the native
incarnation `Scope`, CPU `Frame`, `ViewRevisionBefore` and `AcceptedRevision`.
Join scope/tag/view generation to actual native acquisition/composition feedback;
a completed CPU composition alone does not prove display presentation. Earlier
or in-flight compositions retain their unarmed identity. A new view invalidation
after arming fails the pending observation rather than labeling newer state with
the older view generation; already completed historical frames remain immutable. The accepted revision
never substitutes for native frame generation. Detached history retains no window
references or expectation closures. A routed
cancellation keeps its bounded request/window reservation only until the native
input/owner terminal or window retirement.

Capacity1..65536 counts every admitted measured request, including failures and
cancellations, without eviction. Only one session request is pending. Original
input groups contain at most four events. Cancellation before routing removes
the queued request; after routing it fails measurement but cannot retract the
event group's possible effects. The trace stays reserved until actual input and
owner terminals, or window retirement. Nonvisual setup inputs use ordinary
`Input`; expected visual updates cannot be reclassified as no-ops to improve a
measurement population.

This checkpoint supplies the correlation capability and headless regression
evidence only. Software composition must refuse native timing success;
controlled Core composition terminals are not live Metal feedback. The existing
workspace measurement consumer/journal schema, managed Cocoa-input and AX method
adapters, actual resize callback correlation and five-run representative60Hz
campaign remain separate work. No M3 or full Mac release gate is closed here.

Grouped CPU diagnostics accumulate actual originating-window owner drains and
reconciliation after the first original event has dispatched, including work
between the original events. Those accounting tokens do not close the input or
owner terminal: the full original group and a following drain must still finish
before checked arming. Admission/setup work before original dispatch is excluded.
`OwnerNS` measures the observed owner-drain intervals, which can also contain
other admitted app work from the same batch; it is not exclusive per-callback
CPU attribution. Synchronous owner work inside original input dispatch is already
part of `InputNS` and is not counted twice. `TotalUIWorkNS` sums these observed
elapsed stages; the receipt-to-presentation host-clock latency independently
includes all intervening queue waits and elapsed work.

Unarmed compositions after the first original dispatch contribute their observed
render/compose stages too. Their acquisition identity remains unarmed; they do
not become the final response by later acceptance. A known earlier trace failure
blocks the state-observation callback. Failed grouped records retain their final
observed input/owner stage sum even when failure preceded the remaining terminal
work. Drawable acquisition/present-call waits remain separate counters rather
than additions to `TotalUIWorkNS`, consistent with existing frame diagnostics.
