# Mac widget demo

This private personal-use gallery includes color wells, token fields, font and
path pickers, protected text entry, and the earlier Mac input controls.
For application code, see the [public-API integration guide](mac-widget-integration.md).

After native source preparation, from `drivers/cogent` run:

```sh
go run ./cmd/gallery -mac-widgets
go run ./cmd/gallery -mac-widgets -mac-dark
```

The packaged **Vellum Mac Widgets.app** starts the focused light gallery directly;
**Vellum Mac Widgets Dark.app** starts the dark version. Scroll to reach all
sections. The ordinary macOS title bar remains draggable.

- **Secure entry:** the synthetic `Vellum-demo` value is masked. Show/Hide changes visible glyphs; accessibility and native surrounding-text queries stay protected. Enter demo text and press Return to commit.
- **Location:** use the ellipsis or segment menus to choose an ancestor. This changes the displayed virtual path only; relaunch to restore the sample.
- **Font:** choose Helvetica Neue, Times New Roman or Menlo, choose a typeface, and adjust size. The sample previews the accepted selection.
- **Tags:** enter a tag and press Return, or choose a suggestion. Click a chip (or Tab to it and press Return) to remove it.
- **Drawing color:** open the swatch; edit RGBA, opacity or hex, choose a recent color, and close with Done or Escape.
- **Drawing units:** type a custom value and commit with Return, or use the
  dropdown. Tab to the dropdown then Down/Return opens its keyboard menu.
- **Context actions:** right-click or Control-click the demo button. Shift-F10
  also opens its menu. Try Inspect or toggle Show grid.
- **Quick inspector:** open the anchored panel, press Done, Escape or release
  outside it. Keyboard focus returns to the button.
- **Objects:** click to replace selection, Command-click to toggle, Shift-click
  for a range. Command-Shift adds a range. Shift-Up/Down/Home/End extends by
  keyboard. Command-A selects all enabled objects. Selected rows become inactive
  gray when keyboard focus leaves the list or the native window loses activation;
  returning to the window preserves its focused row and selection. The reject
  checkbox demonstrates controlled values.
- **Date/time:** type `2026-10-24 14:30` and commit, or choose a calendar day.
  Month arrows browse without changing the accepted value. Tab/Shift-Tab and
  Return/Space reach calendar buttons; Escape cancels. Invalid dates leave the
  accepted value in place and display feedback.

**Reset fields** restores every sample value, selection, validation message and
rejection toggle, closes editors, and masks secure entry again. It preserves the
chosen light/dark appearance. Sample edits are not saved between launches.

## Current limits

New controls reuse managed components rather than shared system color/font
panels. Color recents are a caller-supplied palette (maximum 16). Token suggestions
use the dropdown; there is no per-keystroke completion filtering, comma-triggered
commit, inline chip editing or rearrangement. Long chips abbreviate their visible labels;
accessibility and removal keep the full value. In this standalone demo, accepted
chip removal returns keyboard focus to tag entry. Fonts use a caller-supplied catalog,
may substitute if unavailable, and support sizes 6–144 logical pixels. Path
controls describe virtual locations with explicit segment-count truncation, not
filesystem access or width-responsive collapse. Secure entry holds ordinary Go
strings and does not enable OS secure-event-input mode or erase secret memory;
use only synthetic text in this demo.


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
recorded separately in [checkpoint reviews](reviews/mac-personal-widgets/README.md).
