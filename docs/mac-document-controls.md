# Mac document controls

## Document list

Retain `ui.DocumentList` with up to 256 `DocumentItem` values. Metadata and
immutable original-color images come from the application; no filesystem scan,
preview decoding or file launch happens inside this widget. Columns are name,
kind, modified date and byte size. Use `DefaultDocumentColumns()` for independent
mutable widths (60..600 points). Sort proposals update Sort; folder rows sort
first and text comparison is case-insensitive, stable and not locale-aware.
`Resize` proposes widths through pointer dragging or Left/Right on header grips
(Shift adjusts ten points). The shared horizontal scroller keeps header/row
columns together. Vertical scrolling keeps the header visible.

Selection uses the existing controlled source-list Command/Shift/Command-A
routes; accept with `event.Apply(list.SelectedIDs)`. Return/double-click requests
Activate separately. Disabled/removed items and stale component sources cannot
act. Size is shown in bytes, dates in English. Context menus and automatic
filesystem operations are outside this component. Components → Document list
demonstrates all operations using in-memory sample metadata.

## Mac outline

`ui.OutlineView` composes the existing bounded `virtual.Tree` index with managed
source-list selection. Provide <=256 flat parent-linked `OutlineItem` values,
depth <=16, stable IDs, original-color images, optional badges and disabled items.
Larger indexed datasets retain the existing worker-built `DataTree` route.
Accept Disclosure with `event.Apply(ExpandedIDs)` and selection with
`event.Apply(SelectedIDs)`. Collapsed descendants retain application selection;
only visible selected rows are published to the native list. Disabled ancestors
disable their descendants. Right opens/enters, Left closes/returns to the parent,
Up/Down moves rows and Return/double-click opens a leaf or toggles a branch.
Disclosure also has accessible expand/collapse actions. Optional ContextMenu
reuses existing source-list contextual actions. There is no item dragging,
filesystem traversal, lazy child loading or built-in persistence.

## Native Colors panel

Get `Application.ColorPanels()` and call `ChooseColor` on a worker with an exact
live parent WindowRef and ColorPanelOptions. The actual AppKit NSColorPanel offers
its normal pickers, optional opacity, and a small accessory with Use Color/Cancel.
Apply returns an sRGB straight-alpha `color.NRGBA`; extended-gamut values clamp to
sRGB bytes. Cancel/Escape/close returns ErrCanceled. Preview is optional, runs
serially on that calling worker, and coalesces to one latest value. Publish through
window-scoped posts, guarded by your session epoch; commit only the returned value
through a reserved completion. The gallery demonstrates that separation.

One native document panel is admitted per process. A Colors panel previously
created outside this service is refused with ErrServiceBusy because AppKit exposes
no target/action getters to restore a foreign owner safely. Do not mix direct
AppKit ownership of the shared panel with this service. Parent closure, context
cancellation and shutdown revoke the session, remove observers/child attachment,
hide the panel and restore its prior appearance/accessory state. Hidden/minimized
parents refuse without activating another Space. Each session centers on the
parent and clamps to its display’s visible frame, including offset/negative
display origins. If a display is smaller than AppKit’s minimum panel size, the
minimum is preserved and the top-left/titlebar stays reachable; the body can
extend below/right of that unusually small display. The shared panel’s old position is not reused across parents. Other platforms return unavailable.
Offscreen tests compile the native path and cover request/reply parsing, exact
parent revocation and bounded preview delivery without creating any native panel.
Actual AppKit interaction is checked separately after the final package.

## Native Fonts panel

`Application.FontPanels().ChooseFont` opens the actual AppKit NSFontPanel on a
worker, with an exact parent WindowRef. Options specify an installed PostScript
Name and Size; an unavailable name refuses. The panel exposes installed families,
faces and sizes; effects such as underline, shadow and text color are excluded.
Sizes clamp to 6..144 points, matching the managed FontPicker. The accepted result
includes exact PostScript Name, Family, Size and Bold/Italic descriptor hints.
Use the exact Name for font identity; hints do not describe every variable axis.

Preview, Apply/Cancel, ownership and cancellation follow the Colors service. The
shared NSFontManager target/action/selection and panel accessory/delegate state
are restored after completion when a prior selection existed; first use leaves
the system panel selection cached. A foreign preexisting shared Font panel refuses;
Colors and Fonts sessions cannot overlap. The demo maps accepted choices into
the existing managed FontPicker and previews live family/style/size changes.
Its extra preview caps display size to 72 points while reporting the actual size.
Font rendering may substitute if the text backend cannot resolve an installed
face; no font installation or document/rich-text mutation is performed.

## Circular angle dial

Retain `ui.AngleDial`, accepting AngleEvent by assigning Degrees. Canonical values
are finite [0,360); zero is up and increases clockwise. Step defaults to 1, must
divide 360 and is 0.1..180. Set Size to 64..160 logical points (default 96). A
full-color optional header IconImage is preserved; the dial face uses the scoped
semantic palette or the native scheme. Tick/needle painting uses the existing
vector painter; no new rendering backend is introduced.

Pointer capture converts the actual laid-out circular face coordinates, including
display scale and scrolling. Center dead-zone avoids unstable angles. Release,
capture loss, disabled/hidden scope and window retirement stop the gesture without
a trailing click. Arrows adjust one step; Shift adjusts ten, Home zero and End the
last step. Return/Space and accessibility Increment advance; Decrement and Set are
available through the slider accessibility role. The accompanying numeric field
shares normalization (360 wraps to zero). Proposals are live; applications own
undo transactions and acceptance/rejection. The gallery shows 64/96/128pt variants.

AppKit can show a newly created shared panel using remembered visibility. The adapter takes ownership during that first creation; already existing foreign panels and visible owned panels remain refused.

The widget gallery starts a fresh Cocoa window session using a process-local argument-domain override. It does not erase saved application state or preferences; toolbar customization still uses its own persistence. Cocoa cannot reconstruct an asynchronous chooser's worker, so adapter-owned panels are non-restorable. Apps using the services retain their own window-restoration policy.

Packaged-gallery restoration isolation runs in its native bootstrap before Cogent initializes GLFW in a Go dependency initializer. It is limited to the two widget-demo bundle identifiers.

The Fonts panel reloads its default family list on first creation only. Reusing the existing AppKit panel preserves its internal table views; repeated native reloads caused a macOS 26 deallocated-table crash during reopen checks.

Document activation and outline navigation ignore Caps Lock and Num Lock state, matching native macOS key events (which report Num Lock). Command/Control/Option-modified activation remains separate.
