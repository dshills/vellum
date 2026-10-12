# Differential native Go allocation diagnostics

Two separately labelled native diagnostics identify repeated glyph-outline
construction as a large allocation cost in both the CPU and opt-in GPU scene
renderers. Accessibility publication, measurement snapshots and JSON encoding
also allocate substantially. The GPU process adds scene-packet construction and
encoding costs. These observations guide the next implementation work; they do
not establish a speedup or qualify the unchanged warm latency targets.

## Capture and verification

The loadedgrid consumer now provides `-perf-alloc-profile=/new/prefix`, requiring
`-perf` and a separate mode from CPU profiling, source timing and input ledgers.
Two exclusive mode0600 files have a64MiB cap each. Both cumulative Go `allocs`
snapshots force two GCs outside measured events, accounting for the runtime's
possible two-cycle reporting lag. The baseline follows warmup, and the final
snapshot follows all samples and observer work. `MemProfileRate` is recorded and
left unchanged. Partial/empty owned outputs survive failure, without overwriting
existing files; an error prevents successful completion.

Setup, ready, all samples and end identify this mode. Ordinary record validation
refuses it independently of other diagnostic permissions. The ordinary benchmark
CLI never launches it, and explicit allocation diagnostics cannot qualify warm
latency targets. The root runner's character-count prefix preflight is narrower
than Go's UTF-8 byte guard: overlong multibyte paths can reach the child, which
refuses them before program construction. Filesystem path/component limits can
also cause explicit refusal. No successful file admission is inferred from
preflight alone.

Two fresh native processes each use fixed stored10k rows,128 guarded closed
warmups and120 measured scroll events. Exact completed geometry, one callback per
tag, ordering, pending/dropped completion and renderer/profile mode pass the
existing native record checks. Source/prepared/binary hashes remain stable through
both runs. Four saved snapshots are valid inputs to eight successful pprof
commands: flat/cumulative views of differential allocation bytes/objects for each
mode. Each command uses the corresponding baseline snapshot with `-base`.
The recorded memory-profile sampling rate is524288 bytes in both processes.

The consumer race tests, vet, optimized build and40 CPU-only Python protocol
tests pass. Initial check logs precede the full hash maps; they are not presented
as a separately instrumented before/after test closure. A test-helper mismatch in
expected refusal order was corrected and its initial failure is recorded. No
renderer, dependency patch, public API, OS input source, VoiceOver or other OS
preference changed in this checkpoint.

## Observed costs

Values below are pprof's sampled estimates for the whole interval, rounded to
MiB. The tool labels these `MB` but its unit scale is2^20. Inclusive values overlap
with their children and must not be added together. These are one diagnostic
process per mode, not replicated ordinary performance measurements.

| Allocation site | CPU MiB | GPU MiB | Interpretation |
| --- | ---: | ---: | --- |
| Differential profile total | 1916.69 | 2452.92 | Sampled process-wide total |
| Font outline retrieval | 834.28 | 792.72 | Inclusive; CPU `GlyphDataOutline`, GPU `GlyphDataOutlineBounded` |
| Accessibility synchronization | 547.10 | 543.93 | Inclusive, overlapping bridge/tree work |
| `FrameMetrics.Snapshot` | 125.28 | 120.17 | Flat; measurement observer cost |
| `sceneplan.Builder.add` | absent in selected CPU rows | 201.74 | Flat; GPU packet construction |
| `sceneplan.Packet.Encode` | absent in selected CPU rows | 149.20 | Flat; GPU packet encoding |

CPU outline retrieval accounts for43.53% of its sampled total; the GPU bounded
outline path accounts for32.32%. GPU span append/flush also dominate parts of the
object-count view. Full raw top/cumulative outputs, records and profile hashes
are retained in the checkpoint archive. Sampling, observers, profile encoding
and concurrent runtime/adapter work affect these figures. No subtraction of
helper costs, exact allocator census, retained-memory or native/GPU-memory claim
is made. Profiled timing is excluded from the preceding ordinary comparisons.

## Next implementation and Mac acceptance

The next renderer optimization should reduce repeated static glyph-outline
construction with bounded ownership and invalidation, retaining existing font
admission, fallback, exact geometry and pixels. Validate it with independent CPU
and native pixel fixtures, then a matched ordinary campaign before any latency
claim. Scene packet allocation and accessibility publication are additional
measured candidates; observer allocation needs separate diagnostic treatment.

A separate read-only Mac services audit binds37 existing source/evidence files.
It confirms remaining acceptance gaps: actual Open/Save panel selection and
Cancel on owned temporary paths, installed native menu activation, and Finder
file drops. Existing document smokes inject chooser paths. The prior physical
Finder attempt timed out; it remains incomplete. The existing general-clipboard
write flag only restores text, so it cannot be used as preservation-safe on an
ordinary desktop. Private-pasteboard coverage would be a separate fixture; a
full general-service roundtrip needs bounded flavor preservation and interference
handling. This audit ran no new native workflow and closes none of those gates.

Internal source and evidence reviews cover this diagnostic checkpoint between
significant Prism reviews, following the reviewed sparse-band checkpoint495260e.
The GPU renderer remains opt-in. Full Mac, P4/P6/P9, large-tree, sustained/native
memory, human IME/VoiceOver, cross-display and public distribution acceptance
remain open; Windows/Linux qualification is deferred.
