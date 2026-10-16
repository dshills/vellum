# Windows and platform services

Vellum supports multiple desktop windows within one native process main loop.
The portable `app` package remains independent of native dependencies.

## Window lifecycle

`Application.Open(Window)` registers windows before or during `Run`. Every ID
is unique across the application lifetime, including closed windows. The view
runs under the shared application UI owner; two windows never evaluate model
state concurrently. Native widget/render resources belong to their window. Stateful native extension
instances (including workbench, virtual views and GPU canvas adapters) must be
distinct for each mounted tree/window; applications may share their domain models
under the application owner. Reusing one adapter instance across mounts is outside
the extension contract.

`WindowContext(id)` returns a context derived from the application lifetime.
`CloseWindow(id)` removes its registration, cancels that context, drops pending
`PostWindow(id, callback)` delivery, and asks the driver to close its native
window. Other window contexts and accepted application work remain active.
Closing the final window or calling `Close` ends the app and cancels all workers.
User/native requests first evaluate Window.CloseRequest on the owner; refusal
preserves registration and native widgets. RequestCloseWindow invokes that policy;
CloseWindow/Close force teardown for confirmed decisions or cancellation. Native
CheckCloseWindow permission is separate from body cleanup. See [document and close
contract](document.md). An accepted native close follows the same lifecycle. Callers handle invalid IDs and
`ErrClosed`; a closed ID cannot be reused for a different document.

After closing a window, the application replaces its stored context with one
already-canceled marker and removes its stored cancel function. Captured original
contexts and their descendants are canceled first and retain their normal Context
semantics; later lookups of closed IDs need not preserve context object identity.
Used ID strings and map entries remain as tombstones to prevent reuse, so this
compaction reduces per-window retained cancellation state without establishing a
global bound on historical IDs or caller-retained contexts.

The driver uses an event source and coalesced wake flag for each window. A wake
notifies all live window sources; each window refreshes its own tree while owning
its native event/render context. A closing native window creates any accepted remaining
registrations synchronously before its event loop can end; this preserves topology
handoff even when no shown event source remains to receive a queued wake.
Native close requests run without holding the
application UI owner or another window's render lock. All windows share one
`core.Wait` / native main loop; a second run in the process is rejected.

Window titles must contain valid UTF-8 and no embedded NUL characters.

Owned native patch 0021 permits shutdown before main-loop queue publication. Stop
is idempotent and closes a stable signal without publishing a work queue. Startup
workers and unaccepted delivery waiters are released. An accepted callback retains
synchronous completion before its caller returns; loops acknowledge canceled
deliveries as well as executed work. Native window construction is one admitted
main-thread callback and returns an explicit error on refusal. The legacy
RunOnMain API returns no refusal error; worker service transports use the stopped
signal and their own context/error protocol.
The portable stop method does not post a native wake after signaling shutdown.
Windows and Linux retain the desktop loop's finite 100 ms wait. The Mac loop
owns an event-driven ingress and stop wake, joined before native termination;
see [native dispatch](native-dispatch.md). The event barrier refuses posts after
native termination or before initialization.

## Services

`Application.Services()` is available after a native window is shown and returns
`ErrUnavailable` beforehand or `ErrClosed` after shutdown. The adapter installs
an implementation of `app.Services`; headless callers may inject a fake with
`SetServices`.

Call service methods from workers, outside `Do`, `Drain`, view evaluation, and
widget callbacks. Service contexts scope pending operations; window-specific
requests should use `WindowContext(id)` and deliver their result with
`PostWindow`. All requests also end when the app closes. Context cancellation
prevents queued native operations from starting and releases waiting callers.
An OS operation already executing may complete before cancellation is observed.

- `ReadClipboard(ctx)` returns plain text, or an empty string if no plain text is
  available. `WriteClipboard(ctx, text)` replaces clipboard contents with plain
  text. Invalid UTF-8 and embedded NUL characters are rejected by this adapter.
- `ChooseFile(ctx, options)` opens an OS single-file open or save panel. `SaveFile`
  chooses a destination and never writes the file. Optional extensions omit dots.
  All option strings must contain valid UTF-8 and no embedded NUL characters.
  User cancellation returns `ErrCanceled`; caller/app cancellation returns the
  context error and requests panel dismissal. Paths remain ordinary local paths.
- `SetMenu(ctx, groups)` installs process-wide native application menus whose
  entries reference IDs in the driver's command registry. Labels and enabled
  states are captured under the owner at installation; call again to update that
  presentation. Native group and command labels must contain valid UTF-8 and no
  embedded NUL characters. Installation atomically transfers callback ownership
  before native menu replacement; cancellation during construction preserves the
  prior menu, and cancellation after acceptance retains its installed callbacks.
  Activation posts to the owner and rechecks current availability
  and application lifetime before executing. Existing keyboard shortcuts continue
  through the driver's command routing. Standard application About/Hide/Quit menu
  entries are preserved.

The current native service implementation uses Cocoa `NSPasteboard`, `NSOpenPanel`,
`NSSavePanel`, and `NSMenu` on macOS with cgo. Cocoa requests use asynchronous main
queue dispatch, avoiding a sender blocked on a native loop during shutdown.
Other OS builds return `ErrUnavailable` for these services; Windows/Linux runtime
and native service support remain future acceptance work. This phase does not
certify cross-platform accessibility, text input, or device-loss recovery.

## Verification

Portable race tests cover opening during Run, independent cancellation, closed-ID
reuse rejection, suppressed late callbacks, surviving-window delivery, final-window
shutdown, and topology wakeups. Native adapter tests cover command availability
rechecks and service cancellation before OS dispatch. `cmd/multiwindow-smoke`
opens a second window after Run, closes the first, refreshes the second, installs
an application menu, reads clipboard text without modifying it, auto-cancels a
file panel, and checks a pending panel caller ends on final-window shutdown.
`cmd/window-handoff-smoke` registers a replacement in the old window's native
Close callback, before queued wakes can run, then verifies surviving dispatch.
Native menu tests deterministically cancel before and after callback-ownership
commit to verify the existing/installed menus retain functional callbacks.
Actual menu clicking, file selection/save path interaction, and clipboard writes
need additional interactive acceptance; the smoke does not claim those results.
