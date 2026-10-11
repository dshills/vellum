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
