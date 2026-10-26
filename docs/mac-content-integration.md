# Using the Mac content widgets

Keep each component on the application's UI owner. A `View()` describes accepted
state and returns an error; handle that before composing your window. Native
PDF/media paths are local sources, not documents/history owned by the framework.
Use the prepared checkout or extracted private SDK; the public module namespace
is github.com/dshills/vellum. Do not import the gallery, Core, GLFW or driver
internals into application models.

## Controlled state and history

The runnable [checklist history example](../examples/sdk-contract/content_widgets_test.go)
proposes a completion through its exported focus ID, records one before/after
change in an application journal, and proves Undo/Redo retain the same item ID.
The date-range example accepts a validated inclusive civil range. Both compile
in a separate consumer module with no native driver dependency:

```sh
(cd examples/sdk-contract && go test -race ./... && go vet ./...)
```

In an application use `command.Registry` for Undo/Redo and `command.AfterFocusShortcuts`
so text editing has first refusal. Commit field drafts before structural commands;
[Reminders](../examples/macapps/reminders/README.md) is the full runnable example
of native field settlement, bounded history, checkbox proposals and range filtering.
Range filtering is view state there; it deliberately creates no domain history.

## Asynchronous path selection

The [complete Session consumer](../examples/sdk-contract/contentexample/session.go)
retains FilePickerField, PDFViewer and MediaPlayer using only public Vellum APIs.
Construct it after `OpenWindow` using that exact `WindowRef`; call `Session.View`
in the window view. Native services become available after the driver shows a
window. FilePicker.Request runs on owner and reserves one window completion;
a worker obtains `Application.Paths()`, calls the native chooser, validates
metadata and publishes a result. Only the delivered owner callback changes the
accepted path or viewer. Cancel keeps the old path. Overlapping requests refuse
and configuration changes invalidate a late result. No file contents are written.

The session's fake chooser tests verify owner-only delivery, exact opening
identity, overlap refusal and close cancellation without showing an OS chooser.
They do not qualify native panel input; native evidence is recorded separately.

## Theme, layout and cleanup

Use `ui.MacTheme`/the supported Mac system appearance adapter on the native driver.
Propagate the chosen Dark value to PDFViewer and MediaPlayer. Preserve immutable
IconImage/PosterImage objects for original-color artwork; don't mutate their source
or share mutable component instances across owners. Native PDF Find keeps a
usable width and wraps below480 native points. Allow at least280 points of content
width and enough vertical room for the native controls.

When retiring a Session, call Close on owner after native dispatch stops. It
cancels/aborts its operation and drops accepted native sources. If retiring it while
a window remains live, remove its view and invalidate that exact live window on
owner; do not reuse its WindowRef for a later opening of the same ID. The native
driver removes embedded views and pauses media on node/window retirement. Never
block owner callbacks on dialogs, file I/O or worker completion.

PDF/media native overlays are absent from offscreen captures. Arbitrary overlapping
managed popovers remain unsupported; modal projection removes the background host.
Native decoder/file admission and mutable media-path limits are in
[control contracts](controls.md). This work does not alter the rendering backend,
release performance gate, accepted private-v1 archives or human qualification.
