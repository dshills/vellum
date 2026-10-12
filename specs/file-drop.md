# External file drops

`ui.Node.FileDrop func(ui.FileDropEvent) bool` opts any managed node into external
file-reference drops. The callback executes synchronously on the application
owner using the current retained callback. It must not nest `Application.Do` or
`Drain`, block on I/O, or invoke native operations synchronously. Schedule file
processing through window-scoped tasks and commit results through their owner.

The payload contains the managed target ID, registered window ID, private path
slice and target-local `ui.Point`. Coordinates use Dp relative to the unclipped
outer layout origin, including margin and padding. Native scene offsets and
scrolling are applied before conversion. The clipped rendered bounds determine
eligibility, not the coordinate origin.

A true result handles the event once and invalidates its window. False leaves it
available for the nearest eligible ancestor, whose paths are independently
copied; declining callbacks should leave model state unchanged. Disabled nodes or
ancestors, inactive tabs/workbench panels, dismissed or closed dialog hosts,
removed widgets and closed application/window lifetimes cannot accept drops.
The visible modal host receives a Dialog callback; native modal routing blocks
background scenes even when the pointer lies outside the dialog.

The Mac driver receives native GLFW file URL drops as absolute text/plain path
records. Internal framework drag data, arbitrary MIME data, relative names and
URI-looking text are refused. `ui.CopyFileDropPaths` validates the whole batch
before returning a copy: 1..256 absolute host paths, at most 1 MiB combined UTF-8
bytes, no NUL. It preserves spelling, order, duplicates, Unicode, spaces and
newlines. It does not normalize paths, check existence, read contents or partially
admit invalid batches. These limits bound delivered/copy payloads, not prior
allocations made by GLFW or the operating system.

Acceptance transfers references only. Vellum never moves or deletes the source.
The pinned GLFW Cocoa implementation returns a generic drag operation and
acknowledges `performDragOperation` before model admission. A false callback
controls propagation within Vellum; it cannot report refusal to Finder. Drag
hover feedback, source sessions, arbitrary formats and native acceptance feedback
are outside this initial API.

The workbench demonstrates dropping files on Board A and replacing a bounded
file-reference list as a document history change. Undo/redo and dirty revisions
include that list. It does not import or persist file contents.

Acceptance needs portable payload tests, native routing/lifetime tests, actual
Cocoa file-URL receive-path tests, and a separate physical Finder drag. A selector
fixture alone does not establish the Finder workflow or OS feedback behavior.

The Cocoa receive fixture also requires positive restored background frames
after Escape and controlled dismissal of a label-only modal, without a resize,
pointer move or style change. A subsequent drop uses that restored frame. Its
timeout reports the last reachable accessibility inspection; native shutdown
cancels and joins the fixture worker so the diagnostic cannot be lost to a
result/deadline race. A five-second grace period bounds that join and reports an
explicit failure if the worker does not finish.
