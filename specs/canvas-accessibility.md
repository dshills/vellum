# Named Canvas2D objects

`ui.Node.CanvasObjects` optionally supplies `ui.CanvasObject{ID, Name, Help}`
for existing `canvas.Scene.Objects`. IDs are stable scene identities, independent
of names and paint order. A name is required; help is optional. Labels may repeat.
Keep each ID attached to the same domain entity. Removal observed by the adapter
retires its action; identical removal/reinsertion entirely between observations
cannot be detected. Apps own the scene and metadata on their application owner.

The metadata is a subset of the scene, at most 10,000 descriptions and 1 MiB of
combined ID/name/help bytes. IDs are at most 256 bytes, names 1,024 and help 4,096;
all are valid UTF-8 without NUL. Scene count/geometry validation precedes lookup
allocation. Missing/duplicate identities, blank names and excessive input refuse.
Managed node IDs reserve the `canvas-object:` prefix for generated targets.
An empty list preserves the generic canvas host and existing navigation.

Each named rectangle intersecting the actual completed canvas clip becomes a
button-like accessibility child. Geometry comes from the same detached picture
and completed paint as its selection, viewport, origin, dp scale and ancestor
scroll clipping. Decorations, grid lines and placement previews have no object
semantics. Traversal and native hit testing preserve scene paint order regardless
of descriptor order. Overlapping children remain individually selectable by ID;
physical point selection still chooses the topmost painted object. Offscreen
objects have no children or paging actions; pan/zoom to expose them.

`canvas.Scene.ObjectAt` is a pure reverse-paint-order hit query. `Scene.Select`
selects an existing stable ID (or clears with an empty ID), refusing unknown IDs
without changing selection. Managed pointer, object accessibility press and
keyboard traversal use this same owner-held selection helper and deliver one
`Change` notification per accepted action. Selection remains single-object and
is separate from placement/history; apps decide whether to journal it.

The canvas retains one ordinary focus stop. Option/Alt+Left/Right cycles its
visible named objects; Option+Home/End selects the first/last in scene order.
These commands require Alt without Control/Meta/Shift. Plain arrows retain pan;
other modified arrows retain the existing canvas handler behavior. A semantic
object press focuses the host through the existing namespaced Workbench route.
Objects do not create per-object native widgets or extra Tab stops.

Object actions retain their originating completed frame, binding, input scope
and pointer-capture epoch through queued owner delivery. Pending layout/model or
label changes, disabled/hidden/modal transitions, observed removal/replacement,
retained ancestor/dialog/Workbench transfers, capture begin/loss and window
closure refuse stale work. Transfer withdraws the frame before panel callbacks.
A byte-identical new picture does not repeatedly invalidate an idle window.

Canvas semantics and frames share the existing bounded pending/completed scope
transaction. Its legacy wire field names remain `grid_*`; the native validator
allows only table, outline and explicit canvas hosts. Pending publication retains
only copied last-completed values/frames, disabled and actionless, with no action
closures in the cache. Real scope retirement removes it. Ready completion restores
current actions and atomic frames; unversioned geometry cannot overwrite them.

`cogent.ManualWindowSnapshot.Canvases` returns detached `CanvasInspection` values,
including separately copied `Objects`, completed viewport, clipped window-pixel
bounds, generation/incarnation and total named count. `CanvasObjectID` builds an
object transport target from host ID and scene ID; it is not an action capability.
Workbench qualification prefixes targets just as for managed grid/tree targets.
Canvas logical coordinates use dp (DPI/160); window pixels and canvas world units
must not be interchanged with the grid's logical px (DPI/96).

The independent placement demo derives names from its objects and supports the
same pointer placement, Undo/Redo and keyboard selection through public APIs.
Automated offscreen routing and inert Cocoa element fixtures do not qualify
physical pointer delivery, live AppKit attachment, cross-display transforms or
human VoiceOver. This rectangular retained Canvas2D contract does not add GPU
MeshCanvas semantics, multi-selection, arbitrary path geometry or offscreen
accessibility navigation. Same-native-window palette placement limitations remain
in [placement](placement.md). Full P3/Mac acceptance remains open.
