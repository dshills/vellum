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
