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

Click canvas objects to select. Drag empty canvas space or use arrow keys to pan; scroll
to zoom about the pointer. Fit and Reset change the view. Toggle grid snapping,
reject all drops, or reject notes to inspect target policy. Undo/Redo buttons and
Cmd+Z / Cmd+Shift+Z replay exactly one placement with the original object ID.
The example limits the document to 512 objects and history to 64 changes.

Drag a placed component to move it. The press offset is preserved under pan,
zoom, scrolling and display scale; snapping rounds the proposed origin to the
20-unit grid. A completed drag is one Undo entry. Escape, focus/capture loss,
release outside the canvas, or a changed camera/display scale cancels the preview.
Canvas origin, visible size and clipping may settle during capture; placement
keeps the initial world grab point and validates release against the latest clip.
Clicking without moving creates no history entry. Empty-space drags still pan.

The Move selected Left/Right/Up/Down buttons (or Cmd+Option+arrow keys) nudge by
20 world units with snapping enabled and by one unit with snapping off.
Delete selected, Backspace, and forward Delete remove the selected object.
Undo restores its exact ID and original paint order; Redo repeats the deletion.
Move/Delete participate in the same saved/unsaved drawing history as placement.

Enable **Resize mode**, then drag a component to change its width/height with
its top-left corner fixed. The initial grab offset is retained through pan,
zoom, display scale and canvas layout changes. Disable Resize mode to move
components again. Empty-space drags always pan. A resize crosses the same
six-logical-pixel threshold and uses the same cancellation/release-consumption
path as moving: Escape, focus/capture loss, outside release, camera/display-scale
change or Close restores the original dimensions without a history entry.
Each valid completed resize makes one Undo entry, preserving ID, position,
selection and paint order. Width/Height −/+ buttons use the same command/history
path and are keyboard accessible through Tab/Return. With snapping enabled,
changed dimensions round to the 20-unit grid; with it off, buttons change one
unit. Unchanged axes retain their original size. Sizes clamp to 8–2000 units;
objects never flip when the pointer crosses the fixed corner. The selected size
is displayed beside the buttons. This app uses a mode rather than corner handles.

Saved drawings now use version 2, which preserves component dimensions. Existing
version-1 drawings open at their original type dimensions and are upgraded on
Save. Older builds cannot open version-2 drawings. Invalid/missing dimensions
in version 2 are rejected before replacing the drawing.

Use **Save** (Cmd+S), **Save As…** (Cmd+Shift+S) and **Open…** (Cmd+O) to
persist drawings through native file choosers. `.vellum` files contain bounded,
versioned JSON for component types, stable IDs, positions and dimensions, selection,
pan/zoom, snapping and the next object ID. Window dimensions and Undo history
are not saved. Opening validates the entire file before replacing the drawing,
clears the previous document's history, and preserves the current canvas size.
New placements after opening still support Undo/Redo.

Open and Close ask before discarding unsaved changes. Cancel Close to save
first. Navigation, selection and snapping do not mark the drawing dirty, but their
current values are included whenever Save is used. Save writes a complete file
atomically; failed/canceled Open keeps the current drawing. Only this example
uses the existing `document.Session` API; framework behavior is unchanged.

Run the model checks or deterministic software routing checks separately:

```sh
go test -race ./...
go test -race -tags offscreen ./...
go vet -tags offscreen ./...
go build .
```

The offscreen suite routes real managed-widget events at owner batches 1, 7 and
64 using bundled, checked fonts. It exercises all component types, keyboard
activation, validation, selection, pan/zoom and Undo/Redo. The application-local Core extension that captures placed-object drags is excluded
from the qualified descriptor runner; its mapping/history have focused model
regressions, and its input path needs the packaged native check. It does not prove
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
Ordinary tests use the upstream offscreen driver and never open its window. A ten-minute deadline and a session-owned `stop-live` marker in the private
output directory end the session. The test controller must create this marker
when desktop permission is withdrawn or the user returns. Historical markers
from other sessions are kept intact and do not control a new probe. The external operator supplies all pointer/key actions.
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

The [first native desktop attempt](../../docs/reports/macos-placement-desktop/README.md)
reached one completed native accessibility snapshot, then expired while external
app binding was pending. No pointer/key action ran; native placement workflow
acceptance remains open. The offscreen/race checks and failed native deadline are
reported separately, with source/build pins and owned-process retirement evidence.

The [fresh session-owned probe check](../../docs/reports/macos-placement-session/README.md)
passes cancellation-isolation/routing race tests and reaches native window/AX
readiness, but the bounded PID-directed click/Return attempt creates no object
and fails its observation bound. It is sealed without retry. Native placement
needs direct/manual verification; posting receipts do not prove app delivery.
