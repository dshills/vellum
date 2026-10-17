# Bounded GPU encoding workspace and current native performance

Stage 0078 adds renderer-owned reusable GPU encoding byte buffers. Aggregate
retained capacity is at most 1 MiB; differently shaped growth resets backing
before replacement, and accepted larger encodings keep detached allocation.
Integer counts/positions and mask lookup remain per-call. Detached Encode and
EncodeGPU output, packet admission, painter order and CPU reconstruction stay
unchanged. Only the explicitly borrowed encoding uses scratch. Synchronous
checked queue writes consume it before reuse; packet journals/GPU retirement do
not retain it. Resize/release discard the workspace.

## Engineering result

The same-process offscreen benchmark takes five observations per variant with
Go 1.27.2, CPU2, 200 ms minimum, no race/profiler and a 4,800-mask-operation
fixture. Median detached encoding is 415.775 us / 508,112 B / six allocations;
borrowed encoding is 351.274 us / 74,491 B / two allocations: 15.5% lower time
and 85.3% fewer allocated bytes. This is a local encoding result, not native
latency or total memory qualification. The borrowed first allocation amortizes
into the benchmark and its retained backing remains separately bounded.

The five interleaved native pairs use the same optimized ARM64 consumers,
Go 1.27.2, no race/profiler, 128 warm events and 120 measured events per process.
The normal release runtime has GOMAXPROCS/GOMEMLIMIT unset. Both request GPU
scene mode, fixed 10k rows and identical observed 400 nonempty clipped cells
(50 rows/eight columns, scale 1.6875, 477 retained cells). The client is
1519x1266 on this M4 Pro / LG HDR DQHD host. Every pair's geometry matches.
This fixture is not the entire original Mac workload/configuration matrix.

Median run-p95 UI work is **5.271 ms before / 5.281 ms after**, with every run
above the unchanged 4 ms target. The candidate does not meet the 4/6 ms p95/p99
pair. No meaningful latency improvement is established. Median per-run median
process-wide Go allocation observations are 8,511,824 / 7,459,744 B; these
include observation helpers/runtime activity and cannot be called exact
framework allocation or retained memory. Native GPU routes/presentation are
not proved merely by the ordinary mode request.

## Basis and rejected shortcut

Before this change, five interleaved ready-snapshot baseline/candidate pairs
at GOMAXPROCS=2 produced median run-p95 5.211 / 5.270 ms. A separate five-run
production-default candidate also failed (median run-p95 5.178 ms). Thus the
previous accessibility shortcut has no demonstrated native frame gain, and
lifting the processor cap does not close the failure. The broader table matrix
was stopped at its first failed workload rather than repeating known failures.
An initial controller failed at summary field lookup after one passing baseline;
its raw failure is retained and that run is not pooled into either experiment.

Separate current-source diagnostics attribute median GPU host capture/encode/
input intervals of 1.150/0.425/0.151 ms. All 120 detailed draws were accepted,
with no CPU fallback. A differential allocation profile over 120 measured
frames reports 937.91 MiB process allocations: accessibility publication is
373.88 MiB cumulative; GPU drawing is 270.22 MiB cumulative, including the
74.42 MiB flat encoding allocation. Those buckets overlap and are not summed.
Profiling is diagnostic; native main-loop waits dominate its sampled cgo stack,
so that stack is not treated as an isolated renderer CPU cost.

## Correctness and review

Full affected sceneplan/rasterx/gpuscene offscreen race tests pass. Final real
Metal sceneplan/gpuscene race tests pass in 12.977/1.780 seconds without skipped
native cases. They cover exact mixed-operation encoding, detached ownership,
alternating sizes and cross-shaped capacities, exact/first-over-cap boundaries,
empty padding, original tile-budget refusal/recovery, resize/repeated release,
actual checked-write failure/recovery and pixel readback after a different
encoding overwrites scratch before waiting for the submitted work. Vet passes.

The full pinned patch chain regenerates all 5,961 upstream files exactly.
The first whole-tree comparison found 17 extra raster test output PNGs, while
all regenerated source already matched. Outputs were preserved outside source;
no reference/golden was rewritten. The corrected comparison passes. The existing
11-method preparation suite passes after its stage-sequence assertion advances
to include 78. The initial Go invocation from the wrong module directory is
retained separately; it stopped before compilation.

Internal review found the stage assertion and requested the exact-cap/native
overwrite proofs, all resolved. One Prism Anthropic Claude Sonnet 5.5 call
reviews the complete six-file implementation/test diff (one complete snippet
chunk, no truncation). It returns zero retained findings and one medium discarded
only for snippet path mismatch. That raw concern is independently refuted:
reference counting and filling traverse the same retained rectangles in painter
order and overwrite every logical word; empty padding is cleared. Alternating
exact-byte tests cover reuse. No unchanged provider re-review was requested.

Raw commands/logs/samples/profiles, scope corrections, source hashes, failures
and provider results are in records.tar.gz, indexed by records.json. Native
receipts establish direct child exit/EOF and group-kill attempts, rather than
strict recorded PID/PGID whole-group absence. The old global stop stays sealed;
these new authorized cases use fresh local stops. No physical pointer action,
input-source/clipboard/VoiceOver change, signing or public release was performed.

## Next action and scope

**M3 remains failing.** The practical latency result rules out further encoding
scratch tuning as the way to close the roughly 1.2 ms gap. Next implementation
work must reduce full-scene capture and/or coherent accessibility construction,
with a bounded retained representation, unchanged correctness and workload, then
one matched ordinary comparison. This checkpoint adopts a measured allocation
reduction only. Full table/tree/general UI, memory, presentation, endurance,
human accessibility/display and distribution gates remain open.

Baseline is clean 5482f9b; the candidate is that source plus the stage0078 working
tree patch. It is not a clean historical build. Both optimized executable hashes,
exact patch outputs and prepared aggregates before/after are retained. The
previous accessibility-only baseline uses a one-file compiler overlay and is
also explicitly distinguished from a historical clean executable.
