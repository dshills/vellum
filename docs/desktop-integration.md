# Native integration sessions

`cogent.NewDesktopSession()` provides optional integration routing for an actual
native driver. Install it with `cogent.Driver{DesktopSession: session}` and keep
one consumer goroutine. The driver starts and retires the session; it runs once.
Ordinary applications can leave it nil. It is mutually exclusive with FrameMetrics.

After the driver has started and the original window has a live native binding:

1. Call `session.Frame(ctx, window.Ref())` to wait for an actual completed
   composition. The native scheduler remains autonomous.
2. Call `session.Snapshot(ctx, ref)` for detached control/focus metadata, scene
   size/DPI and completed grid/tree/canvas geometry.
3. Call `session.Input(ctx, ref, cogent.DesktopInput{Frame: snapshot.Frame, ...})`
   with initialized Go key, pointer or scroll values. The input and its owner
   acknowledgement enter that window's actual native FIFO. It invokes the usual
   routing, hit testing, capture, keyboard, command and history paths.
4. Obtain a fresh completed frame before the next input or snapshot. Input can
   change focus, caret or paint even when no application model changes.

Coordinates and deltas use scene pixels. `Target`, when specified, must be the
exact current `DesktopElement.ID`, with enabled/completed bounds (and current
focus for a key target). `DomainID` is descriptive, not a replacement capability.
Use the original `app.WindowRef`; closing and reopening the same window name
cannot revive it. A frame token from one window cannot admit another window.

Never call the blocking methods while holding the application owner or from a
native callback. Startup can report pending/not-yet-bound; stop on fixture
closure or the caller's deadline. Concurrent requests return ErrDesktopBusy.
An inspection that needs layout/composition returns ErrDesktopPending. Budget
exhaustion refuses the inspection rather than returning partial success.

One request and at most 64 live windows are retained. Inspection limits include
8,192 semantic nodes, 8,192 copied composite members, 1 MiB per value and 8 MiB
of cumulative retained semantic strings/values. Desktop canvas inspection keeps
actual painted objects/selection/placement preview separate from the existing
CanvasInspection annotation count. Returned slices and JSON values are detached.

Cancellation removes undispatched request-owned FIFO entries. Input already
dequeued can have effects and cannot be retracted: retire the fixture on failure.
A canceled frame keeps its single pending ticket until actual input/owner/
composition terminal conditions or real window retirement. `State()` reports
bounded accounting; driver teardown clears hooks, traces, window roots and owner.

This API generates Go events, not OS input. It does not drive an IME, VoiceOver,
clipboard, native event scheduler, cross-window drag transport or physical pointer.
It is not a deterministic `replay.Backend`; software Manual remains explicitly
unsupported on desktop. A completed composition does not establish display
presentation or timing targets. Commands that could change OS clipboard/settings
or invoke global fallback shortcuts are refused by this transport.

The [workspace-native checker](../examples/replay/cmd/workspace-native/README.md)
consumes the fixed professional workspace record with native geometry anchors
and a narrow Mac Select All intent translation. It retains source-specific
receipts; consult its report for the current outcome and remaining gates.

For actual menus, inspect `snapshot.Popups`, then use the current item's exact
`ID` and center of its clipped `Bounds`. Persistent option semantics in
`Elements` describe model choices and are separate from the open popup's actual
item geometry. Only enabled items of the top interactive popup are targets.
Reacquire a frame/snapshot after each input, including menu navigation. Closed or
replaced popup/button identities cannot be reused, including in the same frame.
Custom/nested menus retain their existing inspectors.

The standard gallery has a runnable bounded integration mode:

```sh
cd drivers/cogent
go run ./cmd/gallery -control-native-check
```

It uses the ordinary public Selector/NumberField constructors and emits one JSON
result after its 57-input/21-observation scenario and tracked shutdown. See the
[gallery instructions](../drivers/cogent/cmd/gallery/README.md) and
[current result](reports/macos-native-controls/README.md). This command routes
Go events through the native FIFO; physical/OS input remains separate evidence.

Current status: all four original native attempts remain failed. The last reached
24 routed inputs/11 observations before a missing-refocus guard. The corrected
scenario now adds four explicit routed refocus clicks after numeric Enter:
57 inputs/21 unchanged model observations. Its software focus/refocus regression,
gallery race checks and build pass; the revised native record has not run. Full
numeric/disabled continuation and final original-reference closure remain pending.
