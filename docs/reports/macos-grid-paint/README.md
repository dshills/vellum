# Render-local grid paint styles: measured and rejected

The private four-style-per-pane rendering experiment is rejected. It reduces
sampled Go allocation substantially but fails the frozen ordinary latency bar.
No production, public API, prepared dependency or SDK change is adopted. The
latest complete eight-case table measurement remains the
[current `a2b4143` matrix](../macos-grid-current/README.md); full M3 and Mac
completion remain open.

## Change and correctness evidence

The original `renderCellSurface` copies a large resolved style for every visible
cell. Minimum Go 1.25.6 compiler output explicitly moves that per-cell `style` to
the heap. The candidate instead copies four resolved classes once per pane
render and computes invariant batch eligibility once per class. Compiler output
moves this one `paintStyles` array outside the cell loop to the heap. Every render
owns fresh copies; resolved templates remain immutable. There is no cross-frame
cache or new retained field. Paint order, cell clipping, projection, input,
accessibility, fallback and native editing stay unchanged.

The internal source review approves the candidate's ownership and class invariants.
The complete driver-root package race tests, loaded-grid example race tests, vet and optimized
ordinary/observer native builds pass with exact source pins and terminal owned
group absence. The new lifetime regression covers 96/120/144/162/192 DPI, normal,
selected and active-cursor classes, meaningful inherited color changes, stable
resolved templates/shape epochs, independent ordinary-widget pixels and geometry,
and replay of saved render operations and detached ScenePackets after later
renders and value-surface retirement. The earlier focused run preceded only an
added positive class-mask assertion; the final driver-root package race run includes it.
Compiler evidence precedes that test-only change; exact old inputs and the
unchanged production bytes are preserved.

These are automated software/routing proofs. They do not establish physical
pointer behavior, a real display transition, screen-reader use or presentation.

## Fixed native comparison

The clean source base is `51dbe40521d34c76abcc1ab82300b394a55fdb01`, with unchanged
`a2b4143` production code. The baseline is the current normal `81b85e39…` binary,
not the earlier widget-based SDK binary. The normal candidate is `0da934c8…`;
`93412302…` is a separate observer build. Ordinary timing uses optimized Go
1.25.6 without race, offscreen, escape-analysis or observer instrumentation.
Both variants use the unchanged retained private SDK's source-built `786a…`
native archive, with stage0080 Go overlays. This does not qualify the default
checkout native archive or constitute a refreshed SDK.

All measured runs have 10,000 stored fixed-height rows, 400 or more nonempty
clipped cells, the same 1519×1266 native render size and logical DPI 162. Each
fresh process executes 128 guarded warm events and 120 measured events. Five
interleaved baseline/candidate pairs run separately for CPU and GPU. All ten
paired geometry arrays match exactly. Percentiles are nearest-rank per run,
then median across five runs; samples are never pooled for percentile claims.

| Ordinary result | CPU baseline → candidate | GPU baseline → candidate |
| --- | --- | --- |
| Median run-p95 | 4.748167 → 4.690332 ms | 4.254502 → 4.104204 ms |
| Median p95 improvement | 0.057835 ms | 0.150298 ms |
| Improved pairs | 3/5 | 5/5 |
| Median run-p99 | 5.162665 → 5.243124 ms | 4.577459 → 4.442754 ms |
| Runs with p95 ≤4 ms | 0/5 → 0/5 | 0/5 → 0/5 |
| Runs with p99 ≤6 ms | 5/5 → 5/5 | 5/5 → 5/5 |
| Median per-run allocation p50 | 4,160,840 → 3,457,944 B | 5,749,032 → 5,043,288 B |
| Median terminal process Go heap | 26,007,216 → 26,043,696 B | 27,376,960 → 27,409,600 B |

The admission bar was frozen before execution: at least 0.4 ms lower median
run-p95 and four improved pairs out of five in **both** modes, nonincreasing
median p99 and count of runs above 6 ms, no more than 1 MiB extra median terminal
Go heap, exact geometry and correctness. CPU fails the gain, pair-count and
median-p99 guards; its count of runs above 6 ms stays zero, so that separate
run-count guard passes. GPU fails the gain guard. All twenty ordinary p95s still
exceed 4 ms. Lower allocation alone does not justify adoption under this bar.

Exactly 26 native processes execute: two separate functional diagnostics
(240 measured frames), twenty ordinary timing runs (2,400 samples), and four
separate allocation captures (480 samples). Both diagnostics positively observe
current value layouts/shapes and zero native body children; CPU proves fresh
raster/copy, GPU a fresh final accepted scene with five input writes, one submit
and no fallback. The validator permits valid intermediate reuse while requiring
a fresh final scene. Its pure controls retain three positive and 27 negative
cases, including fresh-prefix/reused-final rejection.

The reviewed plan permits 36 further processes only if the paired bar passes:
six remaining route diagnostics and thirty remaining candidate timing runs.
Rejection seals the scope after the first 26; all 36 stay unexecuted. No new full
candidate table matrix is claimed. Every executed process has complete raw
streams, stable pins, shared-lease cleanup and terminal owned-group absence.
No failed or canceled native case is retried. No live app remains from this scope.

## Separate allocation attribution

Four before/after cumulative Go allocation captures use sampling rate 524,288;
twelve offline `go tool pprof` commands subtract the warmed baseline and inspect
flat/cumulative allocated space and allocated objects. These sampled process
profiles include concurrent adapter/observer work and profile encoding. They do
not measure native/GPU allocation, retained memory, framework-only budgets or
ordinary latency. Pprof's printed MB units here correspond to MiB.

| Sampled allocated space | CPU baseline → candidate | GPU baseline → candidate |
| --- | --- | --- |
| Total process difference | 495.00 → 387.19 MiB | 666.25 → 561.30 MiB |
| Flat `renderCellSurface` attribution | 74.63 → 1.01 MiB | 90.15 → 2.52 MiB |

The compiler and sampled profiles support the intended allocation mechanism.
They do not turn a failed latency comparison into a pass or qualify the general
64 KiB/100-allocation frame budget. The small terminal heap increases stay inside
the experiment's retention guard, but full M4 resource/endurance gates remain.

## Failure history, review and next work

The first preparation refuses SDK drift before any native launch: Python imports
from pure validator controls create three bytecode cache files. Their exact
bytes/hashes are preserved, only those generated files are removed, bytecode
creation is disabled, and the complete original 7,880-file SDK vector (7,879 payload files plus
`SDK_MANIFEST.json`) is restored
before fresh preparation and independent admission. No original SDK source
changes. This failure and restoration remain in the records.

The archive retains the private candidate and baseline source, exact compiler
versions, raw native/profile records, conditional plan, controls and owned
execution receipts. Native executable payloads are referenced by hash rather
than copied. The checkpoint must record internal results and final reviews, one conserved
Prism review with Anthropic Claude Sonnet 5.5, actual completed coverage and all
retained/discarded finding dispositions before commit. The
checkpoint and commit/promotion receipts follow the archived evidence snapshot.

Retire this style experiment without further style/cache micro-tuning. The next
concrete integration deliverable is M2's complete native professional workspace
workflow at 10k/1M rows, using the existing records and domain assertions; the
18-action smoke remains insufficient. M3 still needs a materially different,
measured implementation decision and the complete table/tree/general-UI matrix.
The [Mac delivery plan](../../../specs/macos-delivery.md) retains every remaining
M1–M7 exit criterion and the unchanged performance/resource budgets. No soak,
physical input, system preference change or public distribution follows from
this experiment.
