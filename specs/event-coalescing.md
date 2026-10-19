# Event queue coalescing contract

`events.Source` and `events.Deque` remain independently consumable Go APIs. This
checkpoint fixes queue-side aggregation beyond the desktop adapter's per-sample
checks. It keeps normal adjacent compression and introduces no native input or
cross-window API.

## FIFO aggregation

Back/FIFO compression accepts only eligible adjacent events. Both older and
newer events must be non-Unique and non-Deferred, and `IsSame` must hold. No
ordinary event may erase a unique event or a deferred reservation/publication.
Only the existing deferred `Ready`/`Merge` protocol combines deferred FIFO nodes.

Compute an entire proposed movement/scroll merge before replacing the tail or
changing incoming metadata. If aggregation is unsafe, append the newer event and
retain the older event in FIFO order. Do not saturate, clamp, drop either part,
partially add an axis, or alter a refused sample's position/time metadata.

Scroll input components and their float32 sums must be finite for aggregation.
Individually finite parts whose sum would overflow remain separate. Existing
float32 rounding is retained for ordinary representable sums; exact arbitrary
precision accumulated distance is not promised. Nonfinite caller-supplied
samples remain separate; the queue does not sanitize otherwise invalid input.

MouseMove/MouseDrag retain the newest location, generation/start times, drag
anchor and local offset, and inherit the earliest previous position/time. The
raw `Prev` and `PrevLocal` snapshots must describe the same origin under the
newest event's offset so `PrevDelta` agrees before and after normal localization.
The compressed raw delta, new previous-local coordinate and current local delta
must fit `image.Point` component integers before merging. Incompatible custom
payloads with movement/scroll type identifiers remain separate instead of being
cast destructively. Initial input validity remains the producer's responsibility;
this does not validate arbitrary malformed fields or future caller offsets.

## Priority, ownership and cost

Front/priority retains its existing latest-notification replacement policy,
normally used for window paint/resize. It does not newly aggregate pointer data.
It also honors both-side Unique/Deferred barriers. Front remains LIFO with
priority over Back; this does not introduce deferred-readiness gating in Front.
Existing `SendDeferred`, `UpdateDeferred` and `nextBack` protocols are unchanged.

`Send` and `SendFirst` transfer event ownership. Producers must not mutate an
event after publication; consumers own events after removal. Queue locking and
condition signalling retain their existing semantics. Admission/merge work is
constant time. Normal movement/scroll bursts still compress; unsafe samples may
require more nodes. The queue remains an infinitely buffered API and does not
promise bounded length for arbitrary producers or consumers that stop draining.

## Qualification

Automated controls cover finite scroll overflow/refusal atomicity, signed origin
bounds, nonzero local offsets and virtual times, unique/deferred barriers,
priority/latest behavior, custom payload fallback and concurrent transfer. A
big-integer oracle checks native-integer subtraction. Original-code overlay
controls must fail for the named bugs before accepting positive evidence.

Fresh pinned dependency preparation, events/driver/core race tests, vet and
private SDK build/relocation/package checks are source/offscreen evidence.
They do not qualify physical pointer behavior, OS input/accessibility, actual
presentation, native allocation/idle/endurance or measured display performance.
Mac release requires measured 60 Hz responsiveness; 4 ms is an optimization
target. Remaining Mac M2–M7 acceptance and distribution gates stay open.

The changed events and desktop/base driver packages pass default Go vet. Broader
Core vet still reports twelve inherited diagnostics, verified byte-for-byte
against the committed baseline. This checkpoint does not waive those diagnostics
or claim a full Core vet pass; their cleanup remains part of the final source
quality audit. No analyzer was disabled.
