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
