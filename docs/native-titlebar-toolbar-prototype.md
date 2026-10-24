# Native macOS titlebar toolbar prototype

This is an opt-in driver experiment using AppKit `NSToolbar` and
`NSWindowToolbarStyleUnifiedCompact`. macOS owns traffic-light spacing, titlebar
hit testing, button sizing, overflow, hover feedback and native accessibility.
It keeps GLFW's original content view, window delegate, first responder and GPU
layer. Standard content stays below the native chrome; no guessed header padding
or rendering/pacing option is enabled. The application root fills the client
surface without the backend Scene’s default 8-point light margin.

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
  for the window opening. Native order, visibility and display choices are
  customizable separately and remembered by AppKit.

From the prepared driver module:

```sh
go run ./cmd/gallery -mac-widgets -native-titlebar
# dark variant:
go run ./cmd/gallery -mac-widgets -mac-dark -native-titlebar
# retain the earlier content-only gallery:
go run ./cmd/gallery -mac-widgets -native-titlebar=false
```

## Native customization and saved settings

Right/Control-click the native titlebar toolbar and choose **Customize Toolbar…**,
or click **Customize Titlebar…** in the gallery. This opens AppKit’s standard
palette. Drag actions to reorder, remove them by dragging out, or drag missing
actions back from the palette. Fixed and flexible spaces are available. Drag the
default set onto the toolbar to restore its initial items. Native display options
(Icons, Text, or Icons and Text) come from AppKit, with their normal OS appearance.
The recovery button also shows a toolbar hidden through AppKit’s menu.

AppKit autosaves the configuration under the toolbar’s stable identifier in the
application’s user defaults. Light and dark demos use separate identifiers ending
in `.light.v1` and `.dark.v1`. Order, omitted items, spaces, display mode and native
visibility survive app relaunch. No content-toolbar JSON files or system settings
are changed. New catalogue items are available in the palette; existing saved
layouts are respected rather than reset on each launch. Unknown retired actions
cannot be reconstructed by the delegate. Interactive layouts are capped at 32
slots, including repeated spaces; oversized restored layouts are trimmed to that
bound. At the cap, remove a slot before adding another space.

The original full-color, non-template images are reused in the toolbar, palette
and overflow. A fresh native item is created for each delegate request; moving or
re-adding an action retains its stable command ID. Native inactive/disabled
appearance is controlled by macOS. Customization itself never invokes commands.
Opening the native palette and toolbar membership changes revoke pending actions;
managed dialog scope still blocks native commands and palette requests. Native
palette previews remain draggable even for disabled commands.

Customization requires macOS 15+ because that is where AppKit documents palette
visibility observation. The original unsaved prototype still supports macOS 11+.

## Driver experiment

Set `cogent.Driver.TitlebarToolbarPrototype` before `Application.Run`. Supply the
original `app.WindowRef`, an optional stable `CustomizationID`, and 1–16 `cogent.TitlebarToolbarItem` entries. Buttons
need unique IDs, labels, bounded PNG bytes and an Activate callback; fixed and
flexible spaces need only an ID and kind. PNGs are checked fully before native
startup (at most 64 KiB and 128×128 pixels), and copied into an immutable native
payload. Labels/IDs are bounded valid UTF-8 without control characters. Action
closures remain application-owned; do not mutate the catalogue during Run.
A nonempty `CustomizationID` (at most 128 ASCII letters/digits/dot/underscore/
hyphen) enables native customization and autosave. Give independently saved bars
different stable IDs; do not encode the ephemeral WindowRef in that ID. An empty
ID retains the prior fixed, unsaved behavior. From an owner callback,
`cogent.CustomizeTitlebarToolbar(ref)` asynchronously opens the native palette and
recovers hidden chrome, with scope/lifetime checks again on native main.

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
standard-chrome window opening (macOS 15+ when customization is enabled). Extra
or reopened windows and `ChromeUnified` are refused. The action catalogue is
immutable during that opening; its native presentation is user-customizable. It
does not bridge content-toolbar configuration files or offer search/group/split
controls, titlebar accessory views or cross-window dragging.
Native keyboard/VoiceOver behavior comes from AppKit; focused tests and recorded
native interaction checks are not a claim of human VoiceOver qualification.

The content toolbar implementation and persisted customization remain available.
No changes to renderer, frame scheduling or release qualification are required.

Apple references: [NSToolbar](https://developer.apple.com/documentation/appkit/nstoolbar)
and [NSToolbarDelegate](https://developer.apple.com/documentation/appkit/nstoolbardelegate).
