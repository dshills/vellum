# Mac widget essentials

These public components preserve full-color pixels and use scoped Mac themes.
The Components demonstration selector includes each example as it is added.

## Image well

Retain `ui.ImageWell` and accept `ImageWellEvent` requests. The component displays
immutable `Image` pixels, a placeholder, Choose/Replace/Clear and controlled
`Status`/`Busy` feedback. File drops accept one absolute PNG/JPEG path; they never
read files on owner. Decode on a worker with `ui.DecodeImage`, publish through a
window completion and assign Image. The demo uses the native chooser and reads
only the selected file; its sample button needs no disk or OS dialog.

The backend acknowledges external drags before admission. Model rejection affects
framework propagation, not Finder's cursor. Result/loading feedback is supported;
pre-drop external hover feedback is not yet available. PNG/JPEG decode bounds
remain 32MiB encoded and 8192px per dimension. No file is moved, modified or deleted.

## Searchable icon picker

Retain `ui.IconPicker` with <=256 stable `IconPickerItem` values and <=32
categories. Set Image for original full-color pixels or Icon for a symbol. A
same-window popover offers search, categories, a responsive grid and up to eight
recent accepted choices. Arrow keys highlight; Return/double-click or Choose
requests selection. Accept by setting SelectedID in Change. Rejection leaves the
popover open and does not record a recent choice; Cancel/Escape changes nothing.
Closed/replaced sessions and filtered/removed/disabled items revoke old actions.
No images are loaded by the picker. Catalogue search is bounded and local.

## Settings pane

`ui.SettingsPane` supplies a full-color sidebar and grouped detail pages. Up to
16 pages remain mounted in a `TabsHideHeader` stack so native editors survive
navigation. Accept Navigate by setting SelectedID. Disabled pages and stale
navigation options cannot act. Inactive content and the hidden tab header leave
focus/accessibility scope. Use `SettingsNavigationID` for initial focus.
Embed the pane or put it in an ordinary application window. Persistence remains
an application policy; the demo stores preferences for the current launch only.

## Native alert service

Get `Application.Alerts()` and call `ShowAlert(ctx, parentWindowRef, options)` on
a worker. The optional macOS service presents an actual NSAlert sheet on that
exact opening, with 1–3 stable-ID buttons, default/cancel/destructive roles,
optional suppression checkbox and an original-color PNG icon. A missing icon uses
the native app icon. Other platforms return ErrUnavailable. Selecting Cancel
returns its button ID; cancellation without a button returns ErrCanceled.

A hidden/minimized parent returns ErrUnavailable; the service does not activate
a window or switch Spaces. A retired native parent returns ErrWindowClosed.
One native alert is admitted at a time; a parent with an existing sheet refuses
with ErrServiceBusy. Context, exact parent lifetime and app shutdown revoke the
request. Native cleanup closes owned sheets and never retargets a reused window
name. PNG input is <=256KiB, <=512px per dimension. Deliver worker results through
a window completion. Suppression persistence and destructive work belong to the
application; the demo changes counters only. Offscreen tests cover option/reply
validation, parent/operation lifetime and the Foundation recipe parser; actual
AppKit presentation remains unverified until desktop checks are authorized.

## Calendar and date picker

`ui.Calendar` provides an embeddable controlled month grid. Navigate accepts a
Month proposal; Change accepts a date. Value's civil time and location are kept
using the existing DateTimePicker range/DST engine. Monday/Sunday/etc first
weekday, a pure AllowDate predicate, original-color header image and a caller-
provided Today are supported. English labels are currently used. Left/Right move
between enabled cells; Up/Down retain the weekday column; Return selects. Bounds
are inclusive and years are limited to 1..9999.

`ui.DateTimePicker` now uses this same month grid in its popover, retains its
published focus IDs, and exposes Today/FirstWeekday/AllowDate/CalendarIconImage.
Typed dates and Today use the same bounds/predicate. Closed or replaced popover
sessions revoke old day actions. The demo disables weekends and shows both the
month view and compact field. Locale labels, range selection and event agendas
are outside this version.
