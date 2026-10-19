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
