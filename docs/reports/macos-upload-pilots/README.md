# Mac upload optimization pilots

Two tested renderer prototypes were rejected: neither reduced warm native UI
work in its bounded pilot. No native patch was added and the production renderer
remains unchanged at this checkpoint. This prevents adopting additional memory
or lifetime complexity without a demonstrated benefit.

A current optimized GPU-scene process recorded 512 fixed10k scroll samples with
a CPU profile. Cumulative CPU samples included 0.64s under resources.uploadInput
and Queue.TryWriteBuffer, compared 0.18s under Packet.Encode. About 4.42s appears
under native WaitEventsTimeout; this is process-level sampling and must not be
classified as UI CPU execution. The diagnostic run cannot qualify warm targets.

The first prototype packed config and four inputs into one bounded upload while
keeping shader bindings and painter order. Exact native integer/text/path parity,
shared-device fallback, resize, resource retirement and allocation-refusal tests
passed; internal review caught and fixed a separate padding-bound discrepancy.
The second kept the original five GPU buffers and skipped writes whose bytes
matched detached bounded copies. Native repeated/mutable input, replacement,
pixel parity, fallback and retirement tests passed. Neither prototype was adopted.

Each pilot ran four sequential release processes: two alternating baseline/
candidate pairs, 128 guarded warmup callbacks and 120 samples per process. All
paired 400-cell geometry and callback guards passed. Both use the same observer
and opt-in GPU scene mode. Profiling was disabled for all ordinary processes.

| Prototype | Baseline / candidate p95 (ms) | Baseline / candidate p99 (ms) | Mean UI work (ms) | Mean Go allocation (MiB/sample) |
| --- | --- | --- | --- | --- |
| Packed upload | 6.622 / 6.647 | 7.463 / 7.398 | 5.712 / 5.755 | 9.341 / 10.251 |
| Unchanged upload copies | 6.621 / 6.840 | 7.400 / 7.666 | 5.712 / 5.827 | 9.312 / 9.301 |

Two runs of one case are a screening pilot, not statistical proof of a general
regression or a complete P6 campaign. The observations provide no useful gain
that warrants a full campaign or production adoption. Every process still fails
the unchanged 4ms p95 / 6ms p99 targets. Process allocation includes observers,
helpers and runtime, rather than framework-only attribution.

Current baseline mean phase observations are approximately 1.39ms reconciliation,
1.94ms layout/paint and 2.26ms raster/upload. These are separately measured
boundaries in this fixture, not independent percentiles to add. The next priority
is cell layout/render reuse and reconciliation costs; another upload-only change
is not justified by these pilots.

[Qualification](records/qualification.json) and [archive index](records/evidence-index.json)
retain both prototypes, source reviews, raw native logs/receipts, the profile and
producer scripts. The archive excludes executables; their hashes remain in build
and measurement receipts. Prepared source stayed exact at 6202 files. No full P6/P11, native memory, GPU completion,
presentation, human input/accessibility or distribution gate is closed.
