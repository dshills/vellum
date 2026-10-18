# Native Go integration routing

The optional `cogent.DesktopSession` is a single-consumer, run-once capability
installed on Driver. It is separate from the offscreen deterministic Manual and
replay backend. The native driver retains autonomous scheduling and wall time.

## Input and completion

Requests use the original `app.WindowRef`, a current completed-frame token, and
an optional exact qualified semantic element ID. Reopening a window name cannot
revive old authority. Inputs are initialized through a session-owned event Source
and enter the originating native FIFO without merging, followed by an owner
acknowledgement. OS Source.Last and global keymaps are not modified. Every input
invalidates its completed capability, including focus/caret-only changes.

Frame completion requires actual Core FrameTrace completion. Unsupported timing
or failed/canceled composition never publishes a successful token. Origin Custom
requests reach their original owner before popup routing; unrelated events retain
normal modal routing. Frame metrics and sessions cannot share the trace hook.

## Inspection and lifetime

Snapshots contain detached semantic values and actual focus, bounds/DPI, completed
grid/tree geometry, and painted canvas objects/selection/placement preview.
Workbench tab focus reads the actual retained header; selection alone is never
focus. Current semantic/modal scope controls which elements are published.
Pending layout or composition refuses a snapshot instead of exposing stale state.

At most one pending request and 64 live windows are retained, with limits of
8,192 semantic nodes and 8,192 composite members, 1 MiB per value and 8 MiB total
semantic strings/values. Concurrent requests refuse. Canceling input removes
undispatched request-owned FIFO entries; already dequeued effects cannot be
retracted. Canceled frames retain their bounded ticket until a real terminal
condition. Window/driver retirement clears trace hooks, roots, maps and owner.
Blocking calls must run outside the application owner and native callbacks.

## Acceptance boundary

The dedicated workspace checker uses all 216 immutable operations and 27 domain
captures at both stored row counts. Explicit targets and authored composite
anchors resolve against actual completed native geometry. Only authored Select
All intents 17/122 translate to the actual Mac Command+A binding. Grid vertical
scroll uses an independent fixed-row/source-anchor/wheel/clamp reference that
first reproduces the software goldens; candidate offsets never define expected
values. Other domain assertions remain exact, with the documented one-logical-
pixel camera-center allowance and independently validated viewport dimensions.

Regression coverage distinguishes synthetic oracle controls, actual software
FIFO routing and native composition evidence. Native Go integration does not
qualify physical pointer input, IME, VoiceOver, clipboard/services, cross-window
palette drag, frame presentation, performance, full catalogue or resource
endurance. Tracked cleanup and process exit do not establish Cocoa deallocation.

See [usage](../docs/desktop-integration.md) and the
[Mac delivery plan](macos-delivery.md) for actual qualification status.
