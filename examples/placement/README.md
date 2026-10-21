# Palette-to-canvas placement

An independent public Go SDK consumer demonstrating a reusable typed tool palette
and a managed `ui.Canvas2D`. Block, Connector and Note use the same validation,
command registry and bounded history path for pointer and keyboard placement.

The right-hand **Layers** panel lists components front first. Group selectors
appear before each group's frontmost member; component rows retain the exact
paint order even when flat group members are interleaved. Click a component or
group selector to select its whole group. **Add selection** makes clicks toggle
components/groups in the current selection; **Clear** deselects everything.
Checkmarks/highlights follow canvas selection and Undo/Redo. Selection is transient
and does not add history or dirty the drawing. **Forward / Backward** in the panel
use the same one-layer commands as the toolbar; absolute Front/Back actions remain
in the toolbar. Reorders are one Undo/Redo entry and persist through Save/Open.
The list scrolls independently of the canvas and supports keyboard button focus.
Group selectors are membership markers, not synthetic paint layers; there is no
nested layer tree.

Drag a component row or group selector past the 6-point movement threshold to
reorder. Dragging a selected row moves the complete current selection; an unselected
row moves its component/whole group. The upper/lower half of a visible target row
means above/below it in this front-first list. A blue insertion line
identifies the proposed position; release commits exactly one history entry. A group
selector targets its outer group boundary; a member row targets that component.
The group boundary indicator clamps to the visible list edge when its last
member is offscreen; the drop still targets the entire group boundary.
A no-op adds no history. Short clicks keep normal selection/Add behavior. Escape,
outside/invalid/self release, focus/capture loss, resized panel geometry or display
scale, removed rows and window closure cancel without reordering or changing
selection. The final release is consumed.

While holding a threshold-crossed row drag, use the wheel/trackpad to scroll the
layers list (including with Shift held). Hold near the top or bottom inside edge
to autoscroll; moving back to the middle or outside the list stops autoscroll.
Only the list scrolls; canvas pan/zoom and the drawing stay unchanged until the
drop. Partially clipped source/target rows use their full row midpoint, so the
visible portion does not change above/below semantics. Sources can scroll out of
view without losing their stable identities. Held drags suppress row tooltips.
Scrolling itself adds no history; the eventual valid release adds one entry.

The native adapter owns one bounded 50ms autoscroll pulse at a time and retires
its timer and reserved delivery on finish, cancellation or destruction. Native
scrolling rebases only verified vertical translations of every row; a structural
layout change still cancels. Cross-window dragging remains unsupported. Existing
reorder buttons remain the keyboard path to the same bounded document history.

Use **Bring to front** or **Send to back** to place selected components and
whole selected groups above or below every unselected component. Each action
preserves relative paint order within the selection and within the remaining
objects, along with geometry, group membership and the primary selection.
A changed order is one Undo/Redo step; an already front/back selection adds no
history. Save/Open preserves the stacking order. Groups remain flat membership
sets: this does not create nested layers or change their internal paint order.

Use **Bring forward** / **Send backward** for one component layer per action.
Every member of a selected group participates. Each consecutive selected run
crosses one adjacent unselected component, preserving relative order within both
sets. Disjoint selections step independently; runs already at an edge stay there.
Groups are flat membership sets, so members can remain interleaved with other
components; unselected groups are not synthetic layers. The commands never
compact unrelated groups. A changed order is one Undo/Redo entry; an edge or
all-selected no-op adds no history. Save/Open preserves the resulting order.

Use **Group / Cmd+G** to bind at least two selected components or groups into
one flat group. **Ungroup / Cmd+Shift+G** removes membership from selected groups,
leaving geometry, IDs and the selected components unchanged. Each is one Undo
entry and marks the drawing dirty; Undo/Redo restore the exact memberships and
selection. Grouping existing groups merges them; nesting is not supported.

Click or keyboard-select a member to select its whole group. Shift-click toggles
the whole group; marquee contact with any member selects the whole group. Empty
space still pans/clears selection normally. Group members move, resize and delete
together. Copy/Paste/Duplicate preserve groups but allocate independent group IDs
for the new components. Undoing deletion restores both geometry and membership.
Ungroup first to edit a single member. Named accessibility objects include their
group ID; the managed canvas still exposes one primary member.

Alignment/distribution counts groups as units and translates members rigidly.
One selected group alone has nothing else to align with; select another component
or group. Handles resize the complete selected set as before. Groups are metadata,
not extra canvas objects, so object counts and paint order stay unchanged.

Drawings that have allocated groups save as app format **version 3**, including
membership and the group-ID counter even after all groups are removed. Version-1/2
drawings still open; never-grouped drawings still save as version 2. Older internal builds cannot read grouped version-3 files. On Open,
a saved primary belonging to a group selects that whole group, and history resets.
Up to 256 disjoint groups of at least two members fit within the 512-object limit.

Use the **Arrange** toolbar with at least two selected components or groups to align left,
horizontal center, right, top, vertical center or bottom to the selection bounds.
**H gaps / V gaps** require at least three components or groups and distribute them with
equal horizontal/vertical edge gaps inside the same bounds. Components are
ordered by their current axis position; ties keep paint order. Distribution
requires enough room for non-overlapping gaps; a crowded set is left unchanged
with an explanation. Move components apart first if needed.

Arrangement changes only the chosen axis positions, preserving IDs, dimensions,
paint order, selected set and primary. It ignores grid snapping to keep alignment and gap relationships to floating-point precision, regardless of pan/zoom. Each action is one Undo
entry; an unchanged result adds none. Invalid origins reject the whole action.
Existing busy/file/gesture guards apply. Undo/Redo restore exact geometry.

Copy selected components with **Copy / Cmd+C**, paste with **Paste / Cmd+V**,
and duplicate with **Duplicate / Cmd+D**. Copy writes versioned Vellum component
JSON as plain text to the native macOS clipboard. Paste accepts that format;
unrelated text or invalid data leaves the drawing unchanged. Duplicate does not
read or change the clipboard. Shortcuts defer to focused native controls first.

Each insertion preserves component types, sizes, colors, relative positions and
paint order, assigns fresh IDs and selects the inserted set. The first Paste
shifts the copied world positions by (+20,+20); repeated pastes of identical
clipboard text use (+40,+40), (+60,+60), and so on. Copy resets that sequence. The counter counts successful pastes (up to 1000);
Undo and Duplicate do not reset it. Origin limits can reject an insertion sooner.
Duplicate shifts the current selection by (+20,+20). These are shared world-unit
translations regardless of snapping or camera; members are never individually
snapped. Use Fit if copies lie outside the visible canvas.

A completed Paste/Duplicate creates exactly one history entry. Undo restores the
prior selection and geometry; Redo restores the same inserted IDs. New edits
after Undo allocate fresh IDs rather than reusing discarded ones. Copy creates
no history and does not mark dirty. Save/Open retain inserted geometry through
the version-2/3 drawing format; transient paste counters reset on Open.

Clipboard data is bounded to 256 KiB and 512 components. The existing document
capacity, origin/dimension limits and placement rejection switches apply to all
members atomically. Clipboard work runs outside the UI owner, with one operation
at a time, a five-second deadline and window-scoped result delivery. Existing
edit/file/navigation actions are gated while it is pending. Close cancels and
retires pending delivery so a late Paste cannot mutate the drawing. As with the
existing native clipboard service, cancellation can race an OS write already
in progress. There is no interoperability with other shape formats.

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

Click an object to select it. Shift-click toggles a component or its group in a bounded,
paint-ordered selection set; Shift-click empty space keeps the set. Blue outlines
mark every selected component and the toolbar/status show the selected count.
Selection changes alone do not make Undo entries or mark the drawing dirty.

Shift-drag from **empty canvas** to draw a marquee. After the six-logical-pixel
threshold, a blue rectangle previews the selection area. Release inside the
canvas to add every component whose footprint intersects the rectangle with
positive area, including partially covered or overlapping components. Either
drag direction works. Existing members and the existing primary are kept;
otherwise the last selected object in paint order becomes primary. Shift-drag
starting on a component keeps Shift-click toggle behavior; start in empty space.
Click empty space first for a fresh selection. A zero-area rectangle or empty
Shift-click keeps the previous set. Snapping does not affect the selection area.

Selection commits only on valid release and creates no object edit or Undo entry.
Escape, outside release, focus/capture loss, Close, or camera/display-scale change
cancels and keeps the original selection. The captured release cannot also pan,
click a component or activate a palette tool. Plain empty-space drag still pans,
plain arrows pan and wheel zooms. During a marquee, navigation/file/edit commands
are gated by the existing capture lifecycle. There is no edge autoscroll.

Drag any selected component to move the whole set, preserving relative offsets.
With snapping, the dragged component determines one snapped translation shared by
all members. A group drag or Move button/shortcut is one Undo entry; if any member
would leave ±10000, no member is committed. Escape, focus/capture loss, outside
release and camera/display-scale changes cancel the entire preview. The captured
release cannot also trigger a selection click or canvas pan. A plain click with
no drag replaces the set with the clicked component or its group. Clicking an unselected
component replaces the set immediately; clicking empty space clears it and keeps
ordinary pan navigation. Drag empty space or use arrows to pan; wheel zooms.

Delete selected, Backspace or forward Delete removes the complete set in one
history operation. Undo restores exact IDs, positions, dimensions, paint order
and the selection set; Redo repeats the operation. New placement selects its new
object; Undo restores the previous set. Cmd+Z / Cmd+Shift+Z replay history. The
example limits the document to 512 objects and history to 64 changes.

Corner handles surround the selected set's bounding box. Drag any corner to
scale all selected positions, dimensions and spacing, keeping the opposite
bounding corner fixed. Horizontal and vertical scales are independent. With
snapping enabled, changed group bounds round to the 20-unit grid; members are
not individually snapped. Each member stays within 8–2000 units: a shared scale
clamps at the strictest member limit, without flipping or distorting the group's
relative proportions along either axis. Origins must remain within ±10000;
invalid movement cancels the entire edit. One group resize is one Undo entry,
restoring exact geometry, IDs, paint order and selection. A corner click without
a drag keeps the group and adds no history. Escape, outside release and existing
capture/camera/lifecycle cancellation restore every member.

Width/Height buttons still resize exactly one component. Managed keyboard and
semantic object selection replace the set with a primary and its group. Additional
members are described in accessible names; standard selected semantics expose
the primary only. The set is transient: Save retains that primary in the version-2/3
drawing format, and Open selects it and its group. All resized object geometry is saved.

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

Never-grouped saved drawings use version 2, which preserves component dimensions.
Grouped drawings use version 3 as described above. Existing
version-1 drawings open at their original type dimensions and are upgraded on
Save. Older builds cannot open version-2 drawings. Invalid/missing dimensions
in versions 2/3 are rejected before replacing the drawing.

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

## Layer properties (internal app)

Select a row or group marker and choose **Rename…**. Names are app metadata: stable
IDs, paint order and group membership stay unchanged. Empty names restore the ID
label. Names allow at most 80 characters/128 UTF-8 bytes with no controls. Each
changed name is one Undo/Redo entry and persists in app drawing format 4; formats
1–3 remain readable. Deleted/ungrouped metadata is pruned and restored by Undo.
Framework APIs, renderer and clipboard formats are unchanged in this phase.

**Hide / Show** in the Layers panel changes an undoable, saved visibility property.
Hidden objects stay in the document and layer order but are excluded from canvas
painting, picking, keyboard/accessibility traversal and marquee selection. Groups
remain interaction units: hiding/showing a member controls its whole group. A hidden
row can still be focused for Show/Rename and dragged to reorder. Fit uses visible
objects. Canvas navigation continues through the app's visible projection; no
framework APIs or renderer changed.

**Lock / Unlock** persists an undoable protection property. Locked objects remain
painted but cannot be canvas-selected, moved, resized, deleted, grouped or layer-
dragged. Locking a member controls its whole flat group; imported member locks
also protect that group. Layer rows remain focusable for Unlock, Rename and
visibility controls. Locking clears the editable selection; Undo restores it.
Unlocked components may still be reordered around locked layers. No framework
behavior changes.

**Cut / Cmd+X** writes the selected components to the native clipboard first, then
removes the original selection in one Undo/Redo entry. A failed, canceled or stale
clipboard operation keeps the drawing. Cut shares Copy/Paste validation and
retains component/group names, geometry and membership; Paste allocates fresh IDs.
Native focused text fields retain their standard Cut shortcut.

## Private recovery (internal macOS app)

While a drawing is dirty, the app checks for settled changes every 15 seconds and
writes a bounded private recovery copy to
`~/Library/Caches/VellumPlacementInternal/recovery.vellum`. This never changes the
Save path or marks the document saved. After an interruption, startup offers
**Restore Drawing**, **Discard Recovery**, or **Close; Keep Recovery**. Restore
creates one undoable edit and remains untitled/unsaved until an explicit Save.
Names, visibility, locks, grouping, geometry and camera state are retained.

A clean close or an explicit discard-and-close removes recovery after joining
the worker. Failed operations keep the last completed copy. Capture is attempted every 15 seconds; recovery retains the last completed copy.
Slow or failed storage and busy gesture/file/clipboard work can defer a newer copy. Recovery is not a replacement for Save or a document version log.
One app instance owns the private slot at a time; another instance stays editable
with recovery unavailable. macOS is the recovery target for this internal build. Shutdown joins local file
IO before releasing ownership; stalled storage can delay shutdown.
Framework scope and experimental rendering/pacing settings remain unchanged.

## Recent drawings (internal app)

**Recent** lists the last ten successfully opened or saved drawings, newest first,
with their locations. The bounded private list survives relaunch. A missing or
invalid drawing leaves the current document unchanged; opening a recent drawing
uses the same unsaved-change prompt as Open. Cancel keeps the drawing and clears
the pending destination. Save As suggests a base name without repeated `.vellum`
suffixes; the OS-confirmed destination is never renamed by the app.

## Compact command menus

Edit, Move, Resize, Arrange and Layers menus retain the existing commands and
shortcuts. Open/Save/Recent, Undo/Redo, Fit/Reset and Snap stay directly available.
Drop-rejection test switches are beside the palette. The compact controls replace
four wide command rows to leave more vertical canvas and layer-list space.

## Numeric geometry

**Geometry…** edits the current selection's world X/Y and bounding width/height.
Groups and multiple selections scale members and spacing together. Exact numeric
values ignore Snap. Apply validates every component before one Undo/Redo change;
invalid, stale or canceled drafts keep the drawing. IDs, order, names, visibility
and locks remain intact. Individual sizes remain 8–2000; origins remain ±10000.

## Aspect-locked resizing

**Lock aspect ratio** preserves one uniform scale for a selected component or set.
Alternatively hold Shift when pressing an existing selected corner handle; that
gesture locks the ratio without changing the checkbox. Shift elsewhere retains
selection behavior. The captured lock stays fixed through the gesture. Only the
driving dimension snaps; all members share the strictest size/origin limit, keep
the opposite corner and do not flip. One committed resize is one Undo/Redo entry.
The numeric inspector derives the other dimension when one size changes; incompatible
explicit sizes are rejected. Step resizing of a single component also respects lock.
The checkbox is session state and does not change saved drawing formats.

## Subtractive selection

Enable **Subtract selection**, then Shift-click a selected component or Shift-drag
from empty canvas to remove intersected components and whole flat groups from the
selection. A touch with any member removes that complete group. The rectangle
uses positive-area intersections in either drag direction. Subtraction never adds
components; the remaining set stays in paint order, keeping the primary if it
survives. Disable the checkbox to restore additive Shift-marquee/toggle behavior.
Selection changes do not dirty the drawing or add history. Escape, invalid/outside
release and capture/camera/lifecycle cancellation keep the original marquee set.
Plain empty drag still pans and the wheel still zooms. Shift on a selected corner
handle retains temporary aspect-lock resizing. Mode/group membership are captured
at press; preferences cannot change a gesture midway. The mode is session state.

## New drawing (internal 19)

**New / Cmd+N** starts an empty, untitled drawing. Unsaved changes require an
explicit Discard and New decision; Cancel keeps the document and history. New
resets history, selection, groups, camera and the Save destination, like Open
replaces document history. Existing files and Recent entries are kept. The next
Save asks for a destination. Framework scope remains frozen.

## Selection commands (internal 19)

The Edit menu offers **Select All / Cmd+A**, **Deselect All / Cmd+Shift+A**
and **Invert Selection**. They operate on all visible, unlocked components,
regardless of viewport position, and keep flat groups whole. Selection-only
changes do not enter Undo history or mark the drawing dirty. Focused text
fields retain their standard text-selection shortcuts.

## Explicit zoom controls (internal 19)

**Zoom** offers Zoom In, Zoom Out, 100% and Fit Selection. The adjacent percentage
reflects the current view, including wheel zoom. Menu zoom keeps the world-space
view center; Fit Selection includes complete selected groups. Fit continues to
fit all visible objects; Reset returns to origin at 100%. Navigation changes no
object geometry or Undo history, and is blocked during gestures and dialogs.

## Grid controls (internal 19)

**Grid** shows/hides the grid independently of **Snap**. The Grid spacing menu
offers 5, 10, 20, 50 and 100 world units. The selected interval applies to placement,
move and corner/step resizing, including proportional resizing; numeric geometry
remains exact. Edits keep their usual one-step Undo/Redo. Visibility and spacing
are session preferences (default visible, 20); existing saved Snap behavior stays
compatible with earlier drawings. Settings cannot change during a gesture.

## Layer filtering and collapsed groups (internal 19)

**Filter layers…** searches names and stable IDs, ignoring case (128 characters
maximum). A member match retains its group marker; a group-name match reveals
all members. **Groups** offers per-group Collapse/Expand and Collapse All/Expand
All. Filtering reveals matches even inside collapsed groups; Clear filter restores
the previous collapse state. Collapsing does not change flat group membership.

Filtering and collapsing change only panel rows, never canvas visibility,
selection or history. Clicks still select whole groups. Drag reordering is disabled
while filtering; clear the filter to drag. Collapsed group markers remain valid
whole-group drop targets. Projection settings are session-only, clear on New/Open,
and cannot change during a gesture.

## Descriptive history (internal 20)

Undo and Redo include the next operation's journal label, for example **Undo Move
selection** or **Redo Resize selection**. Empty stacks show plain disabled Undo
or Redo. Command identities and keyboard shortcuts are unchanged; captions follow
replay, new history branches, and New/Open history resets.

## Exact nudges (internal 20)

**Nudge** offers Left/Right/Up/Down at exactly 1 or 10 world units, independent
of Snap and grid spacing. Complete selected groups move rigidly. Each accepted
action creates one Undo/Redo entry; a bounds violation rejects the entire move.
The existing Move menu and Cmd+Option+arrow shortcuts retain grid-aware movement.

## Paste at View Center (internal 20)

**Edit → Paste at View Center** centers the copied set's complete bounds on the
current world-space view center, preserving relative positions, group membership
and names with fresh IDs. It bypasses snapping and creates one Undo entry; Redo
restores the same IDs. Invalid/out-of-bounds payloads are rejected atomically.
Existing Paste keeps its offset cascade; centered Paste does not advance it.

## Save a Copy (internal 20)

**More → Save a Copy…** writes a detached drawing snapshot to a separate file.
The active filename, Save destination, dirty state, selection and Undo history
are unchanged. The native chooser confirms overwrites; choosing the active file
or an existing alias of it is refused. Successful copies appear in Recent.
Cancel or failed storage keeps the drawing. IO uses the existing bounded document
worker and atomic replacement; closing during output first cancels that operation.
