# Current SDK stored-tree cost boundary

The unchanged minimum-Go private SDK from clean `abe67c6` completes four fresh
stored-tree observations on this Mac: CPU and GPU-requested ordinary runs, then
two separate source-detail diagnostics. They use the already compiled Go 1.25.6
release executable with stage0079. No production, prepared dependency, SDK or
executable is changed. Both calling checkouts are clean at `d47e166` at terminal
observation; their changes since the SDK source are documentation only.

Both ordinary runs fail the unchanged **4 ms p95** target. These are one process
per mode, useful for identifying the current cost boundary, not the five-run
10k/100k wide/deep acceptance matrix or a causal renderer comparison. Diagnostic
timings are excluded from ordinary results. The earlier [stored-tree renderer investigation](../macos-tree-gpu/README.md)
retains its rejected scanline prototype and historical measurements separately.

## Ordinary observations

Each ordinary run has 128 closed warmup events and 240 measured native scroll
events. Each diagnostic run has the same 128 warmup events and 120 measured
events.
The source fixture requires one owner callback, matching generation, requested
offset, current stamp and completed frame for every measured event. The external
validator additionally checks exact event/ticket counts, integer phase accounting,
per-event viewport, clipped workload and repeated-phase layout hashes. There are
480 ordinary and 240 separate diagnostic samples. All four processes exit zero,
reach complete log EOF with empty stderr and have their owned groups observed
absent. Every completed case is sealed; the shared live lease remains held through
owned cleanup and terminal input checks.

| Ordinary request mode | p50 ms | p95 ms | p99 ms | Maximum ms | Mean ms | p95 target |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| CPU | 5.822125 | 7.183709 | 7.759875 | 9.043041 | 5.889203 | Fail |
| GPU scene | 5.799749 | 7.217669 | 8.199166 | 9.035626 | 5.886423 | Fail |

Coefficient of variation across events is 0.1322/0.1192 respectively; this is not
between-run confidence. Median process allocation is 2,635,040/4,656,520 bytes
and 22,415/30,864 allocations per measured event. It includes observers and other
process goroutines, and does not establish framework-only bytes or baseline
allocation acceptance. No sample or failed limit is dropped.

All four runs match phase geometry: 68–69 nonempty clipped rows with at most 500
materialized/retained rows, the same bounds, scale, normalized row rectangles and
local font bytes. The diagnostic scene image is 1519×1114 pixels. Requested window
sizes do not qualify physical-display variants or the original maximum-visible
workload. There is no comparison against older results with different geometry.

Mean ordinary elapsed phases, measured in each ordinary process:

| Phase | CPU ms | GPU-requested ms |
| --- | ---: | ---: |
| Reconciliation | 1.493 | 1.669 |
| Layout/paint | 1.553 | 1.758 |
| Raster/upload or scene-source draw | 2.708 | 2.331 |
| Encode/submission | 0.125 | 0.118 |

The remaining input/owner/gather regions complete the per-frame sum. These are
elapsed CPU-side regions including pauses, not OS-thread CPU or GPU execution.
Queue, acquisition and Present-call observations remain separate. Independent
phase percentiles are never added. The requested modes and sequential order do
not establish a causal renderer-speed or thermal comparison.

## Separate route diagnostics

The CPU diagnostic records 120 scene sources and 120 image API attempts, each
with 6,768,664 logical RGBA bytes: 812,239,680 bytes in total. Mean source draw is
3.015 ms, with 2.167 ms raster and 0.848 ms copy intervals in that same diagnostic
process. API bytes are not physical transfer or live GPU memory.

The GPU diagnostic records 120 fresh accepted scene-host draws, zero fallback,
zero reuse and zero image API uploads. Mean host capture, encode and lease
intervals are 1.287, 0.498 and 0.319 ms. Resource work (0.314 ms) contains input
work (0.205 ms) and is nested inside lease; these nested values must not be added.
The source draw mean is 2.107 ms in that diagnostic process. These observations
establish actual route admission for the diagnostic corpus, not GPU completion,
compositor presentation or every ordinary frame. No diagnostic value qualifies
ordinary latency or is subtracted from another process's totals.

## Next bounded implementation

Current `syncTree` assigns native rows by visible index. A scroll changes the
projection and calls the recursive host `Update`, which revisits retained row/text
styling and full scene layout. The grid already has domain-key native retention
and an explicitly fixed-allocation local-layout root. The next private prototype
will apply that existing ownership/layout boundary to the tree:

1. Keep only the current admitted projection, at most 500 rows. Retain native
   rows by stable domain key, consume departing rows once as donors for entering
   keys, and retire surplus rows. This is not an offscreen row/paint cache.
2. Preserve unchanged text/style within overlapping rows. Use Core's existing
   local-layout and passive-text update APIs only when the actual originating
   scene, parent, allocation, resolved style, scale, row configuration and accepted
   immutable index match. Equal source stamps/counts alone cannot prove index
   identity; carry an opaque comparable non-retaining index identity in the same
   snapshot or take the full fallback. Refresh each row's current OutlineRef,
   epoch and AX authority on every generation; retained widgets confer no stale
   input authority. Preserve fractional-DPI sizing and queued-text cancellation;
   key identity alone does not prove unchanged intrinsic sizing.
   Changed content/state receives the required update and sizing; moves do not
   silently waive those inputs. Keep complete scene repaint in painter order.
3. Initially admit viewport-only changes. Pending, selection, topology, uncertain,
   style/theme/font/DPI/width and ownership changes use the full existing updater/
   layout route. Footer text, pager disabled states and inherited selected/disabled
   row styles and scrollbar state are part of the dirty/fallback proof, not just
   row labels.
   Preserve completed geometry
   publication, pending/accessibility fences, selection/disclosure, clipping,
   modal/disabled/stale refusal, callback counts and retirement.
4. Compare against a frozen original route for geometry, software pixels, native
   semantics/actions and lifetime behavior. Prove actual native fast-path use
   separately from ordinary timing. After race/vet/native checks, compare five
   interleaved original/candidate pairs in both CPU and GPU modes. Adoption keeps
   the existing >=0.4 ms lower median run-p95 and >=4/5 improved pairs in each
   mode, without tail/correctness/retention regression. Reassess feasibility after
   90 minutes; reject rather than tuning the prototype indefinitely.

This is a tree update/layout experiment, not another AX serializer, packed
upload, tile-culling, retained paint-fragment or resolved-shaping cache variant.
No implementation or adoption is claimed here. It does not fix the still-failing
table matrix; full tree and general UI qualification remain required afterward.

## Evidence and limits

The producer binds the exact SDK vector, executable, fourteen helper/control
pins, actual owned process receipts and native raw logs. Its owned executor is
the previous reviewed executor with only an explicit ordinary environment
function replacing the forced two-thread setting; GOMAXPROCS, GOMEMLIMIT,
GOFLAGS and GORACE are absent. Watchdogs, bounded logs, EOF, cancellation,
exclusive lease and owned-group retirement remain intact. Direct execution makes
no enforced network-denial claim. Native scroll events use the actual originating
FIFO/handler and frame tickets, separate from physical pointer delivery. No
placement operator, clipboard, preferences, input source or VoiceOver is used.

Eight pure malformed-trace controls reject callback, generation, accounting,
offset, workload, terminal, diagnostic-mode and font changes. Their historical
raw shape fixture has explicit in-memory metadata adaptations and supplies no
current runtime qualification. An initial pure postprocessor comparison failed
because integer dictionary keys/tuples differed from their serialized JSON form.
The original failure/script is retained; the corrected comparison normalizes the
recomputed summary through JSON without changing metrics, native inputs or pins.
No native case is rerun to hide that failure.

Internal review covers the plan, outcomes and bounded implementation decision.
Prism reviewed the three documentation files with Anthropic Claude Sonnet 5.5
in one provider call. Its one retained medium finding asks for explicit per-mode
sample/warmup counts; the report and both summary specs now provide them. The
raw CLI exit one and finding are preserved, with internal closure and no provider
rerun. The evidence archive excludes SDK/executable/font payloads and retains
raw records, source-bound validators, controls, original postprocessor failure,
identity proof and reviews. Original open lifecycle anomalies remain in the
preceding checkpoint. Native/GPU memory, presentation/pacing, physical/human
input, cold caches, 500 visible rows, expansion/filter, repeated wide/deep matrix,
full P6/P11/Mac acceptance and the other delivery milestones remain open.
