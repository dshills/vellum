# Current managed-grid Mac performance matrix

The current Go grid implementation at `a2b4143` has a complete eight-case
stored-table timing matrix: CPU/GPU, fixed/variable heights and 10k/1M rows.
One configuration meets both unchanged targets in all five runs; the full
matrix still fails 4 ms p95. All forty runs meet 6 ms p99. No runtime change
or full Mac completion is claimed by this evidence checkpoint.

## Ordinary results

Each cell below is the median of five independently computed run percentiles,
not a pooled percentile. Every run contains 120 measured native FIFO scroll
events after 128 unmeasured closed-trajectory warm events. All 4,800 samples
and outliers are retained. At least 400 clipped cells, 50 row identities and
eight column identities remain required. Native render size is 1519×1266 at
162 DPI / DPR 1 on the 5120×1440 LG display.

| Mode / row heights | Stored rows | Median run p95 (ms) | Median run p99 (ms) | Runs meeting 4 ms p95 | Runs meeting 6 ms p99 |
| --- | ---: | ---: | ---: | ---: | ---: |
| CPU / fixed | 10,000 | 4.997499 | 5.267750 | 0/5 | 5/5 |
| GPU / fixed | 10,000 | 4.323542 | 4.564208 | 0/5 | 5/5 |
| CPU / fixed | 1,000,000 | 4.758751 | 4.970416 | 0/5 | 5/5 |
| GPU / fixed | 1,000,000 | 3.840415 | 4.073543 | 5/5 | 5/5 |
| CPU / variable | 10,000 | 4.905001 | 5.437711 | 0/5 | 5/5 |
| GPU / variable | 10,000 | 4.452875 | 4.774832 | 0/5 | 5/5 |
| CPU / variable | 1,000,000 | 4.846960 | 5.061792 | 0/5 | 5/5 |
| GPU / variable | 1,000,000 | 4.046042 | 4.296291 | 2/5 | 5/5 |

The fixed-height million-row GPU case passes both targets in all five runs.
The variable-height million-row GPU case passes p95 in two of five; every other
case passes it in zero of five. Overall, seven run-p95s pass and thirty-three
fail. All twenty ordinary 10k/1M paired geometry arrays match, as do four
separate diagnostic pairs. This qualifies matched visible work, not physical
input-to-presentation, the full tree/general-UI matrix or M4 resource budgets.

## Exact implementation and route evidence

Both calling checkouts are clean `a2b4143` during measurement, with 5,320 tracked
paths and 6,261 prepared stage0080 leaves. The retained private SDK remains
clean-source `abe67c6`, stage0079, with 7,879 payloads plus its manifest. The
optimized Go 1.25.6 normal executable uses the exact four committed grid runtime
files and the Core stage0080 admission helper through a compiler overlay. Its
separate admission test is excluded from the normal executable. The
full source diff, prepared vectors, whole SDK, compiler and binary hashes bind
that association. Other checkout changes since the SDK are documentation,
tests, packaging scripts and dependency patch metadata.

This is the current Go implementation over the unchanged SDK source-built
graphics archive `786aaad8…`, rather than qualification of the checkout default
native archive or a refreshed SDK. The ordinary executable contains no surface
observer. A separately instrumented executable proves 960 measured functional
frames across all eight cases, with 468–550 current layout values and nonnil
shapes matching each completed generation/frame, and zero native body widgets.
CPU final scenes show fresh raster/copy and checked RGBA upload attempts; GPU
final scenes show accepted fresh capture/encoding with five checked input writes,
positive input bytes and one submit attempt, without fallback or input errors.
Those diagnostics are excluded from ordinary timing.

## Preserved failure and corrected composition validation

The initial scope completed five validated diagnostics, then recorded a validation
failure on GPU variable-height 10k. All six applications exited zero with complete
logs, stable pins and retired/absent owned groups. The sixth ticket (generation
237, frame 239) includes an accepted unchanged intermediate scene, followed by
a fresh accepted final scene with five writes, zero errors, 553,264 input bytes
and one submit. The original validator incorrectly required every intermediate
composition to be fresh. The trace contract accumulates intermediate and armed
final compositions into one ticket, so this did not establish a rendering defect.

The original scope remains FAILED and sealed. Its sixth case lacks a completed
post-validation whole-vector/lease flag because the validator raised first.
A subsequent headless reconciliation verifies both complete current checkouts,
prepared trees and the SDK against the original plan. Independent review admits
revalidation of the six unchanged raw cases under a final-composition rule,
preserving their original statuses. No diagnostic is rerun or relabeled as a
new native execution.

The corrected check requires exactly one final scene source and a fresh final
CPU/GPU path, while retaining schema/order/ticket/geometry checks and requiring
all available GPU hosts to be accepted without fallback or input errors. Three
positive and twenty-seven malformed parser controls include a fresh prefix
followed by a reused final scene; that mutation reaches and fails the final
freshness check. Controls create no native evidence. A new reviewed scope runs
only the remaining two diagnostics, forty ordinary processes and four separate
allocation captures. All forty-six finish successfully; both scopes are sealed.

The diagnostic count is five original passes + one original failure later
revalidated + two new diagnostics = eight. The new scope is two diagnostics +
forty ordinary processes + four allocation captures = forty-six new processes.
Six prior native processes + forty-six new processes = fifty-two unique native
executions. Revalidation launches no application and adds no execution.

## Allocation attribution and next implementation

Four separate 120-event warmed captures compare original/current CPU and GPU
processes at fixed 10k, with the unchanged 524,288-byte Go memory-profile rate.
Before/after cumulative allocation profiles subtract startup and warmup. Twelve
offline pprof views and one source-line view retain exact profile hashes, logs
and process cleanup. These sampled process-wide costs include observers,
concurrent adapter work and profile encoding; they are not latency acceptance,
framework-only allocations, retained memory or native/GPU memory accounting.

The profiles display binary MiB values with an MB label. Sampled allocation
totals are CPU 420.19→458.98 MiB and GPU
588.34→668.46 MiB. Current `renderCellSurface` contributes 77.13 MiB flat on CPU
and 74.63 MiB on GPU. Source-line attribution places the CPU allocation at
`grid_surface.go:376`, the per-cell copy of the resolved paint style. This
supports investigating render-local style ownership; a compiler escape cause
or successful fix has not yet been demonstrated. Accessibility publication
remains a substantial shared baseline/current cost, and `syncCellSurface` also
allocates 39.42/33.93 MiB in these captures. Do not add cumulative stack totals
or equate sampled capture differences with ordinary per-frame percentages.

The next implementation should prepare the four paint-style classes once per
pane per render, preserving immutable per-operation lifetime, selected/cursor
backgrounds, ordering and unmodified resolved-style templates. It must add no
cross-frame cache or hidden widget substitutes. Exact original pixels/geometry,
template stability, style fallback and operation lifetime require regression
coverage, then an ordinary matched comparison with the existing adoption bar.
This is a proposed fix, not an adopted optimization.

## Review and remaining gates

Internal launch reviews cover the initial scope and corrected recovery; the
independent raw-results review confirms the statistics, geometry and cleanup.
One complete Anthropic Claude Sonnet 5.5 Prism call reviewed all four staged
documents (18,083 diff bytes, one chunk), with CLI exit 0, no retained findings,
one discarded medium count-clarity suggestion and no suppressed findings. The
discarded suggestion was still assessed and accepted as an editorial improvement:
the equations above make original executions, revalidation and new processes
explicit. Raw provider output is preserved; no second call is made. Final
internal closure and archive identities are indexed in `checkpoint.json`.

The archive retains both plans, original failure, revalidation, controls, all
52 unique native raw chains, sampled profiles and review dispositions. No native
case is restarted.

Tree/general-UI/canvas/layout/text/startup/input/pacing/resize, full memory and
endurance, native/human workflows, display/IME/VoiceOver, shipped-component
rights and final SDK/package/audit gates remain open. The original targets
and Mac completion scope are unchanged. No public push, release, tag, signing
or notarization occurs.
