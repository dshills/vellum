# Palette-to-canvas placement

This contract defines reusable application-owned palette placement. The initial
adapter supports one native window per gesture. A cross-window target never
commits. This is distinct from OS file drops and dock-tab rearrangement.

`placement.Controller[T]` owns an immutable catalog of up to 256 items and up to
sixteen targets, with at most one active gesture. Item IDs and target IDs are
stable nonempty UTF-8 identities. Payloads retain their Go type through preview,
validation and command delivery. Reference-bearing payload values remain
caller-owned and immutable; the count limit does not bound their allocation size.
Sources/targets are opaque handles borrowed by managed UI records. The caller
closes the controller when retiring its palette; retiring a native binding only
cancels the gesture borrowing that binding.

Pressing a palette source captures the primary pointer. Movement must cross an
explicit pixel threshold before showing a drag preview or accepting a target.
The native router supplies current measured content geometry and a visible clip.
An unclipped content origin can lie outside that clip when scrolled. Pointer and
clip coordinates are window pixels. `Geometry.Scale` converts pixels to logical
canvas units once, then the existing `canvas.Viewport.World` applies pan/zoom.
Managed canvas viewports and their zoom use these logical canvas units. Bounds
at the far clip/content edges are excluded. Native display scaling and actual
scroll clipping must be tested separately from portable transform arithmetic.

The target's optional Snap runs before Accept. Both callbacks are pure owner-held
queries; reentrant controller command/validation admission refuses. Pointer
movement shows the proposed footprint and accepted/rejected target feedback.
Release evaluates fresh geometry, snapping and acceptance again. It retires the
gesture before invoking its single application Command callback. A repeated
release refuses; an outside/rejected release creates no object. Applications
route Command through their command registry and undo journal. The callback must
apply one complete change or return an error without changing the document.
Undo removes that placement; Redo reuses its captured object identity and bytes.

Escape, capture loss, removal/replacement of source or target, modal/visibility
exclusion, and window closure cancel pending placement. Cancellation permanently
retires its gesture and clears its payload, candidate, target and callback
references. The native capture must also clear its widget references/previews
and consume the pending release, so it cannot become an unrelated click,
selection, pan or drag completion. No hidden worker, timer, retry or clipboard
operation participates in placement.

`Source.PlaceKeyboard` takes an explicit target and world position and uses the
same Snap, Accept and Command callbacks. Palette controls expose this path through
ordinary focusable button activation; applications choose the insertion position
(for example the canvas center). Queued keyboard/semantic button actions retain the originating source binding
and visibility lease; hide/replacement/modal transitions permanently revoke them.
Keyboard placement refuses while a pointer
gesture or command is active. A visual preview alone never changes document
state. Selection, panning and zooming retain their existing canvas operations.

The dedicated independent demo `examples/placement` provides several component types, canvas
selection, pan/zoom, optional snapping, and Undo/Redo. Its public API consumer,
SDK inclusion, portable history/transform tests and actual offscreen routing
checks are checkpoint requirements. Offscreen routing and software image checks
do not qualify physical Mac pointer capture, cross-display movement, external
accessibility or GPU presentation. Physical checks remain deferred while the
human is using the machine. This feature does not complete P9/P11 or the wider
Mac acceptance gates.

## Public integration

Create one `placement.Controller[YourComponent]` with an immutable catalog and a
synchronous `Command(ctx, Request[YourComponent])` callback. Register a target
with `NewTarget(TargetConfig[YourComponent]{ID, Snap, Accept})`. Bind a source
handle to a normal `ui.Button.PlacementSource`; its `Activate` callback calls
`Source.PlaceKeyboard` with a live window context, target and world position.
Bind a `canvas.Scene` and target to `ui.Node{Kind: ui.Canvas2D, Canvas: scene,
PlacementTarget: target}`. `CanvasGrid` is optional world-space grid spacing.
The scene uses existing rectangular objects, selection and viewport navigation;
applications retain domain geometry and history ownership. `Change` observes
selection in `SelectedID` and zoom in `Number`, after native owner-held edits.

Keep each Scene, Source and Target in at most one live native binding. Bindings
are bounded to 512 per native window. The managed adapter captures at press and
uses a six-logical-pixel movement threshold. One reserved application completion
per gesture guarantees loss cancellation admission even when the ordinary
event queue is full; if completion capacity is exhausted, the press refuses
and its release is consumed. Painting reads detached object/preview snapshots.
Completed viewport dimensions coalesce into owner feedback, with a root
invalidation fallback when the ordinary queue is full. No motion worker or
automatic clock advance is introduced. Copied/zero transport handles refuse.

Remove the managed bindings before closing the caller-owned controller. Native
retirement cancels but never closes that controller. Model and endpoint access
shares the application's UI owner, including mutation from application commands.
Snap/Accept must validate without partially changing application state; Cancel
and Close during validation are honored, while nested Move/Release/placement
admission refuses. A callback panic propagates according to the surrounding
owner/command contract; it is not silently retried.

Canvas wheel input zooms the canvas; ancestor scrolling can be driven by its
scrollbar. This initial rectangular retained canvas is not a CAD geometry
kernel, arbitrary scene extension or OS drop target. Cross-window capture,
external payload transport and auto-scrolling while dragging are unsupported.
