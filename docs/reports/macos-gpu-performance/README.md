# Native CPU/GPU scene comparison

The opt-in GPU scene renderer preserves the tested numeric-table route, but this
campaign does not justify making it the default. Its ordinary frame results are
small and mixed, and it adds roughly 11 MB of process-wide Go allocation per
observation. Both modes fail the unchanged warm targets of p95 ≤4 ms and p99 ≤6 ms.
Mac/P4/P6 completion remains open.

## Protocol and qualification

The runner now accepts `--gpu-scene`, forwards the consumer's `-gpu-scene` switch,
and checks the requested boolean mode in setup, ready, every sample and end.
Missing historical fields mean false; an unmarked or mismatched GPU run is
refused. This selects a renderer which can fall back to CPU; it does not prove
GPU eligibility by itself. The added field is emitted after the completed frame,
outside recorded UI work. Thirty-eight CPU protocol tests and loadedgrid race/vet
checks pass, and the optimized release build is qualified. Internal source review
approves this four-file protocol change against 1f8d93a. The renderer implementation
and all 6,166 prepared files are unchanged from that reviewed checkpoint.

Forty sequential native processes use one identical optimized binary, alternating
CPU/GPU option order across five repetitions for each fixed/variable stored
10k/million-row case. Each completes 128 closed-trajectory guarded warmups and
120 measured events. All 4,800 ordinary observations and outliers are retained.
No profiling, source/input diagnostic or GPU readback is used in these processes.
Full paired geometry matches across modes and row counts. Source/prepared/binary
identities remain stable through the campaign; this is not a complete compiler,
dynamic-library or external module-cache closure.

The independent audit reparses all 4,920 stages from raw logs and checks terminal
exits, flags, warm guards, tickets, eleven raw timing phases, geometry and
recomputed statistics. A reviewer-only initial argv expectation omitted
`-variable-rows`; that failed helper attempt is preserved, its expectation was
corrected, and no source or native result was changed or rerun.

## Ordinary results

Pooled nearest-rank p95 / p99 of recorded total UI work, in milliseconds;
negative change is faster. Phase percentiles are not added together.

| Case | CPU p95 / p99 | GPU option p95 / p99 | p95 change |
| --- | ---: | ---: | ---: |
| fixed-10000 | 12.055 / 12.800 | 11.903 / 12.440 | -1.26% |
| fixed-1000000 | 11.784 / 12.514 | 11.356 / 12.521 | -3.63% |
| variable-10000 | 12.221 / 12.935 | 12.169 / 12.433 | -0.42% |
| variable-1000000 | 11.908 / 12.698 | 12.204 / 13.158 | +2.49% |

The variable million-row GPU tail is slower. Other pooled p95 differences are
small; process variation and all maxima remain in the raw summaries. No general
speedup or target pass is inferred. Queue/acquire/present observations stay
separate; GPU execution and compositor presentation are unavailable.

## Allocation and follow-up

Median process-wide Go allocation between observations, in decimal MB, plus
observed GC-cycle sums across the five runs. These intervals include measurement
helpers and other Go activity; they do not isolate framework allocations.

| Case | CPU MB | GPU option MB | CPU / GPU GC cycles |
| --- | ---: | ---: | ---: |
| fixed-10000 | 16.50 | 27.44 | 288 / 477 |
| fixed-1000000 | 15.73 | 26.48 | 45 / 75 |
| variable-10000 | 16.54 | 27.68 | 284 / 473 |
| variable-1000000 | 15.78 | 26.58 | 45 / 75 |

Retained Go heap and end-only process RSS/CPU remain in the receipts. Domain and
projection memory are reported separately; subtracting them from RSS would not
establish private UI or native/GPU allocation. No native-memory acceptance follows.

Two separate fixed-10k CPU sampling processes use the same binary after guarded
warmup. They are diagnostic and excluded from ordinary results. Native cgo calls,
including event waits, dominate the profiles; these samples do not attribute the
extra allocation or independently establish OS-thread CPU use. ScenePacket,
outline processing and packet Encode appear. A source audit identifies full-target
span-row headers, row append/compaction, full-target encoding arrays and glyph
outline/mask work as allocation candidates, distinguishing shared CPU costs.
The next implementation target is bounded span storage for touched rows, with
unchanged pixel arithmetic, ordering, refusal and fallback semantics. It is a
testable allocation hypothesis; its benefit is not established by this campaign.

## Separate GPU route evidence

Four additional source-detail diagnostic processes retain 120 samples per case.
All 480 main-scene records use the deferred route, with raster/copy availability
false and zero SetFromGoImage attempts. These records are not pooled with ordinary
timings and do not count physical transfers or GPU completion. They establish the
tested numeric-table route, not broader modal/sprite/window or professional-scene
eligibility. The earlier one-process ordinary comparison is preserved as exploratory
evidence and is also excluded from the 4,800-observation campaign.

Internal source and independent raw-evidence reviews cover this measurement
checkpoint. The significant renderer batch was previously reviewed with Prism
using Anthropic Claude Sonnet 5.5 at 1f8d93a; no new provider review is claimed for
this small measurement-protocol delta. The renderer remains disabled by default.
No input sources, VoiceOver, display, power or other OS preferences were changed.
Human/native services, cross-display behavior, broader scene support, native/GPU
memory, sustained targets, tree/workbench workloads and public distribution remain
open requirements.
