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

Current status: the complete 57-input/21-observation native record passes with
the inherited-state and actual multi-click focus fixes. Six preceding attempts
remain failed/sealed. This is automated native Go routing in the controlled
Go2/SDK786a CPU-scene configuration; physical/OS input and full Mac acceptance
remain open. See the [result](reports/macos-control-refocus/README.md).

The gallery's `-choice-native-check` is a separate bounded mode for shared
choices, toggles and Slider keyboard input: 30 routed inputs/23 observations pass
with callback/rejection/disabled/focus/closure checks. Progress/status publication
uses numeric semantics and measured layout, not native pixel validation. It
uses the same public API and current exact group-qualified option identities;
repeated option DomainIDs across groups are not input capabilities. See the
[native choice result](reports/macos-choice-catalogue/README.md). Physical/OS input,
Slider drag, presentation and full Mac gates remain open.


Use `gallery -navigation-native-check` for the separate native ListBox/MenuButton
record: 47 packets/41 observations. It uses actual group-qualified row bounds and
actual current popup item IDs; repeated action labels are distinguished by item
order/identity and verified callback IDs. Reopening never revives an earlier
popup target, even with a new frame. Read-only list keys can change focus and
scroll while accepted selection stays fixed. See the
[native list/menu result](reports/macos-navigation-catalogue/README.md). It does
not establish native pixels or physical/OS/assistive-technology input.


`gallery -structure-native-check` exercises tabs/disclosure/split/scroll and
controlled document closes through public DesktopSession Go FIFO, with an authored 50-input/42-state record. Its first native run fails at41/35
before document selection/close observations; full native qualification is pending.
The runnable consumer checks actual targets/focus, divider geometry, scrolling,
callbacks, stale refusal and terminal cleanup without substituting semantic actions. The [result](reports/macos-structure-catalogue/README.md)
keeps this evidence separate from physical input, native pixels and broader Mac
qualification. Ordinary tab keyboard navigation is documented in the
[controls contract](../specs/controls.md).

The structural native prefix above predates the subsequent Prism tab activation
and disabled-focus fixes. Those fixes pass actual Core queue/manager routing and
complete driver-root/gallery race suites. At that prior checkpoint they had not
received a new native run. The failed diagnostic attempt below now exercises
those fixes but does not complete structural qualification. Both failed scopes
remain sealed and the full structural/native gate stays open.


### Measured acquisition failures

With the prepared stage85 backend, a measured render-pass failure includes a
detached acquisition description: source, stage, actual attempt count and first/
last backend statuses. A surface unavailable due to occlusion may yield an error
such as `gpu: measured render pass unavailable (source=surface stage=unavailable
attempts=1 first=occluded(196609) last=occluded(196609))`. A successful acquisition
does not identify the cause of any later encoder/pass failure. Backend errors
retain their existing error identity and text with additional acquisition context.
Legacy measured implementations may report unrecorded detail; offscreen render
textures have no surface status/attempts. No extra acquire, poll, present, wait or
readback is performed, and acquisition timing boundaries are unchanged. Metadata
work counts as UI work. Supported=true, Submitted=true and no failure are still
required for a completed Frame. Native occlusion-to-visible scheduling acceptance
is open; see the [diagnostic checkpoint](reports/macos-acquisition-diagnostics/README.md).


### Typed unavailable results

`errors.Is(err, cogent.ErrDesktopFrameUnavailable)` identifies checked known
surface unavailability from `DesktopSession.Frame`. `errors.As` can extract a
`*cogent.DesktopCompositionError` with detached `Failure` text and `Unavailable`
metadata. The existing error string is preserved; no string parsing is used for
classification. Generic/legacy failures do not acquire this identity. Context
cancellation and closed-window/session errors retain their standard identities.

An unavailable or failed composition terminal publishes no frame capability and
releases capacity after its actual input/owner/composition terminal. Cancellation
does not retract dispatched or committed work, and its return may race an already-
committed reply. Publication is blocked when cancellation is observed before
terminal collection; a late canceled wait does not erase a committed capability.
Frame does not retry or wait automatically.
An application can explicitly request another Frame following a meaningful
lifecycle/application event; shown/focused state and this error do not prove
drawable readiness. The new scalar is also available on `FrameSample`, omitted
from JSON whenfalse. Contradictory submission flags and source-diagnostic faults
stay failed generic terminals. Timing/ownership/first-composition contracts remain
unchanged. See the [checkpoint and verification limits](reports/macos-frame-unavailable/README.md).


The Mac adapter now requests an owner redraw when its sampled native occlusion
state changes from hidden to visible. Applications using on-demand rendering need
no visibility polling. This redraw also works while a framework popup is open;
it does not change application selection or dismiss the popup.

A session consumer can inspect `session.Visibility(ctx, originalRef)` even after
a failed Frame. The returned `DesktopWindowVisibility` copies `Known`, `Visible`
and an attachment-local `Sequence`. Unsupported/unobserved backends return
`Known=false`; no completed frame is required or granted. Use the same original
WindowRef, deadline and one-consumer rule as other session operations. This is a
Cocoa state observation, not a drawable-readiness fence. Frame may still yield
`ErrDesktopFrameUnavailable`; the framework does not retry failed session frames.
The [visibility checkpoint](reports/macos-window-visibility/README.md) separates
software/notification ownership checks from actual native-window qualification.


If the visibility sequence reaches its uint64 limit, the next change retires
this attachment's automatic visibility redraw state and removes any queued wake.
`Visibility` then returns `ErrDesktopBudget`; the counter never wraps or revives
old authority. Ordinary application invalidation and painting remain available.
A fresh window attachment has its own sequence. This failure mode is covered by
a queued-wake retirement regression.
