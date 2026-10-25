# Shared Mac presentation

Use `ui.MacTheme(dark)` for compact light/dark controls. It uses
`ui.MacFontFamily` (`Helvetica Neue, emoji`) and `ui.MacTypography(size, weight)`
so custom font queries include the system emoji family instead of LastResort.
No system font is redistributed. Original-color image icons are not tinted.

Style sizes are content-box constraints; padding and the reserved focus border
add to actual allocation. The shared single-line field minimum is16 Dp of
content with5 Dp vertical padding, yielding approximately28–32 Dp controls with
normal typography. Larger fonts may grow naturally. Explicit node styles win.

Use `ui.MacProseStyle(dark)` on `ui.MultilineText` with `Prose:true, SoftWrap:true`
for an unframed document surface. Caret/selection show focus. It changes no
document bytes, editor history, input settlement or global OS preferences.
Code editors retain their native frame/gutter by default. App-specific column
widths and selection accents still belong to the app. Notes consumes this shared
style; it does not implement a second text renderer or font fallback policy.
