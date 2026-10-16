# Palette-to-canvas placement

An independent public Go SDK consumer demonstrating a reusable typed tool palette
and a managed `ui.Canvas2D`. Block, Connector and Note use the same validation,
command registry and bounded history path for pointer and keyboard placement.

From the repository root, prepare the pinned native dependencies:

```sh
./drivers/cogent/prepare.sh
. ./drivers/cogent/env.sh
cd examples/placement
go run .
```

Drag a tool at least six logical pixels, then release inside the canvas. Green
feedback shows an accepted proposed footprint; red feedback shows rejection.
Escape, release outside, focus loss or removing a source/target cancels without
creating an object. A palette click only focuses its button. Tab to a tool and
press Return to place it at the current canvas center through the same path.

Click canvas objects to select. Drag the canvas or use arrow keys to pan; scroll
to zoom about the pointer. Fit and Reset change the view. Toggle grid snapping,
reject all drops, or reject notes to inspect target policy. Undo/Redo buttons and
Cmd+Z / Cmd+Shift+Z replay exactly one placement with the original object ID.
The example limits the document to 512 objects and history to 64 changes.

Run the model checks or deterministic software routing checks separately:

```sh
go test -race ./...
go test -race -tags offscreen ./...
go vet -tags offscreen ./...
go build .
```

The offscreen suite routes real managed-widget events at owner batches 1, 7 and
64 using bundled, checked fonts. It exercises all component types, keyboard
activation, validation, selection, pan/zoom and Undo/Redo. It does not prove
physical Mac pointer delivery, display transitions, GPU presentation or external
screen-reader interaction. These require a separate available desktop session.

The first implementation supports palette and canvas in the **same native
window**. It does not transfer captures or payloads between native windows or
applications, use OS drag sessions, or offer cross-window autoscroll. Managed
Canvas2D uses logical view units; the older GPU MeshCanvas API uses its existing
physical viewport contract. See [the framework contract](../../specs/placement.md).

Placed objects expose stable accessible names through `ui.Node.CanvasObjects`.
Focus the canvas and use Option+Left/Right to cycle visible named objects, or
Option+Home/End for the first/last in paint order. Ordinary arrows pan. Object
selection uses the same stable-ID owner path as a semantic accessibility press;
it does not add an entry to the placement history. Offscreen objects must be
brought into view before this traversal. See the
[canvas semantics contract](../../specs/canvas-accessibility.md).

For a separate Mac desktop-pointer acceptance session, build the opt-in probe:

```sh
go build -tags vellum_placement_probe -o /private/tmp/vellum-placement-probe .
mkdir -m 700 /private/tmp/vellum-placement-probe-output
VELLUM_NATIVE_PLACEMENT_PROBE=/private/tmp/vellum-placement-probe-output \
  /private/tmp/vellum-placement-probe
```

The output directory must be new, empty, private and free of symlink components.
This dedicated executable follows ordinary desktop initialization on AppKit's
original locked thread and refuses inconsistent/offscreen app identities.
Ordinary tests use the upstream offscreen driver and never open its window. A ten-minute deadline and the shared human-return
marker end the session. The external operator supplies all pointer/key actions.
The probe observes the ordinary view/model, command/history, canvas-change and
palette-activation counts, bounded gesture state and actual Cocoa geometry.
It writes atomic `latest.json` snapshots and at most 8 MiB of `events.jsonl`.
Three diagnostic buttons arm removal of the source, target or window only after
observing a threshold-crossed native gesture. Restore bindings before another
case. Closing the native window joins the observer and records final cancellation
and reservation state. Observer failure is retained; if it cannot join, the
process fails without running shared teardown and the outer owner must reap it.
Window-geometry persistence is locally disabled throughout.

These sampled observations do not measure performance or establish physical
hardware input, continuously visible previews, cross-display behavior, or native
heap reclamation. Keep desktop-automation actions and their verified outcomes
separate from the deterministic software-routing suite above.
