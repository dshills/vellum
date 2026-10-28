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

Reusable workbench controls resolve selection, paused states and separators through
semantic palette roles. `Style.ForegroundRole` and `BackgroundRole` accept `ink`,
`secondary`, `line`, `accent`, `accent-ink`, `panel` or `field`. Roles use the
inherited scoped theme palette (light Mac defaults when absent). A later explicit
foreground/background color overrides its role. Original-color artwork is unchanged.
`Style.BorderRole` resolves an outline color while retaining its configured width.
When a single style supplies both a foreground/background role and a concrete
color, the concrete color wins; use Overlay to replace it with a role.
A BorderRole in the same style as Border replaces only its color. A later
Overlay containing a complete Border clears the previous BorderRole; supply
BorderRole again when changing width while retaining semantic color.

Model-browser selected rows apply accent ink to disclosure/visibility/activation
actions and badges as well as labels. Icon-to-label spacing is seven logical
pixels; original-color item artwork remains untinted. Native focus and disabled
feedback continue through the existing managed button styles.

Explicit toggle-symbol foreground colors and semantic roles reach the checked
and unchecked glyph leaves. This prevents a native primary-blue glyph from
vanishing on an accent-blue selection. Omitted overrides retain native defaults;
full-color image artwork and native focus/disabled behavior are unaffected.
