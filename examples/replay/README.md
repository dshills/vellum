# Managed UI replay consumer

This independent Go module uses public Vellum APIs. Its fixed fixture and
bundled record drive actual managed widgets through the Cogent offscreen router.
The project-name edit comes from pointer focus and keyboard input; table actions,
worker cancellation and undo run through the visible controls. Command-only
checks are separate.

After preparing the repository's pinned dependencies:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-cache
. ../../drivers/cogent/env.sh
go test -race ./...
go test -race -tags offscreen ./...
go run -tags offscreen ./cmd/native -batch 64
```

The native tests execute the same 51-operation record at batches 1, 7 and 64.
Nine checkpoints compare exact domain state, measured elements, actual focus,
UTF-16 selection/directional focus, viewport and software-image digest. They
report event/frame counters separately because batching may change transport
counts while preserving results. A separate authoring test verifies the bundled
record against one measured pilot; acceptance never reauthors it per batch.

`cmd/native` prints JSON and exits after explicitly closing the driver,
application, clock and executor. Its only input option is dispatch batch size.
The bundled record decoder validates trusted fixture data before construction;
this example does not expose an external automation service.

The manual driver requires supplied font bytes, explicit theme/DPI and clock.
The fixture uses checked embedded Cogent fonts and a fixed theme. It runs with
one active managed manual application per process; its windows share the clock.
Callers keep Cogent global palette/settings/keymap/highlighting fixed for the
manual lifetime and inject the borrowed executor explicitly into domain managers.
Mapped screenshot/zoom/refresh and other unqualified manager shortcuts refuse
before admission. Completion, popup and cursor timing use explicit origin work;
shared text-buffer leases, output readers and autosave are outside this workload.
No OS input-source, VoiceOver, accessibility-permission or appearance settings
are changed. Native OS services and arbitrary custom extensions refuse. The
concrete built-in Workbench supports bounded managed replay; the separate
`TestPublicWorkbenchManualConsumer` uses only public APIs to switch actual dock
tabs and route panel-local button input at batches 1, 7 and 64. This test does
not change the bundled 51-operation record. See the
[workbench replay contract](../../specs/replay.md#managed-workbenches) for identity,
fixed-viewport, persistence and explicit-clock restrictions.

These tests qualify software rendering and routed widget behavior. Cocoa IME,
external accessibility, GPU presentation and sustained performance have separate
acceptance workloads.


## Professional workspace

The `workspace` subpackage is a second independent public SDK consumer. It
combines a 100,000-node outline, table retaining twenty `uint32` fields per row,
notes editor, property inspector, command/history journal, dock tabs, a pannable/
zoomable canvas and an independent shared-model inspector window. Open project,
sort/filter and outline rebuilds use injected workers with bounded admission.
Cancel worker prevents installation even when its result is awaiting delivery.

From this example directory after repository preparation:

```sh
go run ./cmd/workspace -rows 10000
# Use 1000000 for the larger stored workload.
go test -race -tags offscreen ./workspace -count=1 -timeout=9m
go run -tags offscreen ./cmd/workspace-replay -rows 10000 -batch 64 > workspace-result.json
go build -o build/vellum-workspace ./cmd/workspace
```

The native application starts with Open project. Its document tabs expose Canvas,
Table and Notes. Drag Block into the canvas, or Tab to Block and press Return to
place at center; the grid is twenty logical units. Use the shared Undo/Redo
controls for property edits and placements; notes keep their own editor history.
Table filter matches `7` in V01 and sort orders V00 numerically descending.
Outline filtering matches Component 000100 and includes its ancestor. Canvas
click selects, drag/arrows pan and wheel zooms. Main-window closure preserves
the independent inspector. Placement remains same-native-window only.

Each bundled workspace record has 216 fixed operations and 27 completed
observations. The acceptance tests read it unchanged at event batches 1, 7 and
64 for both dataset sizes. Goldens include actual software image identities,
measured geometry/semantics, UTF-16 selection/focus and model state. Independent
oracles derive source-byte hashes, sort/filter results and selected-cell values;
they also check cancellation, outline indexes, notes, tab dragging, placement
history and modal/window lifetimes. Transport counters and native capability
tokens are not cross-instance semantic identity. Target/window admission checks
still enforce their captured lifetimes.

Authoring is separate and opt-in: `go test -tags offscreen ./workspace -run
'^TestAuthorWorkspaceRecords$' -args -workspace-record-out=/tmp/new-record-directory`.
The destination must not exist. Changing golden observations requires reviewing
the model assertions and recorded inputs together; acceptance never updates them.
These are software routing results. Native compilation, physical pointer input,
GPU presentation, cross-display scaling, human accessibility and full P9/P11
acceptance remain separate.
