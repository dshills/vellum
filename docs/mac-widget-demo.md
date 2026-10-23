# Mac widget demo

This private personal-use gallery includes the five desktop widget additions and the personal-use color well, plus
the earlier search, validation and compact up/down counter controls.

After native source preparation, from `drivers/cogent` run:

```sh
go run ./cmd/gallery -mac-widgets
go run ./cmd/gallery -mac-widgets -mac-dark
```

The packaged **Vellum Mac Widgets.app** starts the focused light gallery directly;
**Vellum Mac Widgets Dark.app** starts the dark version. Scroll to reach all
sections. The ordinary macOS title bar remains draggable.

- **Drawing color:** open the swatch; edit RGBA, opacity or hex, choose a recent color, and close with Done or Escape.
- **Drawing units:** type a custom value and commit with Return, or use the
  dropdown. Tab to the dropdown then Down/Return opens its keyboard menu.
- **Context actions:** right-click or Control-click the demo button. Shift-F10
  also opens its menu. Try Inspect or toggle Show grid.
- **Quick inspector:** open the anchored panel, press Done, Escape or release
  outside it. Keyboard focus returns to the button.
- **Objects:** click to replace selection, Command-click to toggle, Shift-click
  for a range. Command-Shift adds a range. Shift-Up/Down/Home/End extends by
  keyboard. The reject checkbox demonstrates controlled values.
- **Date/time:** type `2026-10-24 14:30` and commit, or choose a calendar day.
  Month arrows browse without changing the accepted value. Tab/Shift-Tab and
  Return/Space reach calendar buttons; Escape cancels. Invalid dates leave the
  accepted value in place and display feedback.

## Current limits

These are custom-rendered controls using existing macOS window, text and
accessibility integration, not AppKit NSControl wrappers or pixel-exact replicas.
Appearance is an explicit light/dark theme with fixed colors and Helvetica Neue;
OS accent, high-contrast and automatic appearance tracking are not implemented.
Context menus are flat and supported on Button, Canvas2D and ListBox. Popovers
require same-window Button anchors and clamp at edges; they do not detach,
automatically switch their direction marker or cross windows. Lists are bounded
to 256 rows. Calendar is English/Sunday-first, with caller-selected date layout
(default ISO/24-hour); automatic OS locale and calendar arrow-grid navigation
are not available. The combo field opens suggestions through its separate
keyboard-focusable button; Down inside the text entry does not open it.

The private package is ad-hoc signed and unnotarized. It uses this development
machine's prepared native dependencies; it is not a new public release artifact.
Automated offscreen event/geometry checks and live macOS observations are
recorded separately in [checkpoint reviews](reviews/mac-desktop-widgets/README.md).
