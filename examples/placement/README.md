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

Click an object to select it. Shift-click toggles a component in a bounded,
paint-ordered selection set; Shift-click empty space keeps the set. Blue outlines
mark every selected component and the toolbar/status show the selected count.
Selection changes alone do not make Undo entries or mark the drawing dirty.

Drag any selected component to move the whole set, preserving relative offsets.
With snapping, the dragged component determines one snapped translation shared by
all members. A group drag or Move button/shortcut is one Undo entry; if any member
would leave ±10000, no member is committed. Escape, focus/capture loss, outside
release and camera/display-scale changes cancel the entire preview. The captured
release cannot also trigger a selection click or canvas pan. A plain click with
no drag replaces the set with the clicked component. Clicking an unselected
component replaces the set immediately; clicking empty space clears it and keeps
ordinary pan navigation. Drag empty space or use arrows to pan; wheel zooms.

Delete selected, Backspace or forward Delete removes the complete set in one
history operation. Undo restores exact IDs, positions, dimensions, paint order
and the selection set; Redo repeats the operation. New placement selects its new
object; Undo restores the previous set. Cmd+Z / Cmd+Shift+Z replay history. The
example limits the document to 512 objects and history to 64 changes.

Corner handles and Width/Height buttons resize exactly one selected component;
group resizing remains outside this internal app. Use Shift-click to reduce a
set to one, or plain-click a component. Managed keyboard/semantic object selection
still replaces the set with a single primary object. Additional group members
are described in accessible names; standard canvas selected semantics expose the
primary only. The set is transient: Save retains the primary in the existing
version-2 format, and Open clears the group and selects that primary. All group
object edits and dimensions are saved normally.

The press offset is preserved under pan, zoom, scrolling and display scale.
Canvas origin, visible size and clipping may settle during capture; movement
keeps the initial world grab point and validates release against the latest clip.
The Move selected Left/Right/Up/Down buttons (or Cmd+Option+arrow keys) nudge the
whole selection by 20 world units with snapping enabled and one unit with it off.
Move/Delete participate in the same saved/unsaved drawing history as placement.

Select a component to show four blue-and-white corner handles. Drag any handle
to resize directly; no mode switch is needed. The opposite corner stays fixed,
including when minimum/maximum size clamps apply. Drag inside the component to
move it. Empty-space drags always pan. Handles stay the same logical pixel size
under zoom and display scaling, and paint/hit-test using the current canvas
transform and clip. Only visible corner centers are exposed; use Fit to bring
an offscreen corner into view. Handles take priority over overlapping object
fills. Very small projected objects have overlapping hit targets; the nearest
corner wins, so zoom in to distinguish them.

Each resize crosses the same six-logical-pixel threshold and uses the existing
capture/cancellation/release-consumption path. Escape, focus/capture loss, outside
release, camera/display-scale change or Close restores the original rectangle
without an edit. Each valid completed resize is one Undo entry, preserving ID,
selection and paint order. Undo/Redo restores position as well as dimensions for
north/west handle edits. Corner origins must remain within ±10000. An invalid coordinate-limit preview
keeps the last valid rectangle visible; releasing there cancels the entire edit.

Width/Height −/+ buttons remain available through Tab/Return. They keep the
top-left corner fixed. With snapping enabled, changed dimensions round to the
20-unit grid; with it off, buttons change one unit. Unchanged axes keep their
original size. Sizes clamp to 8–2000 units and do not flip when a handle crosses
the opposite corner. The selected size is displayed beside the buttons. Handles
are presentation-only, are never saved or counted as components, and the mount
removes their single sprite on destruction. A coalesced native post-layout hook
restores the detached sprite after window resizing or a Core rebuild resets it. Rendering reads detached owner
snapshots and native geometry, not mutable application state.

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
