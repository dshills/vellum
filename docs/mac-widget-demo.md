# Mac widget demo

This private personal-use gallery includes Sidebar, Sheet, Collection, Shortcuts
and Levels pages, a Toolbar catalogue, plus the earlier controls on the Inputs page. All controls
are reusable through public Go APIs.
For application code, see the [public-API integration guide](mac-widget-integration.md).

After native source preparation, from `drivers/cogent` run:

```sh
go run ./cmd/gallery -mac-widgets
go run ./cmd/gallery -mac-widgets -mac-dark
```

The packaged **Vellum Mac Widgets.app** starts the focused light gallery directly;
**Vellum Mac Widgets Dark.app** starts the dark version. Choose a page in the
top tab strip. Each page fills the available width and uses its fixed light/dark
window surface; scroll within Inputs to reach its sections. The ordinary macOS title bar remains draggable.


- **Toolbars:** compare Small (26 pt), Regular (32 pt), and Large (40 pt) action
  heights. Full-color icons start enabled; turn them off to compare monochrome
  symbols. Switch Icons / Text / Icons and text / Icon above text without recreating
  actions. The stacked option centers artwork over its caption in taller controls;
  scroll the page to see the lower examples.
  A mixed document bar shows per-item choices; a vertical bar demonstrates scrolling.
  Narrow preview constrains horizontal bars to 280 pt. Overflow menu starts on;
  click the trailing » to see hidden actions with full-color icons and labels.
  Turn Overflow menu off to compare scrolling. Tab enters actions; arrows,
  Home and End move/reveal enabled tools. Delete is intentionally disabled.
  Choose Customize… beside a horizontal toolbar to reorder actions by dragging the leading grips (or using arrows),
  show/hide items, or Restore Defaults. Changes are independent per toolbar and
  live; Done/Escape closes the panel and keeps them. Toolbar order and hidden actions persist across launches, separately for light and dark apps.
  Right/Control-click a toolbar (or Shift-F10 on an action) for per-bar display
  options and Customize Toolbar. Mixed demonstrates fixed and flexible spaces.
  Actions update a counter only; no document files are opened or changed.
- **Sidebar:** choose an icon row to navigate. Section triangles collapse/expand
  groups; Tab visits section lists and arrows choose enabled rows. Archive is
  deliberately disabled. Badges show sample counts. This demonstration navigates
  virtual locations only.
- **Sheet:** choose Export drawing… to open a window-modal card. Cancel/Escape
  dismisses, Export confirms once and increments the sample count. Return activates
  the focused action or an unconsumed default; text editors keep their key handling.
  The card is centered and its default action receives initial focus when there
  is no editor. Focus returns to Export drawing… after closing. No file is written.
- **Collection:** resize the window to change the number of columns. Click selects;
  Command-click toggles and Shift-click extends. Arrows follow the grid, Shift
  extends, Command-A selects enabled items. Return or double-click increments the
  opened-item counter. Rejection and disable switches demonstrate controlled state.
- **Shortcuts:** activate Duplicate or Export's recorder, wait for Type shortcut…,
  then enter a chord such as Command-Shift-K. Escape/Tab cancels and Delete clears.
  Assigning the other action's chord demonstrates conflict rejection. The sample
  reserves Command-Q/Command-W. Recorded values edit preferences only; they do not
  register global shortcuts or change the app's commands. Reset restores samples.
- **Levels:** move the sample capacity slider through warning at 75 / critical at 90; continuous
  and segmented meters change together. Tab to a star and press Return/Space to
  propose a rating. Zero is distinct from Clear/Unrated. Rejection and disable
  controls keep application ownership visible.

The earlier controls are on **Inputs**:

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
accessibility and removal keep the full value. Abbreviation is fixed rather than
width-responsive; at very narrow widths a single chip can still be clipped.
Keyboard focus scrolls oversized chips into view at their leading edge; Return
or Space removes the focused chip even if its trailing remove mark is clipped.
See the [integration limits](mac-widget-integration.md#limits) for the reproduced
case. In this standalone demo, tag acceptance, rejection and chip removal return focus to tag entry for repeated
entry. Font, tag and date feedback rows appear only while an error is present.
Tag validation feedback appears above the controls so it stays visible
while correcting an entry near the bottom of the viewport. At 32 tags, focus advances to Drawing color; removing a tag enables entry
again. Disable inputs cancels pending focus requests. This is demo focus policy.
Fonts use a caller-supplied catalog,
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

Explicitly styled input and button borders also return to their resting appearance
when the native window becomes inactive, preserving keyboard focus and text
selection for your return. This includes the shared secure/combo outlines and
numeric field. Automated light/dark pixel and routing checks cover this behavior;
live inactive-border pixels have not been independently captured. This does not
change caret styling or the native defaults of other control classes.

The newer pages also use custom rendering and existing native input/accessibility
integration. Sidebar surfaces use fixed light/dark colors without translucency;
arrow navigation stays within a section. Collection is bounded to 256 items and has
no virtualization, marquee selection, item dragging or Quick Look. Sheets are
same-window managed modal cards, without AppKit sheet attachment/animation.
Shortcut recording is one chord, ASCII/named keys or F1–F24, local to the focused
active window; OS-intercepted keys may never reach the app. Ratings use accessible
star buttons with Tab/Return/Space, without pointer-drag or arrow adjustment.
See the [control reference](controls.md) for public APIs and
[independent integration example](../examples/sdk-contract/mac_widgets_next_test.go).

## Native titlebar experiment

The current personal package also enables the
[native titlebar toolbar prototype](native-titlebar-toolbar-prototype.md) and
starts on Toolbars. Its full-color titlebar commands update the sample action
status. Content-toolbar controls and saved customization remain independent.
Use `-native-titlebar=false` to launch the content-only gallery from source.

Use **Customize Titlebar…** to open the native macOS customization palette, or
right/Control-click the titlebar toolbar. Drag tools to reorder/remove/re-add them
and use the options AppKit offers; it remembers the item layout across launches separately
for light and dark. This button also recovers a hidden native toolbar. Full-color
artwork is retained. Native customization requires macOS 15+.

On the tested macOS26 compact titlebar, native customization remains icons-only;
the content toolbar display buttons do not change native chrome. See the native
prototype guide for platform/style limits and the verified persistence scope.
