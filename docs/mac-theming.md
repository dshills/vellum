# Scoped application themes

`ui.Theme` remains the application-local style contract. `Driver.Theme` supplies
startup defaults. A `ui.Node.Theme` pointer replaces the inherited theme for that
node and its descendants; nil inherits. Returning a different value from an
owner View changes the appearance while preserving compatible keyed controls,
values and focus. Removing a scope restores its parent's current theme.

`ui.MacThemeWithAccent(dark, accent)` supplies light/dark Mac presets and opaque
semantic palette colors for window/panel/field/popup, foreground/secondary/border,
and accent/accent foreground. Zero accent chooses blue. `ui.MacTheme(dark)` stays
available. Custom palettes and kind style overrides can be made from `Theme.Clone`.
Explicit node styles continue to take precedence. Drivers snapshot scoped maps
when accepting a View; callers must not mutate theme maps concurrently with View.
Invalid nested themes reject the whole keyed update.

The Mac driver uses palette accent colors for managed focus cues and menu rows,
and palette popup color for transient menus. Full-color image pixels are never
tinted. Changing a scope retires open menus before publishing new presentation.
Themes do not change process-global Cogent colors or macOS settings. Native
AppKit titlebar appearance and its saved configuration remain separate.

```go
appearance := ui.MacThemeWithAccent(true, color.NRGBA{R: 126, G: 65, B: 194, A: 255})
return ui.Node{ID: "inspector", Kind: ui.Column, Theme: &appearance,
    Children: []ui.Node{/* ordinary keyed controls */}}
```

The gallery's Components page demonstrates live Light/Dark and Blue/Purple/Green
choices. These choices are local to that page and are not persisted yet. Apply
the same scope at an application root to theme its whole managed content. The
preset does not automatically follow system accent, appearance, increased
contrast or reduced-motion settings.
