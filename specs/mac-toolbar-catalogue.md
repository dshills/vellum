# Mac toolbar catalogue

Scope: application-local examples built with existing public nodes. Framework,
rendering, title-bar and release-policy scope remains unchanged.

- Toolbars page in both Mac widget packages, with three size recipes and four
  display modes. Small/Regular/Large control heights are 26/32/40 pt; font sizes
  12/13/16 pt. Icon-only controls are square. Text uses intrinsic width.
- Full-color artwork is the initial icon mode. A comparison checkbox switches
  to monochrome symbols without replacing native actions. Six original 96 px
  sample images are created once and shared; no asset I/O occurs in View.
- Root and bar backgrounds follow the explicit demo appearance. Actions have
  compact rounded, borderless tool treatment and ordinary focus/hover feedback.
- Per-item icon/text choice, separators, mixed bar, vertical bar, disabled Delete,
  global disable and a 280 pt narrow preview are visible and interactive.
- Each size has its own stable IDs; display changes keep those IDs. Every icon action has a name and
  tooltip. Callbacks update owner-held count/status once; no file or global action.
- Retained overflow scrolls and reveals focused actions; no replacement menu or
  item reparenting. Existing toolbar event routing and lifecycle remain in charge.
- Public Manual tests compare geometry in both appearances, verify stable native
  incarnations through all display modes, skip disabled Delete, reveal End/Home,
  and admit no action while disabled. Live observations stay separate.

Usage, recipe and deliberate limitations: [Mac toolbars](../docs/mac-toolbars.md).

Icon-above-text uses public Button.IconPlacement=ui.IconAboveText with a required
nonblank caption and either IconImage or Icon. Empty/IconLeading restores inline.
The gallery's stacked Small/Regular/Large action heights are 52/60/72 pt, widths
64/72/84 pt, with the existing 12/13/16 pt caption sizes. The private icon slot
uses 2 em and managed gap defaults to 4 pt; images contain without recoloring.
Existing actions, callbacks, IDs, caption widgets and semantics stay retained.
MenuButton and ToggleButton do not accept IconPlacement. The taller catalogue
scrolls to reach its lower samples; mixed/vertical bars keep their chosen layouts.

Regression coverage checks both themes, symbol/color sources and all three sizes:
icon/caption bounds fit and center without overlap, color survives, icon and label
each activate once, Return activates once, and leading/default restoration keeps
identity and native icon sizing. Gallery Manual routing covers stacked narrow
End/Home revelation with disabled skipping. Live screenshots/keyboard checks are
recorded separately from these automated checks.
