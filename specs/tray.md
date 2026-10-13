# Mac menu-bar status item

`app.TrayServices` is an optional extension of `app.Services`; legacy service
implementations remain compatible. `Application.Tray()` returns the extension
or unavailable. Capability flags describe implemented transports, not runtime
acceptance. Only the Mac/cgo adapter implements it; other platforms return no
capabilities and unavailable. Call methods from workers outside UI-owner callbacks,
after native services become available. This is a companion to the existing
windowed application lifecycle: closing the final window still ends the app.

`SetTray(ctx, TrayOptions)` supplies the full title, tooltip, accessible name,
optional PNG icon/template flag and a flat command menu with explicit separators.
It creates one application-wide status item and reuses that native object on
subsequent updates. `ClearTray` removes only that owned item. Tray callbacks and
resources are independent of a particular surviving window.

Titles are at most 256 UTF-8 bytes; tooltip/accessibility strings at most 1024,
with no NUL or invalid UTF-8. An accessible name defaults to Tooltip, then Title;
an unnamed icon-only item is refused. The Cocoa button receives the resolved
accessibility label. At least a nonblank title or icon is needed. Human VoiceOver
and native menu keyboard acceptance remain separate from source/API availability.

PNG input is capped at 1 MiB and 256×256 pixels. `TrayOptions.Validate` checks
metadata and PNG header bounds; the adapter additionally decodes the complete
file before any native installation. A valid header cannot admit a corrupt body.
The worker normalizes pixels to 8-bit straight RGBA and sends bounded data to
Cocoa; Cocoa does not decode caller-controlled PNG. Images fit 18×18 Cocoa points
while preserving aspect ratio. Template icons ask Cocoa to supply appearance
tinting. Color-profile processing/16-bit fidelity are not promised. Input slices
must remain immutable until the method returns; native work owns detached data.

Menus have at most 128 entries and reference IDs of at most 128 UTF-8 bytes.
Labels and enabled state are captured on the application owner using direct
registry lookup. Native click callbacks only post owner work. Owner dispatch
rechecks installed tray token, shutdown and command availability before Invoke;
actions queued for a replaced/cleared item are refused. Labels/enabled state do
not refresh implicitly; applications call SetTray with their next snapshot.
Application-menu callbacks share the shutdown/current-token guard, so a queued
command from a retired application menu also cannot run on a still-live owner.

Menu/icon preparation precedes callback ownership transfer. Canceled replacement
before transfer preserves prior callbacks/presentation. A first native item may
be constructed and hidden during an already executing request. Cancellation or
native failure after transfer cannot promise rollback; a context error does not
prove that native presentation was unchanged. In-process PNG decoding/encoding
is bounded but not interrupted mid-step by caller cancellation.

## Shared native service admission and teardown

Mac chooser, application-menu and tray requests share eight process-wide queued
or running credits, separately from the four clipboard credits. Admission occurs
before request encoding/native copying. Canceled callers retain credit until
their native callback acknowledges completion/skip, preventing a stalled main
queue from retaining an unbounded request history. Full admission returns
`ErrServiceBusy`. Every skipped/error/commit-refused branch acknowledges its token.
Only chooser cancellation asks Cocoa to dismiss a panel; menu/tray cancellation
does not schedule extra panel operations.
Dismissal attempts cancel and order-out independently, fences native exceptions
and acknowledges the operation even if completion does not run. An exceptional
owned panel remains tracked for terminal cleanup; at most eight owned panels can
exist, including those retained failures. New chooser creation is refused with
Unavailable at that native ownership limit. Cleanup failures remain observable.

Native JSON is capped at 1 MiB, native string results at 64 KiB before copying
into Go. Chooser titles have 1024 bytes, directory/filename each 64 KiB, at most
64 extensions of 64 bytes. Application menus allow 16 groups/256 total commands,
1024-byte group/command labels and 128-byte command references. Oversize inputs
return `ErrServiceTooLarge`. Root/native text validation and original chooser/
menu behavior still apply. Snapshot count/label limits are enforced by the common
adapter; JSON/reply and eight-credit limits apply to the Mac native transport.

Cocoa request producers and consumers have explicit autorelease pools. Worker
storage is copied before delayed consumption. Native callback tokens do not wrap;
terminal completion retires operation records. Shutdown immediately closes Go
admission/retired actions and releases waiting workers, then requests native
cleanup. A separately registered desktop main-loop hook guarantees native
cleanup before GLFW terminates, even if the asynchronous cleanup request has not
run. The hook closes admission again, cancels owned panels, restores the original
application menu and removes the tray item. Cleanup errors remain observable
through Driver.Run, including earlier asynchronous cleanup failures; model and
cleanup errors are joined. Already queued canceled native payloads remain bounded
until their blocks acknowledge or the process exits.

Owned backend patch45 adds at most 16 main-thread cleanup hooks, with linearized
registration/removal, LIFO terminal ownership, no lock across callbacks and
complete cleanup/termination despite a callback panic. Hooks must not block,
enter the UI owner or redispatch native calls. Application-supplied callback time
and broader panic/GPU resource recovery remain outside this service contract.

Offscreen fixtures exercise Go option/PNG/command ownership, canceled admission,
callback retirement, actual Cocoa JSON/bitmap/alpha/aspect conversion and producer
copy/release. Backend tests inject cleanup/termination functions. They never
create a status item, open a menu/panel or modify clipboard/OS preferences.
An explicitly opted-in `vellum_native_checks` acceptance build inspects the
adapter's actual NSStatusItem/NSMenu objects and invokes installed AppKit menu
actions. It checks title/name/icon/template presentation properties, native item
identity across updates, replacement/clear callback refusal, owner availability,
surviving-window lifetime and main-thread cleanup before GLFW termination.
The probes and diagnostic API are absent from ordinary SDK builds. Retained
stale test targets/items are bounded and released by the terminal observer.
This selector fixture does not establish physical menu clicks, bar-position
pixels, VoiceOver, dark-mode tinting or mixed-DPI behavior.

Primary APIs: [NSStatusBar](https://developer.apple.com/documentation/appkit/nsstatusbar),
[NSStatusItem button](https://developer.apple.com/documentation/appkit/nsstatusitem/button?language=objc)
and [straight-alpha bitmap format](https://developer.apple.com/documentation/appkit/nsbitmapimagerep/format/alphanonpremultiplied?language=objc).
