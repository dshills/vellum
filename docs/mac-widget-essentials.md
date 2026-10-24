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
