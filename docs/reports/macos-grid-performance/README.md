# Native grid timing foundation

The optional `FrameMetrics` recorder follows initialized native grid scrolls
through input, accepted owner generation, reconciliation and native composition.
Each elapsed-work total includes all post-input intermediate update/compose work
plus the first frame armed for that generation. Surface ownership/acquisition
and actual backend present calls are reported separately. GC remains included.
These intervals do not measure kernel thread CPU, GPU completion or physical
presentation. Recording currently requires one active native window.

The optimized stored-grid consumer and runner require at least 400 clipped cells
(50 rows × 8 columns), 20 stored numeric columns, identical paired 10k/million-row
geometry, unique completed tickets and one callback per scroll. Budget runs use
at least five fresh processes per fixed/variable case, 128 warmup steps by default
and 120 measured samples. Optional CPU profiling is diagnostic, disabled by
default and rejected by ordinary budget qualification.

One exploratory optimized fixed-10k Cocoa process completed 128 warmup steps
and 120 samples, with no dropped/pending tickets. It rendered 400 clipped cells
in a 767×624 logical viewport at scale 1.6875, retaining 477–540 native cells.
Its p50/p95/p99 elapsed work is 19.30/20.30/22.77ms, with a 29.26ms maximum.
The unchanged targets are p95≤4ms and p99≤6ms: this observation fails them.
Reconciliation and raster/upload dominate the phase observations; independent
phase percentiles are never summed. This single exploratory run had concurrent
host development work and is not the formal five-run acceptance campaign.
Its historical JSON label `ui_cpu_ns` means `FrameSample.TotalUIWorkNS`, an elapsed
work total. Current tooling names that value `ui_work_ns`.

Root portable/native race suites and vets pass. Tagged frame and affected
upstream timing/acquisition race checks pass twenty repetitions. Stock upstream
vet fails on identical baseline composites/unreachable warnings; other analyzers
pass with those two disabled. No broad upstream vet pass is claimed. Fresh
stage34 preparation reproduces all 6,128 prepared files byte for byte.
Internal review approves the bounded timing contract and corrected cancellation,
dispatch compression, prefix-frame and abnormal composition ownership paths.

`archive.json` indexes raw logs, selected immutable inputs, producer scripts,
review certificates and the source-bound timing observation. Binaries are
identified by hash and excluded from Git. Enumerated source/prepared hashes do
not establish a complete compiler/module-cache/system SDK closure. The later CPU
profile is an exploratory process-wide diagnostic and cannot qualify timing.

P6 remains open. The next phase optimizes the measured reconciliation and raster
costs, then repeats unchanged workloads and targets. Native/GPU allocation
attribution, novel strings/cold caches, tree latency, endurance, physical
presentation and external accessibility remain separate acceptance checks.

The subsequent [single host-update checkpoint](../macos-grid-update/README.md)
removes duplicate per-slot updater/style work. Its interleaved fixed-10k release
comparison observes p95 19.55→17.07ms and p99 20.41→18.00ms while preserving
geometry and software-frame pixels. It is a three-process-per-build exploration;
the unchanged targets remain failed and the broader acceptance checks stay open.
