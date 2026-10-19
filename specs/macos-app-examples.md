# macOS application examples

## Scope

Four separate runnable Go consumers demonstrate a basic Calculator, plain-text
Notes, sample Weather and in-memory Reminders. They use public Vellum APIs;
application models do not import Cogent Core. Each exports New, View and Theme;
editable examples also export Commands. They share executable wiring, not private
widget implementations. Fixtures contain fictional data; no personal accounts,
location, Apple application storage or network services are read.

Source references are the macOS Tahoe 26 Apple application guides and an observed
Calculator on macOS 26.6.2. Reference screenshots are development evidence, not
redistributed application assets. SF Symbols and Apple artwork are not bundled.
The examples use the existing licensed Material library and three separately
attributed, pinned Apache-2.0 weather symbols.

## Framework additions

`app.Window.Chrome` is an immutable creation policy. The zero value preserves
existing decoration. `ChromeUnified` requests a transparent title-hidden native
unified header with OS-owned traffic lights. `AppearanceLight` and
`AppearanceDark` apply to that window only. The macOS driver rejects unsupported
policies. Offscreen sessions reject explicit native chrome. No OS appearance,
input source or process-wide theme preference is changed.

Width and height remain client-view Dp, including the reserved native header
under unified chrome. The original GLFW content view, delegate and Metal layer
remain attached. Initial policy is applied before native surface creation.
`app.Window.FixedSize` is a creation-time native user-resize policy. Its default
false preserves resizable windows; true maps to the existing Core
`Stage.Resizable=false`, `system.FixedSize` and GLFW nonresizable creation hint.
It does not set content minima/maxima or freeze backend/system geometry changes.
Initial Width/Height and OS size constraints retain their existing meanings;
manual/offscreen viewport sizing remains explicit. `app.Window.DisableRootScrolling` independently clips root content to the
native/manual viewport without root scrollbars. Inner `ui.Scroll` containers
remain scrollable. The default false preserves existing root scrolling; fixed
windows may still contain scrollable content.

Unified scenes omit the backend's default 8Dp inset. The native header remains
reserved for Cocoa dragging; Vellum content controls must stay below it.

These demos reserve 52Dp based on the actual tested Tahoe configuration. This
constant is not an adaptive safe-area API. Toolbar items, vibrancy, fullscreen
insets and exact material reproduction remain visual acceptance gaps. Calculator
history uses Command-Y and a dialog; its native header does not contain Apple's
mode/history toolbar items. The examples must not be called visually identical.

`ui.Style` additionally supports GrowX/GrowY (per-axis overrides of Grow),
TextAlign (Start/Center/End), CornerRadius and a solid Border. Lengths are Dp,
finite and nonnegative; explicit zero clears the relevant property. Properties
follow normal theme/kind/node overlay and retained-widget reset semantics.
TextAlign positions actual glyphs inside Label allocations and horizontal
Button/MenuButton content. Rendered text, rune selection and hyperlink geometry
share the same alignment origin. Multiline shorter lines retain their relative
alignment against the longest shaped line.

## Functional acceptance

- Calculator: basic decimal input and precedence, percent/sign/delete, error
  recovery, repeated equals and bounded detached history. Buttons and native
  keyboard commands share Press; generated exponent text cannot be edited into
  an invalid numeric token. Number tokens are at most 64 bytes, exponents ±300,
  and rational numerators/denominators at most 4096 bits after each operation.
- Notes: stable notes/editors, folders/search, create/delete/pin/title/body,
  bounded unified history and composition-safe refusal of navigation/history.
  Native actions settle search, title and body through public asynchronous
  driver barriers before changing selection, structure or app history.
- Weather: stable sample cities, bounded search, drift-free unit conversion,
  hourly and common-scale daily ranges; no live forecast claims.
- Reminders: stable reminders, filters/search, create/edit/complete/reopen,
  atomic validated drafts and identity-preserving Undo/Redo. Native actions settle
  mounted fields before validation. Changed or invalid drafts refuse Redo
  without destroying its history branch; lifetime cancellation releases actions.

## Evidence and completion policy

Portable/race/vet tests establish model and descriptor behavior. Driver software
pixel tests establish actual aligned glyph placement, keyed updates and removal;
Core tests cover selection/link origins. A hidden AppKit fixture checks default
policy, per-window appearance, original view/layer/buttons and client resizing.
Those checks do not establish physical pointer behavior or visual fidelity.
Native screenshot comparisons, pointer/keyboard tests and accessibility checks
are recorded separately with remaining mismatches. The examples join the demo
suite only with runnable commands and documented limits. Public release and Mac
framework completion remain subject to the existing delivery gates.

## Fixed Calculator viewport

Calculator requests a fixed 240 × 420Dp client with implicit root scrolling
turned off. An explicit fixed viewport takes precedence over a previous saved
size/maximized state while retaining restored position/screen selection.
Resizable windows and legacy fixed stages without valid explicit initial Dp
retain their previous restoration behavior. Native backend limits still apply.
`DisableRootScrolling` clips only the implicit Scene/Body/host; applications can
continue to use explicit `ui.Scroll` regions, including independent dialogs.
A software layout regression checks that all 20 keys remain fully visible and
routes actual pointer-down/up events at their completed bounds for 1 + 2 = 3.
This is automated routing evidence, separate from native desktop verification.
