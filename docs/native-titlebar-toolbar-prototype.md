# Native macOS titlebar toolbar prototype

This is an opt-in driver experiment using AppKit `NSToolbar` and
`NSWindowToolbarStyleUnifiedCompact`. macOS owns traffic-light spacing, titlebar
hit testing, button sizing, overflow, hover feedback and native accessibility.
It keeps GLFW's original content view, window delegate, first responder and GPU
layer. Standard content stays below the native chrome; no guessed header padding
or rendering/pacing option is enabled.

## Try it

The packaged **Vellum Mac Widgets** and **Vellum Mac Widgets Dark** applications
open on Toolbars with the native prototype enabled. The titlebar offers New,
Open, Save, Undo, Redo and a deliberately disabled Delete, using the gallery's
original full-color artwork. These are sample commands: they update the action
status in the content area, without opening files or changing documents.

- Click an enabled titlebar action and check `Native titlebar: … · actions: …`.
- Move the window using an unoccupied titlebar area. Content controls retain
  their normal interactions.
- Make the window narrow to try AppKit's native overflow representation.
- Open a content toolbar's Customize panel. Native actions are disabled while
  the managed modal owns input; queued actions from an earlier modal scope are
  discarded. Close the panel to enable them again.
- Content toolbar size/color/display/customization controls affect the content
  examples. The prototype's native catalogue and full-color appearance are fixed
  for the window opening.

From the prepared driver module:

```sh
go run ./cmd/gallery -mac-widgets -native-titlebar
# dark variant:
go run ./cmd/gallery -mac-widgets -mac-dark -native-titlebar
# retain the earlier content-only gallery:
go run ./cmd/gallery -mac-widgets -native-titlebar=false
```

## Driver experiment

Set `cogent.Driver.TitlebarToolbarPrototype` before `Application.Run`. Supply the
original `app.WindowRef` and 1–16 `cogent.TitlebarToolbarItem` entries. Buttons
need unique IDs, labels, bounded PNG bytes and an Activate callback; fixed and
flexible spaces need only an ID and kind. PNGs are checked fully before native
startup (at most 64 KiB and 128×128 pixels), and copied into an immutable native
payload. Labels/IDs are bounded valid UTF-8 without control characters. Action
closures remain application-owned; do not mutate the configuration during Run.

Activate posts through `Application.PostWindow`, rechecking the original opening
and an immutable input-scope permission at drain. It never calls the owner
synchronously from AppKit main. Unknown/disabled/retired actions and actions
refused by the application's bounded queue do not execute. Accepted actions
invalidate the originating window. Commands can call the application's existing
command/history code inside their owner callback.

Native main owns the delegate, images, toolbar and a one-entry token registry.
Retirement revokes Go input, disables items, clears targets/delegate, detaches the
owned toolbar and removes its token. Queued native scope updates resolve tokens
at execution and cannot resurrect a removed attachment. No Go pointer enters
Objective-C storage.

## Deliberate limits

This is a driver-only feasibility prototype, not a new portable `ui.Toolbar`
contract. It requires macOS 11+ with cgo and exactly one registered
standard-chrome window opening. Extra or reopened windows and `ChromeUnified`
are refused. Native items are immutable during that opening. It does not bridge
content-toolbar customization or persistence, native customization palettes,
search/group/split controls, titlebar accessory views or cross-window dragging.
Native keyboard/VoiceOver behavior comes from AppKit; focused tests and recorded
native interaction checks are not a claim of human VoiceOver qualification.

The content toolbar implementation and persisted customization remain available.
No changes to renderer, frame scheduling or release qualification are required.

Apple references: [NSToolbar](https://developer.apple.com/documentation/appkit/nstoolbar)
and [NSToolbarDelegate](https://developer.apple.com/documentation/appkit/nstoolbardelegate).
