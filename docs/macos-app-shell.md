# macOS application shell

`Driver.TitlebarToolbars` is an owner-side factory called once for each window
opening, including later/reopened windows. Return `TitlebarToolbar{Window:ref}`
for that exact opening, or nil to omit its toolbar. Standard chrome is required.
Up to32 concurrent native attachments are supported. The legacy singular fields
retain their one-window restriction; they cannot be combined with the factory.

A toolbar catalogue is immutable for its opening: stable item IDs, PNG artwork,
labels and optional customization ID. AppKit owns overflow, original-color icons,
customization and autosave. Use distinct customization namespaces when layouts
should be independent. Reusing a namespace intentionally shares layout only;
callbacks and command state remain bound to the exact window opening.

Each button may supply `Activate`, or `Command` with `Commands`, exclusively.
The factory inherits `Driver.Commands` when its spec's registry is nil. A
window-specific registry also supplies that window's managed keyboard routing.
The optional `Available` callback and command availability run only on the UI
owner during refresh and immediately before execution. Invalidate the window
after model state changes. Native main reads detached permissions, never models.
Availability/modal/customization transitions revoke queued capabilities, including
an enabled→disabled→enabled round trip. Closing a window revokes its actions;
closing/reopening the same ID never revives its former toolbar.

Native attachment is asynchronous after the exact window is shown. It resolves
a borrowed native window by its live accessibility token on Cocoa main. A failed
attachment is surfaced on the owner; acknowledgment uses coalesced invalidation
rather than a fallible owner action slot. One latest availability snapshot and
one pending native update are retained per toolbar; teardown clears references.

`Driver.CommandMenus` copies a bounded application menu catalogue from the
shared `Driver.Commands` registry. Labels/enabled state refresh after invalidation.
A single owned worker sends detached snapshots to the native service, coalescing
intermediate updates and stopping at shutdown. A failed install is reported on
the owner and permits one later same-state retry; persistent failure cannot
create a self-invalidating retry loop. Native activation rechecks the
live registry. Mac application menus are process-wide, not a separate menu per
window; apps with several document registries must provide their own active-
document command router for global menu actions. Single-character window-first
shortcuts receive native key equivalents. Focus-first shortcuts keep managed
routing so a native menu cannot steal text-widget input.

Notes demonstrates a full-color native toolbar, shared command menus and the
existing text settlement/history/file paths. `go run ./cmd/macshell` from
`examples/macapps` opens two independent command/toolbar windows. These consumers
use public Vellum APIs and original Vellum sample artwork. No AppKit icons are
copied. Automated isolation tests do not substitute for physical interaction or
human IME/VoiceOver acceptance.

Native Open/Save panel completion returns keyboard focus to its exact visible,
nonminimized parent and live responder only while the app is active and no other
window has taken key status. A nil key window is accepted only while that exact
parent remains the application main window. It does not activate background apps or reopen
retired windows. The physical Notes cancellation check covers the ordinary path. Scalar identity
and state fixtures cover refusal policy; they do not simulate closing an actual
NSWindow or native responder delivery.

## Active document menus

`Driver.OnWindowActivated` reports native macOS focus gained on the application
owner with the exact live WindowRef. Its callback must not enter Application.Do.
Use a single process-menu registry that forwards availability and execution to
the active document registry; recheck the reference on every call. Keep separate
per-window TitlebarToolbar.Commands registries. A retired active reference refuses
commands until another live document activates. The callback follows the existing
coalesced native focus observer and adds no polling or global preference changes.

The independent consumer is `(cd examples/macapps && go run ./cmd/documents
-data-dir /absolute/private/workspace)`. A and B use separate repositories,
models, histories, recovery files and native toolbar namespaces. Close one while
retaining the other; OS focus changes route File/Edit menus to the surviving
document. Source tests cover routing and retirement; native observations remain
separate from human accessibility and other-platform qualification.


## Coordinated application Quit

`Driver.OnQuitRequested` is an opt-in owner callback for native application Quit.
It must not nest Do or block for native work. Repeated foreign requests coalesce
while a callback awaits the owner. The handler is installed before native windows
appear and removed after Run. Forced shutdown and no-window termination keep the
existing terminal path. Nil retains per-window close semantics.

The app-owned `notes.QuitGroup` demonstrates sequential input settlement and
Save/Discard/Cancel decisions while holding all participating document commands.
It checks live references and approved journal revisions, clears recovery, then
closes the complete approved set. Cancel, save/clear errors, stale state and
retired references keep surviving windows open. Saving before a later Cancel
retains that successful save. Ordinary red-button window close is independent.
Notes, the generated document starter and the A/B Documents consumer opt in.
Forge and Measure Pad each have one window and retain their existing close path.
