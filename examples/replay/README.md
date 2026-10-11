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
are changed. Native OS services and custom extensions refuse in this checkpoint.

These tests qualify software rendering and routed widget behavior. Cocoa IME,
external accessibility, GPU presentation and sustained performance have separate
acceptance workloads.
