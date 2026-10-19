# Layout and theme contract

Vellum describes layout and appearance with ordinary Go values in `ui`. The
portable module imports only the standard library. Cogent style types remain
inside the native driver.

## Properties and precedence

`Node.Style` contains optional padding, gap, alignment, justification, minimum
and maximum content sizes, growth, foreground, background and typography.
Lengths use device-independent pixels, subject to native user scaling. Padding
uses top/right/bottom/left insets. Gap applies equally on both axes. Alignment
sets the cross-axis content and child alignment; justification sets the main-axis
content alignment. Both support start, center and end. Growth applies equally
on both axes and takes precedence over alignment when extra space is allocated.
Maximum zero means unbounded, as in the native backend.

`Optional[T]{}` leaves a property unset. `Some(value)` sets it, including explicit
zero. Removing an override restores its theme or native default on the next
style update. `Some(Size{})`, `Some(float32(0))` and transparent `color.NRGBA{}`
clear size constraints, growth/gap and color respectively. Typography is an
atomic value: positive size, optional family, weight 0 or 100 through 900 in
steps of 100, and an italic flag. Empty family/weight 0 use the native default.

Resolution proceeds through theme default, theme kind, legacy `MinWidth`,
`MinHeight` and `Grow`, then explicit node style. Positive legacy minimum values
replace their respective axes; legacy `Grow=true` replaces growth with 1.
Legacy zeros/false leave theme properties intact. Without a theme, legacy
behavior is unchanged, including extension widget dimensions on untouched axes.
An explicit node `MinSize` or `Grow` can reset legacy values to zero.

## Application-local themes

`cogent.Driver.Theme` takes `ui.Theme{Default, Kinds}`. The driver validates it
and clones its kind map when the application starts. The snapshot is used by
all retained widgets in that run. No global theme or operating-system setting
is changed. Callers must not mutate the theme while starting a driver. The driver has no runtime
theme-switch API; node styles can change on any redraw. A managed extension that
receives a new theme snapshot propagates it through compatible retained panel
records, with validation before mutation and without replacing those widgets.

For example, callers can provide a consistent type size and container spacing:

```go
theme := ui.Theme{
    Default: ui.Style{Typography: ui.Some(ui.Typography{Size: 16})},
    Kinds: map[ui.Kind]ui.Style{
        ui.Row:    {Gap: ui.Some(float32(8))},
        ui.Column: {Gap: ui.Some(float32(8))},
    },
}
// Pass theme as cogent.Driver{Theme: theme}.
root := ui.Node{
    ID: "root", Kind: ui.Column,
    Style: ui.Style{Padding: ui.Some(ui.UniformInsets(12))},
}
_ = root
```

## Validation and retained updates

Padding, gap, growth and minimum/maximum sizes must be finite and nonnegative.
Typography size must be finite and positive; alignment must be supported. A
positive maximum smaller than its resolved minimum is rejected, including
conflicts assembled from theme, legacy fields and node overrides. Invalid styles
are rejected before mounting or synchronizing a native record. Style-only
updates preserve widget identity and trigger native restyling. Cogent resets its
native defaults before each styler run, so removed overrides do not linger.

Validation covers precedence, explicit resets, invalid/merged constraints,
cloned theme independence, native property application, real retained-widget
reset behavior, and preservation of unspecified extension dimensions. Native
build and tests establish adapter compatibility with the pinned Cogent revision;
font availability, platform appearance and cross-platform visual acceptance
remain backend-specific checks. This API does not promise CSS layout, arbitrary
font registration or main-axis space-between distribution.

## Application surface controls

`GrowX` and `GrowY` override the resolved `Grow` on their respective axes; each
remains an independently optional property through all overlays. `TextAlign`
positions Label glyphs or horizontal Button/MenuButton content at start, center
or end. It does not change container cross-axis `Align`. Rendering, text
selection and hyperlink interaction use the same origin. `CornerRadius` sets
all four corners in Dp; `Border{Width, Color}` supplies a solid outline with
straight-alpha sRGB color. Finite nonnegative widths/radii are required; explicit
zero resets an outline/radius. Omission restores normal theme/native defaults on
retained updates. Borders do not promise shadow or platform material controls.
Native keyboard-focus indicators retain precedence over the resting border,
including when its explicit width is zero.

Unset app-surface properties (`GrowX`, `GrowY`, `TextAlign`, `CornerRadius`,
`Border`) are omitted from JSON, preserving existing theme and replay identities.
Explicitly set zero values remain serialized and restore the corresponding
property on round trip. Existing properties retain their original JSON shape.
Optional omission follows `Set`, independently of an ignored `Value`.

Text allocation layout is opt-in through the new resolved `TextAlign`, `GrowX`,
or `GrowY` properties on a label. Removing all three restores legacy native
placement on the retained widget. Native control internals and scalar `Grow`
alone retain their existing paragraph origin and wrapping allocation behavior.
This avoids shifting inherited center/end paragraph alignment in existing apps.
Explicit allocation layout keeps drawing, selection and hyperlink geometry on
one origin. Button/MenuButton alignment uses container justification.
