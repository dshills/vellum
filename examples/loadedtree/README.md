# Loaded tree example

This independent Go module demonstrates `ui.DataTree` with a genuinely stored
immutable tree, rather than a procedural million-row native callback. The small
window owns up to 10,000 nodes; the large window owns up to 100,000. Their models,
index requests, selection, viewport, filter field and worker credits are independent.
Stored labels/keys/parent links include one huge branch and a separate deep chain.

After preparing the native driver, run from this directory:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-p0-cache
. ../../drivers/cogent/env.sh
go run . -nodes 100000
```

The initial empty controls remain usable. **Load stored nodes** generates and
validates immutable snapshots on a loader worker, then admits request-bound index
workers. **Expand all**, **Collapse all** and disclosure arrows request new indexes;
the last accepted view stays usable while work is pending. A canceled accepted
worker retains its task credit until both worker and owner delivery are terminal.
A capacity refusal preserves the last admitted request. Selected identity survives
collapse/filter hiding, and filter results include matching ancestors. Navigation,
selection, scrolling and source admission on the owner do not walk the full tree.

The filter is case-sensitive label matching. Apply uses `SettleTextField` to seal
prior originating native edits/preedit before reading the controlled text; it
never reads a speculative preview from a toolbar callback. Cancel also revokes a
pending settlement and latest worker token. While the shared initial loader is
pending, the toolbar displays **Cancel load** and revokes that loader plus both
window requests; after loading, **Cancel work** affects the originating window.
A canceled or already-published older loader result cannot install stored data or
admit index work, and its credit remains occupied until both terminals retire. Workers capture only sealed immutable
inputs; accepted selection and viewport changes made during work remain current
when `Outline.Complete` succeeds.

```sh
go test -race ./...
go test -tags offscreen -race ./...
go vet -tags offscreen ./...
go build -race -o /tmp/vellum-loadedtree .
/tmp/vellum-loadedtree -nodes 100000 -smoke
```

The offscreen tests supply fixed fonts/theme and route actual primitive mouse/key
and originating field events through Core, with explicit worker execution and
owner batches 1/7/64. They compare the final measured visible row geometry and
selection, test wide-branch disclosure, deep-chain End, hidden selection,
queued-key→Apply ordering, older cancellation, closed-window revocation and dual
terminal credit retirement. They do not invoke UI callbacks as synthetic input.

`-smoke` is a bounded Cocoa **owner-command and measured-geometry** workflow. It
loads stored data, expands/reveals a deep selected key, collapses and retains the
hidden key, filters/reveals ancestors, uses real filter settlement, cancels work,
and checks an independent survivor after logical domain close. It observes
completed native frames and public model/lifetime state. It does not inject desktop
pointer/key events, join compositor or private worker teardown, or qualify human
IME/VoiceOver, external AX behavior, native services/GPU, frame latency, or full P6.
Native rows are pooled and bounded at 500; deep indentation is clamped, and labels
are single line. Source/index logical storage statistics exclude allocator/maps,
shaping, borrowed previous results and RSS; this is not a process memory limit.

The integration owner records final source-bound checks and Cocoa evidence in
[the durable report](../../docs/reports/macos-large-tree/README.md). Scratch-only
exploratory logs are not a final qualification record.


## Warm native frame measurement

Use an optimized non-race binary for the separate `-perf` mode:

```sh
go build -o /tmp/vellum-loadedtree .
/tmp/vellum-loadedtree -perf -nodes=100000 -perf-warmup=128 -perf-samples=240
/tmp/vellum-loadedtree -perf -nodes=100000 -perf-deep -perf-warmup=128 -perf-samples=240
```

Repeat with `-nodes=10000` at the same effective viewport. The fixture requires
at least 64 measured clipped rows and refuses smaller layouts. Its compact
9-logical-pixel rows and 12-Dp typography define a dense workload at
the recorded native scale; requested dimensions do not certify a physical display.
Each scroll uses actual originating native input and requires one owner callback
and its matching completed native frame. The first scroll is separate from the
closed warm trajectory and is not cold-cache evidence. Geometry persistence is
disabled in this fixture. See the [measurement contract](../../specs/tree-performance.md)
for timing boundaries, pairing, bounded refusal and the unchanged 4 ms p95 target.
`-perf` excludes `-smoke`; performance flags require explicit `-perf`.


Use `-perf-gpu-scene` to select the optional GPU scene renderer for a separate
comparison run. Omit it for the default CPU renderer. Add `-perf-source-details`
only in a separate diagnostic run to observe source routing, raster/copy clocks
and image API attempt counters. It adds measurement work and cannot qualify
ordinary latency targets. Both flags require `-perf`; every JSON event includes
`gpu_scene_rendering` (requested mode, which permits fallback) and `source_details`.
Detailed runs must fit the recorder's 4096-ticket cap, including warmup and both
first/reverse scrolls; an oversized workload is refused before data/UI creation.

```sh
/tmp/vellum-loadedtree -perf -perf-gpu-scene -nodes=100000
/tmp/vellum-loadedtree -perf -perf-gpu-scene -perf-source-details -nodes=100000
```

See the [route investigation](../../docs/reports/macos-tree-gpu/README.md) for
separate diagnostics and a rejected scanline-allocation prototype. No renderer
change was adopted by that investigation.

## Native presentation preflight

`-perf -nodes=100000 -perf-presentation=/private/tmp/new-loadedtree-presentation.json`
records a separate bounded raw journal using the public native presentation API.
It excludes other diagnostic modes and requires a new file. Every performance
event is labeled as diagnostic. The strict validator joins receipt, completed CPU
generation and actual native presentation, retaining losses and teardown state.
These serial runs cannot qualify sustained 60 Hz cadence or physical input latency.
See the [capture contract](../../specs/presentation-capture.md) and current
[Mac release policy](../../specs/macos-responsiveness.md); the earlier 4 ms budgets
remain optimization targets. No real presentation journal has been qualified by
this checkpoint.

## Fixed-deadline 60 Hz workload

`-perf -nodes=100000 -perf-samples=600 -perf-60hz
-perf-presentation=/private/tmp/new-loadedtree-60hz.json` selects a separate
10-second planned workload. Use the flags on one command line and a new file.
It retains every planned slot, late backlog, generation-bound viewport change
and actual native feedback. Per-event stdout/memory profiling stays outside
the cohort; validation/owner polling stays included. The version 2 analyzer uses
N-1 final-state intervals and reports collapsed states and overruns explicitly.
See the [paced contract](../../specs/paced-presentation.md) for exact bounds,
commands and qualification limits. These tools have not yet qualified live 60 Hz
performance, physical input or the remaining Mac release workloads.

Performance fixtures initialize Core settings before disabling window geometry
persistence, and restore the loaded process-local policy after teardown. Saved
settings are read, never saved by this guard. Settings errors or invalid zoom
refuse the fixture before native setup. The ordinary interactive demo is unchanged.

### Opt-in display backpressure experiment

The Mac-only `-experimental-display-backpressure=1s` option requires `-perf`
and a new `-perf-presentation` journal. Its timeout must be 100 ms through
5 s. It enables one display credit per native window before its first native
acquisition; a missing or skipped display callback refuses the run. The journal
records `Config.DisplayBackpressureNS` only when enabled. Zero keeps the existing
rendering and journal behavior.

This is a prospective experiment, with no native performance qualification yet.
The same input workload, acquisition prefix, final presentation requirements and
latency/cadence thresholds still apply. It does not drain startup paint work or
permit late unarmed skipped acquisitions. Run it in a separately reviewed native
scope; existing captures retain their original results.
