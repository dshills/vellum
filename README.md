# Vellum

Vellum is an experimental Go-native professional UI framework. Applications use
ordinary Go values, stable widget IDs, explicit commands and cancellable tasks.
A separate native driver reuses Cogent Core's windowing and WebGPU stack.

The framework implementation starts here; `prototypes/` preserves research and
foundation evidence. Production readiness, native accessibility, international
input, and Windows/Linux runtime validation remain open. System input sources
are not changed by the workbench.

## Build and run

The root module is dependency-free and locally named `example.com/vellum` until a
publication namespace is chosen. Run `go test -race ./...` here. The native driver
is a separate module so logic tests do not initialize an OS or GPU runtime.

Bootstrap the pinned, hash-checked patched dependencies, then build:

```sh
cd drivers/cogent
export P0_CACHE_DIR=/tmp/vellum-p0-cache
./prepare.sh
. ../../prototypes/foundationintegration/env.sh
go run ./cmd/workbench
```

The driver module uses repository-relative replacements for the prepared,
reviewed Cogent and go-text copies. It requires Go, a C compiler, the platform's
native graphics libraries, and a supported GPU. Current runtime evidence covers
macOS only. The initial adapter supports one native window per process and must
run from main. Additional windows and platform adapters remain future work.

`app.Application.Post` provides bounded worker delivery with explicit refusal.
Application state changes in `Do`, posted callbacks, or UI actions; views run
under the same serialized owner. Calls to `Do` and `Drain` are not reentrant.
Views only describe UI; they do not start workers or mutate domain state.

`drivers/cogent/prepare.sh` downloads exact pseudo-versions, verifies patch input
SHA-256 hashes, then applies the committed patches to owned source copies.
`scripts/check.sh` explicitly requires that bootstrap. The macOS CI workflow
runs bootstrap from a fresh checkout before tests and the native build. This is
a repository-local development distribution, not yet a published Go dependency.

Queue capacity bounds pending callbacks, plus the one callback currently executing.
Driver wake functions are nonblocking notifications and never enter the UI owner.
Native wakes use Cogent's lossless custom-event deque; they are coalesced while
one wake is outstanding. A panic closes the application and propagates.

Commands live in an owner-thread registry. Availability is rechecked on invocation
from menus, toolbars, shortcuts and the palette. `tasks.Start[T]` accepts a
context, an ordinary Go worker and a typed completion callback. Worker count and
undelivered results share one bound. Progress is coalesced; completion waits for
mailbox credit or lifetime cancellation. Workers cannot mutate live UI state.

The `workbench` package owns a validated split/tab tree. Move, split and restore
are atomic model operations; malformed or incompatible layouts leave the current
workspace intact. The example provides dock/split/reset commands and resizable
native panes. `-layout /path/layout.json` enables explicit Save/Restore; startup
does not overwrite preferences. Drag-to-dock, floating panes and dynamic window
creation are not yet implemented. Panel topology changes recreate those native
panels; application domain state remains independent.

The example generates one million stored component records and 100,101 tree
nodes, with 20 table-row and 24 tree-row controls. Wheel scrolling and paging
materialize the current viewport; stable domain IDs connect selection to the
inspector. Table sorting runs against an immutable snapshot in a bounded task;
revision checks discard stale results. Tree expansion rebuilds a lightweight visible-node index; subsequent viewport
materialization takes O(viewport) work even at deep offsets.
Domain records, tree indexes and optional sort indexes retain O(N) memory. The
viewport bound is not a claim that all application memory is independent of N.

`-tree-rows` controls the outline's explicit subset of the component dataset
(default 100,000); its label shows the subset and full table count. Versioned
sources expose `Revision()` so same-length edits invalidate active projections.
Providers without revisions explicitly reset projections after edits. Source
growth preserves the viewport offset, clamping only when necessary.

Tree expansion rebuilds a lightweight O(N) visible-node index; scrolling materializes O(viewport) records. Worker projections carry source revision metadata and require an owner-captured table token when installed. Background sources must be immutable snapshots. The Cancel command also cancels sorting.

`canvas.Scene` owns explicit rectangle objects, viewport transforms and picking.
The Cogent `GPUCanvas` adapter uses its public `xyzcore.Scene` direct renderer
on the shared WebGPU device, with one reusable mesh and retained instances.
Wheel zoom anchors the pointer; dragging pans; clicking selects a component in
the inspector; Fit/Zoom commands share the command registry. Native scene
destruction releases GPU resources. The current canvas is a bounded rectangle
renderer, with a 10,000-object limit; arbitrary shaders, image tiles, spatial
indexes/LOD and device-loss recovery remain future work. Runtime verification
covers macOS; the integration is not a production performance certification.
