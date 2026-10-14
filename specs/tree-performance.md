# Stored-tree warm performance

The Mac `examples/loadedtree -perf` consumer uses one live native window and a
sealed 10,000- or 100,000-node dataset with a huge branch and a separate deep
chain. Stored generation/construction and initial expanded indexing precede
mounting and are measured separately. The `-perf-deep` case starts beyond the
first 64 levels of the chain; indentation remains capped by the existing native
adapter. Data/index arrays are outside UI-overhead interpretation.

The fixture uses 9-logical-pixel fixed rows, 12-Dp Noto Sans and a bounded tree
host at the measured display scale. At least 64 nonempty clipped row frames must
exist after every measured native frame, with no more than 500 retained or
materialized rows. A smaller measured workload is refused. Report effective
window/viewport geometry; requested dimensions and nominal style values do not
establish physical display or 500-visible-row qualification.

Each 32-event warm trajectory moves by +12 logical pixels for 16 events, then
−12 for 16 events. At least 32 unmeasured events, in complete cycles, precede the
120–2048 measured events. One first scroll plus its reverse are recorded
separately; initial scene/font/data state already exists, so these do not qualify
cold caches or startup. Warm-up must return to its original viewport/geometry
with exactly one callback per event. Every subsequent event must independently
match its requested offset, generation +1, source stamp, dimensions, bounds and
scale; it cannot rely only on final zero-sum closure.

`EnqueueTreeScroll` is an owner-only opt-in diagnostic. It admits initialized
actual MouseScroll input to the originating native FIFO at current completed
body geometry, using a reserved node that ordinary scroll compression cannot
replace. Pending index work, stale layout, disabled or out-of-scope bindings,
foreign/closed ownership, duplicate tags and exhausted recorders refuse. Context
cancellation retires measurement but does not retract delivered input. Boundary
no-ops, refused/unreturned owner transitions and missing native compositions
cannot produce successful samples. Ordinary tree input keeps its existing path.

The existing FrameMetrics recorder captures input, owner, reconciliation,
layout/paint, gathering, raster/upload and encode/submission work, including
intermediate native paints before final reconciliation. Queue, acquisition and
Present-call observations stay separate. The sum of measured elapsed work
regions is a CPU-side diagnostic, not OS-thread CPU, GPU execution,
compositor presentation or photons. The unchanged tree target is recorded UI
work p95 ≤4 ms. Do not add independent phase percentiles to qualify it.

InspectTree/recorder polling, JSON, memory observations and initial indexing are
outside recorded regions. Go allocation deltas include those observers and other
process goroutines; post-GC heap includes caches and recorder state. Native/GPU
memory needs separate attribution. One callback, matching completed generation,
unique ordered ticket, current measured clips, zero dropped/pending history and
successful process shutdown are required for every accepted run.

Compare wide/deep cases at 10k/100k with identical consumer bytes, effective
viewport, scale, normalized clipped-row geometry and cardinalities at each phase.
Use five optimized release processes per build, alternate build order, retain
raw first-scroll/warm samples and report distributions by run. The constructor
and worker index costs remain separate. This warm-scroll fixture does not close
large-branch expansion/filter interaction, cold caches, maximum-visible workload,
physical display, native memory, full P6 or P11/Mac acceptance.


The consumer accepts `-perf-gpu-scene` and `-perf-source-details` only with
`-perf`, excluding smoke. Every event stamps both requested GPU mode and source
measurement mode. GPU mode permits the existing fallback; use separate source
records to establish routing for the observed diagnostic corpus. Source details
add clocks/counters and are excluded from ordinary latency comparisons. Their
recorder is capped at 4096 tickets including warmup, measured and first/reverse
scrolls; invalid totals refuse before stored-data/UI construction. Defaults remain
CPU rendering and ordinary recording. Never pool CPU/GPU or ordinary/detailed
runs, and retain requested mode plus actual route evidence separately.

The [route investigation](../docs/reports/macos-tree-gpu/README.md) rejected
packed scanline compaction after 20 matched GPU-mode processes: fewer allocation
objects alone did not establish a useful byte or latency improvement. It added
no prepared patch and did not qualify the unchanged 4 ms target.
