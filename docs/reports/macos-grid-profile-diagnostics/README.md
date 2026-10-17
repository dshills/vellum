# Current grid allocation, CPU and native-stack diagnostics

Clean source `30025cc538e7928fea27113a61a852ab91c5fe79` completes three native
million-row fixed-grid diagnostics with 1,024 measured callbacks each. Two collect
Go CPU profiles with CPU/GPU-requested rendering. The third requests GPU rendering
and also receives one ten-second, exact-PID macOS stack sample after warmup and the
first measured callback. All owned app/sampler groups terminate, their logs reach
EOF, and later independent process observations find the groups absent.

These observations guide investigation. They establish no optimization gain,
ordinary frame-budget acceptance, GPU execution/presentation qualification,
retained/native/GPU memory result or full macOS pass. The native stack report's
physical-footprint header is an incidental process observation, not a memory test.

## Build and execution identity

Both clean checkouts contain 4,704 tracked files and 6,249 prepared files at native
admission and terminal boundaries. The newly compiled non-race release executable
is 59,340,930 bytes, SHA256
`2ab852d1e758d047814df5559d0fef21c455364df155536faa7736ff4ae4047d`.
Its saved Go 1.27.2 metadata identifies `30025cc` and `vcs.modified=false`.

The executed build producer contained a conditional-expression defect that could
skip VCS assertions on a non-race build. The original producer is preserved;
actual saved build metadata was independently reconciled, and the future producer
now checks revision and clean status unconditionally. These are separate files;
this report does not claim that the corrected producer performed the saved build.

Each app uses 128 unmeasured warmup callbacks and a fixed one-million-row dataset.
Native validators require exact CPU-only profiling identity, requested renderer,
window/geometry guards, sample order/counts and successful closure. The first two
profiles match all 1,024 sampled geometries and native endpoints. Profiling and
serialization overhead remains in the diagnostic observations. No desktop input,
input-source, VoiceOver, clipboard or OS preference changes are generated.

The exact-PID sampler starts only after a complete checked setup/ready/first-sample
prefix. It independently rechecks the owned app's PID, process group and executable
at launch. The app's log pump continues during sampling. Universal/local stops,
the shared native-test lease, per-child watchdogs and independent group retirement
apply. The non-daemon sampler worker must actually finish before the lease is
released. Twenty-nine non-GUI sampler controls and eighteen CPU-run controls pass;
historical owned-process fault controls are retained as supporting evidence.

## What the diagnostics show

The first two Go profiles place about 89–90% of flat CPU samples in
`runtime.cgocall`; their cumulative stacks include the Cocoa event loop. This does
not identify a native function bottleneck. The additional native wall-stack sample
contains 7,551 main-thread observations, including 6,422 under Mach-port waiting,
772 under full accessibility update blocks, 185 under pending update blocks, 513
under full snapshot commit, 275 under snapshot validation and 259 under the full
JSON reader. Each category counts inclusive descendants once; categories overlap.
These counts are stack occupancy, not CPU percentages or elapsed function times.
The native sample does not justify attributing the Go `cgocall` totals to AX work.

The CPU-profiled frame phase distributions give the following nearest-rank p95
values. They remain diagnostic and do not replace the separately qualified
ordinary-run protocol.

| Requested renderer | UI work (ms) | Reconciliation (ms) | Layout/paint (ms) | Raster/upload (ms) |
|---|---:|---:|---:|---:|
| CPU raster | 5.441 | 1.345 | 1.236 | 2.887 |
| GPU scene | 4.544 | 1.354 | 1.229 | 1.855 |

Phase percentiles describe different samples and must not be added. The three
substantial measured areas remain reconciliation, layout/paint and rendering.
Native event-loop stack occupancy cannot substitute for those elapsed intervals.
The next optimization should compare a bounded change in one of those paths
against an independent correctness oracle and then ordinary matched native runs.

Separate offline allocation analysis reads the before/after sampled profiles
already captured on source `599442a`, described in the
[intrinsic sizing report](../macos-grid-intrinsic-sizing/README.md). Encoded scene
packets, accessibility encoding/copies, capture and layout appear among allocation
sites. Those cumulative process-wide samples include observers, profile encoding
and forced-GC snapshot work; they are neither precise per-frame allocation savings
nor retained/native memory. Four allocation, four initial CPU and two sampled CPU
pprof commands run offline with local symbolization disabled and no provider call.

## Evidence and remaining gates

The [structured result](result.json), [member manifest](evidence.manifest.json) and
[raw evidence](evidence.tar.gz) retain profiles, native stack text, bounded command
logs, ownership metadata, source/prepared hashes, original/corrected producers,
controls, helper sources and internal reviews. Host/test executables are
fingerprinted rather than copied. Tracked/prepared trees remain identified by Git
and complete hash vectors; the bundle is a diagnostic evidence set, not an SDK.

This evidence-only checkpoint changes no runtime source and makes no Prism call.
The preceding implementation checkpoint was reviewed with Anthropic Claude Sonnet
5.5. New results and documentation receive separate internal reviews.

The prior ordinary intrinsic-sizing comparison remains the applicable latency
finding: all twenty candidate processes miss 4 ms p95, with eight meeting 6 ms
p99. P4/P6/P9/P11, broader controls/canvas/professional workloads, native/GPU memory,
full idle/overnight behavior, physical input/display workflows, human IME/VoiceOver,
SDK component/rights closure and public distribution remain open.
