# Current native grid CPU and allocation profiles

Two separate native diagnostics complete on the unchanged packaged SDK from
`238b3b619e5ef274c218cb62d0b348e44e3fcaf8`. Both request the GPU scene route for
a fixed-height million-row stored table, with 128 warmup events and 512 measured
scroll events. CPU and allocation profiling are mutually exclusive fixture
modes; no combined-mode native run is made. These profiled/differential runs do
not qualify ordinary frame latency, a speedup or memory budgets.

## Execution and identity

The native executable is the already compiled private SDK loaded-grid consumer,
SHA256 `ccc37ade4a6c25a7bcb7f4f054c4b1ec1a6a3c1f3c9bfbb2d03209e2af34ece7`.
Runtime source remains SDK `238b3b6`; `cbc25c2` is the separate documentation
endpoint present in both clean checkouts during the runs. Their 4,664 tracked
and 6,246 prepared files, all 7,411 SDK leaves, private archive and binary remain
unchanged. No source change or rebuild is needed for these existing diagnostics.

The first native process records CPU samples only. The second records cumulative
Go allocation profiles before and after its measured events. Profilers start
after warmup. CPU collection runs in its own process without allocation-profile
snapshots. The separate allocation process forces two GC cycles for each before/
after snapshot outside the recorded events, with CPU profiling disabled. The
native writer caps each new exclusive profile file at
64 MiB and reports write/close failures.

Both processes exit with status zero, complete all 512 callbacks with no missing,
dropped or pending tickets, reach EOF on both streams and have their owned groups
observed absent. All 512 paired full grid geometries and native endpoints match.
The client/render extent is 1519×1266 pixels on the recorded DPI162/DPR1 display;
the measured logical grid is 767×624 and satisfies the 400-cell/50-row/eight-
column minimum. These endpoint checks do not prove continuous absence of a
transient display/resize change or physical display behavior.

The reviewed controller uses the shared native lease, universal/local stop
markers, a 180-second per-process watchdog, bounded raw streams and independent
cleanup/finalization attempts. The campaign is sealed after completion. It does
not generate desktop pointer/key events or change input-source, VoiceOver,
clipboard or OS preference settings. Geometry persistence is fixture-disabled.

## What the profiles suggest

Nine bounded offline commands read the three exact profiles with the retained,
source-bound local Go `cmd/pprof` tool. `-symbolize=none` avoids remote/native
symbol lookup. The native application is not executed by this analysis.
The Go installation lacks a `go tool pprof` executable, so the previously built
and verified local tool is reused. An initial restricted analysis attempt fails
at read-only process observation before output creation or pprof dispatch; the
unchanged analyzer succeeds with process-observation permission. This is not a
native rerun or a requalification of the failed observation.

| Function/path | Profile observation |
|---|---|
| `runtime.cgocall` | 75.42% flat CPU samples; native work remains unresolved |
| `syncAccessibility` | 11.46% cumulative CPU samples; 33.47% cumulative allocation-space estimate |
| `reconcileScoped` | 0.33% cumulative CPU samples; 4.60% cumulative allocation-space estimate |
| `gridWidget.LayoutLocal` | 10.84% cumulative allocation-space estimate; 35.60% cumulative object estimate |

These are process-wide sampled profiles and differential allocation estimates,
including frame observers and other goroutines. Cumulative rows overlap and
must not be added. Main-loop/`cgocall` labels do not establish a busy wait, an
idle defect, exact native CPU work or a causal frame bottleneck. Go allocation
estimates do not measure retained UI bytes or native/GPU/private memory.

The source refresh's Reconcile timer includes `syncAccessibility` after
`reconcileScoped`; the latter function alone is not the whole region. The local
grid layout explicitly sizes dirty text with SizeUp/SizeDown/SizeFinal. These
boundaries plus the profiles direct the next investigation toward accessibility
snapshot preparation and repeated grid text sizing, rather than assuming keyed
node reconciliation dominates. No optimization or expected saving is asserted.
Candidate changes require bounded lifecycle/ownership, arbitrary text/style/DPI
correctness, pending/completed accessibility transaction and action-admission
regressions, then matched ordinary native measurements before adoption claims.

## Remaining qualification

The existing 4 ms p95 / 6 ms p99 failures remain failed. Profiling does not close
P6/P11, prove actual GPU execution for ordinary grid frames, physical input,
presentation, native/GPU memory, human accessibility or Mac release acceptance.
Internal prelaunch/result/documentation reviews cover this evidence-only phase;
the implementation's stage 0075 Prism review remains its last provider review.
No additional provider call, public release, signing or notarization occurs.

See [structured result](result.json), [evidence manifest](evidence.manifest.json)
and [raw evidence](evidence.tar.gz). The [ordinary tile-binning comparison](../macos-adaptive-tile-binning/README.md)
and [SDK consumption](../macos-current-sdk/README.md) retain their separate scope.
