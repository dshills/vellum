# Matched-rectangle source-cost result

The corrected one-process diagnostic is valid. **Image copy/upload API work is
the main source-draw CPU cost in this fixture.** This differs from the earlier
grid workload, where CPU rasterization dominated. No production performance fix
or ordinary latency qualification follows from this diagnostic.

| Per measured ticket | p50 | p95 | Maximum |
|---|---:|---:|---:|
| Source draw | 2.770 ms | 3.387 ms | 3.666 ms |
| CPU raster | 0.719 ms | 0.742 ms | 0.997 ms |
| Image copy | 2.074 ms | 2.659 ms | 3.046 ms |
| SetFromGoImage API | 2.065 ms | 2.650 ms | 3.037 ms |

Across the same 120 measured tickets, raster is 22.848% of summed source DrawNS,
copy 77.060%, and the nested image API 76.734%. These shares divide nested interval
sums by the same source-draw sum, not independent quantiles. Raster and copy are
subsets of source draw; image API is nested in copy. They are not additional UI
work, and phase percentiles must not be added.

Every measured ticket has one changed scene source, image 1519×1266, with one
SetFromGoImage attempt and no unchanged-image cache hit. Each logical RGBA input
is 7,692,216 bytes; 120 attempts total 923,065,920 logical bytes. These count attempted
input dimensions, not confirmed successful uploads or physical GPU transfers.
All 248 source records (128 warmup + 120 measured) pass the unchanged existing detailed
scalar validator: composition/source order, maximum 32 per ticket, nested interval
bounds and full-image logical attempt bytes. CPU scene is used; experimental
SceneHost is unavailable/inactive. No prefix composition is discarded.

Execution on 21 October 2026, 21:43:24.189–21:43:57.021 UTC
(16:43:24–16:43:57 Chicago), lasts 32.833 seconds within 60 seconds. The native and
eight configuration-probe groups terminate/retire. Foreground, complete input/
model/presentation accounting, callback closure, source/binary/overlay/match/log
identities, unchanged SDK 8,321-file inventory, original 786aa native archive,
configuration and persistent geometry all pass. The original source/native files
inside the retained SDK are untouched; no Rust rebuild or renderer/pacing/default
or public API change occurs. Extra detailed clocks perturb elapsed time, so
ordinary latency/release acceptance is explicitly false.

## Diagnostic repair and evidence

The first 36ad689 attempt is sealed/failed at its first input after 1.237 seconds,
with no complete cohort: detailed BeginGroup omitted the per-ticket source-detail
container already initialized by Begin. Its raw output, source snapshot and
[refusal](refusal-analysis.json) remain retained and unaccepted.

Stage 0106 in 29f6cf4 adds the same conditional initialization under the existing
trace mutex. Ordinary traces keep nil detail storage. Two new Core regression
methods exercise detached prefix/final scalars, the shared 32-record budget,
33-record refusal, sticky cancellation with in-flight composition and ordinary
isolation. They fail by assertions on original 105 and pass with existing group/
source-trace lifecycle tests under race detection. All 34 preparation controls
and a full fresh prepare pass. The initial range/last-stage and asset-stage lookup
test failures are preserved. This repairs diagnostics, not display latency.

The fresh build overlays three exact existing-code locations: detailed trace
constructor, detached result field, and identical grouped initialization repair.
[Original](build-identity.json) and [corrected](build-identity-v2.json) identities
bind those bytes. Five new wrapper methods with six nested negative subcases pass;
qualified view/routing/pixel and existing parent/scalar-validator tests are reused.
Source preparation/fix receives internal review and Prism Anthropic Claude Sonnet
5.5 before execution. The old final classifier's informational string still says
"two" overlays; its pinned JSON and source snapshots bind all three locations.
Summary metrics are copied from the frozen validation and deliberately are not
regenerated with the updated classifier; metrics_producer identifies its saved
source and SHA. Only that wording is corrected afterward, without changing raw output, numeric
validation or acceptance. Result verification executes the exact saved classifier
with its original source-relative import location.

The next boundary is inside the original image-upload call: distinguish buffer
allocation/mapping, CPU data copy and queue/encoding work before selecting a repair.
The current Go binding borrows the RGBA byte slice directly; do not assume another
Go-to-C image copy. The source-cost interval itself does not isolate native
allocation versus copy. Avoid speculative raster optimizations or unchanged
presentation retries on this evidence.

Features remain frozen, the 33.3 ms receipt-to-actual-presentation gate is unchanged,
and M3–M7 remain open. Sparse detailed diagnostic data do not qualify 60 Hz changing
professional workloads, physical input/pixel readback, resize, endurance,
allocation balance or distribution. See the [protocol](../../../tools/macos-calibration/sourcecost/PLAN.md),
[exact summary](summary.json) and indexed evidence archives. Executables, native
libraries, SDK/cache/target trees, credentials and prior archives are excluded.
