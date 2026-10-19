# Fixed-deadline 60 Hz presentation workload

The stored grid/tree consumers provide an opt-in `-perf-60hz` mode with a separate
version 2 presentation journal. It prepares a continuously changing workload for
the [Mac release policy](macos-responsiveness.md); implementation and synthetic
validation do not establish a native display result or a passing Mac gate.

## Select the mode

Build optimized native binaries using the examples' existing instructions. For a
quiet, explicitly available desktop session, select a new journal for each run:

```sh
/tmp/vellum-loadedgrid -perf -rows=10000 -perf-samples=600 -perf-60hz \
  -perf-presentation=/private/tmp/new-grid-60hz.json
/tmp/vellum-loadedtree -perf -nodes=100000 -perf-samples=600 -perf-60hz \
  -perf-presentation=/private/tmp/new-tree-60hz.json
python3 -B scripts/check-presentation.py /private/tmp/new-grid-60hz.json \
  --sha256 RAW_JOURNAL_SHA256 --source CLEAN_SOURCE_COMMIT \
  --binary-sha256 EXECUTED_BINARY_SHA256
```

These are runnable instructions, not a record of a native run at this checkpoint.
The mode requires `-perf`, a new presentation file and 600..2048 states. At 60 Hz,
600 states have a 10-second planned cohort. Keep the existing complete warmup
trajectory. Grid supports fixed/variable 10k/1M stored rows and CPU/GPU-requested
routes; tree supports wide/deep 10k/100k nodes and both requested routes. Novel
one-way grid scrolling and other source/profile/input diagnostics are excluded.
The requested GPU flag still does not prove actual GPU routing.

Ordinary and serial modes retain their branches and defaults. The new mode emits
explicit `paced_workload:true`, `scheduled_hz:60`, presentation diagnostic and
nonqualification labels. It emits its completion summary after the timed cohort.
Version 1 serial journals retain their format. Both modes add explicit prefix
observation counts as described in the shared capture contract.

## Scheduling, overload and observation

Every slot exists before the origin. Slot i has the immutable deadline
`origin + floor(i * 1,000,000,000 / 60)` nanoseconds. No deadline is rebased after
work, no ticker delivery count replaces a slot, and no late slot is skipped.
Exactly one request runs until its CPU acknowledgement and generation-bound
inspection complete. An overdue slot starts after its predecessor: this is
retained backlog, not recovered on-time cadence. Busy/refused input fails the
run. Cancellation/failure retains failed and unstarted slot identities.

Scheduling uses Go's monotonic clock. Each slot records scheduled/start,
pre-owner dispatch, enqueue-return, CPU acknowledgement observation, inspection
end and iteration end times. The 1 ms Poll and owner probes remain timed observer
costs; CPU completion and its observation are distinct. There is no wait for
actual presentation between requests. The existing callback, immutable stored
value, exact offset/generation/frame, clipped workload and geometry checks stay
active. No-op or clamped changes fail the run.

Setup, explicit GC, output and plan allocation precede the origin. No per-event
stdout emission, ReadMemStats or profile capture runs in the cohort, including
backlog draining. Natural GC stays enabled. Geometry hashing and validation
allocations remain included observer work; their cost is not subtracted. Retain
compact detached initial/slot view summaries, not full mutable cell/row histories.
The plan is bounded at 2048 slots and its encoded evidence at 2 MiB; the existing
journal's 64 MiB, recorder's shared 65536 records and two-second drain limits remain.
The paced scheduler has a two-minute cooperative context limit; the enclosing
consumer/process watchdog remains responsible for a broken blocking runtime.

The planned end stays fixed. If CPU work finishes early, the last state is held
until that end; an overrun remains recorded. This hold is an idle tail and adds
no changing frame. After the cohort, freeze the current acquisition prefix and
require exact receipts/final armed identities before draining its existing
callbacks. No later acquisition repairs a missing final identity. Worker/runtime
acknowledgements and post-close evidence follow the serial capture contract.
The shared [startup observation rule](presentation-capture.md#explicit-pre-input-startup-observations)
retains precisely proven pre-input unarmed skipped callbacks as non-timing
observations. The complete prefix stays frozen and visible; all other prefix loss
and every missing final armed presentation still fail. Startup records never
change the N final-state population, N-1 intervals, latency or cadence thresholds.

## Native timing and denominator

Validation joins every scheduled state to its exact completed CPU generation,
routed receipt and single final armed actual presentation. The primary cadence
population is the N final states in request order, giving exactly N-1 internal
intervals. There is no warmup presentation anchor in this population. First-state
latency and final overrun remain visible. Intermediate/untagged acquisitions stay
in raw diagnostics and cannot enlarge the primary denominator. Same-time or
reordered final presentations are reported as collapsed/ambiguous states and
cannot pass the clean ordered-visibility predicate. Zero intervals never count
as successful updates.

Report interval p50/p95/p99/max, exact passing count/denominator, every interval
above 18.333 ms and every above two periods, raw receipt-to-presentation latencies,
planned/observed duration, start lateness, deadline misses and worst/consecutive
backlog. The numeric predicates retain the release policy's 99% within 1.1 period
and p95 latency<=33.3ms. They are reported separately from execution qualification.
An average or a short successful prefix cannot hide a delayed cohort.

Native receipt and presentation use one host-clock domain. To report scheduled
endpoint bounds, intersect the native receipt's enclosing Go dispatch/return
brackets. The model assumes a constant origin offset and common nanosecond rate
for the cohort; it is not independent clock calibration. The pinned Go 1.25.6
Darwin ARM64 path and CACurrentMediaTime both derive from mach_absolute_time.
Conservative float/integer rounding allowances apply to receipt-origin bounds
and independently to first/final presentation endpoints. Impossible bracket
intersections refuse validation. No arbitrary Go/native timestamp subtraction
is used. Actual native clock delivery and this correlation still need live proof.

## Qualification remains separate

`VALID_PACED_NATIVE_FEEDBACK_JOURNAL` means these bytes satisfy the strict
schema/identity/schedule checks. The report always keeps `cadence_qualified:false`,
`physical_latency_qualified:false` and `release_qualified:false`. A numeric
`cadence_threshold_met` or `input_latency_threshold_met` does not authenticate
execution. Require the external source/build/native-library/binary/launch/raw
file binding, actual display/refresh/scale/power record and real callback/lifecycle
proof. Five measured runs of the declared representative workloads remain
required. Grid/tree do not cover composed workspace, text/editing, canvas or
resize; those release workloads remain open. [Checkpoint evidence](../docs/reports/macos-paced-presentation/README.md)
describes headless verification only.
