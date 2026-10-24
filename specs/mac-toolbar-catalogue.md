# Mac toolbar catalogue

Scope: Mac toolbar examples and public content-toolbar helpers. Rendering,
title-bar and release-policy scope remains unchanged.

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
- Horizontal examples enable WithToolbarOverflow by default. Actual width fits a
  prefix, reserving a trailing More actions chevron when needed. Its flat menu
  preserves labels, symbol/full-color artwork, stable IDs, disabled state and
  callbacks. No native action reparenting/deletion occurs on width changes.
- Overflow menu checkbox restores the scrolling comparison; vertical bars scroll.
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

Overflow menus are bounded to 64 Buttons/Separators in a horizontal toolbar.
WithToolbarOverflow is applied after mixed item presentation is chosen. No nested
editors, placement sources, context-menu hosts, vertical menus or title-bar merge.
Hidden actions have no main-window focus/AX target; menu actions appear only for
the active popup lifetime. Resize/content changes, disabling and removal revoke
stale actions. Menu row artwork is 18 logical pixels, contained and never tinted.
Checked items with artwork show a separate check indicator; plain checked items
retain the existing marker. Menu option IconImage is immutable and exclusive with
Icon. Separators cannot carry either artwork or action state.


Toolbar customization uses public ToolbarLayout (partial Order/Hidden stable ID
lists) and a retained owner-held ToolbarCustomizer. Apply the value to the fresh
complete flat horizontal Buttons/Separators catalogue before overflow. Normalize
separators, preserve artwork/styles/callbacks and visible retained identities;
hiding removes controls from both toolbar and overflow. Unknown/duplicate IDs and
unsupported sources are rejected. Hidden actions can be reordered; showing an
action recreates the same public identity. All-hidden is valid with an external
opener. Max 64 catalogue items, at least one action. New omitted IDs append in
catalogue order; stale unknown IDs require caller migration.

The same-window sheet has full-color decorative artwork, named visibility
checkboxes, leading drag grips and earlier/later buttons, Restore Defaults and Done. Edits emit
controlled proposals against the latest accepted Value, suppress no-ops, and
never invoke toolbar commands. Disabled actions remain configurable. Done/Escape
retains edits, revokes the session and restores focus. Disable, ID replacement,
invalid descriptions and native disposal also revoke stale input. Small, Regular,
Large and Mixed have independent gallery values and openers. No disk preferences, direct toolbar dragging, cross-window customization or NSToolbar system palette.

Portable tests cover normalization, malformed layouts, callback identity,
accepted/rejected/queued edits, restore, hidden-all and session revocation. Public
SDK consumption is separate. Native Manual routes real pointer events through
both-theme sheets, checks retained identity/artwork on reorder, removal from
ordinary and overflow actions, focus restoration, disabling and window teardown.
Packaged screenshot/AX/native-key observations remain separate evidence.


Customization-row drag contract: generated ToolbarReorderHandle capabilities are
stable across harmless View refreshes, bounded by the 64-item catalogue and
revoked on order/session replacement. Native primary pointer capture begins on a
grip; six logical points activates the insertion preview. Full row midpoints and
the current clipped viewport determine insertion before a row or at list end,
including under scrolling/display scaling and the modal scene origin. Native
routing consumes every captured release before ordinary click dispatch. A valid
release queues one owner-checked DropBefore proposal against the captured order
and current Hidden membership. Rejected/no-op changes do not edit caller state.
Outside release, Escape, capture loss, source/target removal, order replacement,
disable, sheet dismissal and window closure remove capture/preview and references.
Native reads use immutable owner scope publications; application callbacks cannot
inspect native parents. Wheel scroll is confined to the list; moving at an edge
scrolls a bounded step. No stationary-edge autoscroll timer is provided. Existing
arrow controls remain keyboard/AX reorder access. Preserve original artwork and
all hidden actions. Public Manual regression tests cover both appearances, two
scales, long-list scrolling, threshold/no-op/duplicate release, outside release
consumption, Escape and lifecycle cancellation; packaged pointer checks remain
separate evidence.

The personal Mac gallery opts into versioned application-owned preferences at
UserConfigDir()/Vellum/MacWidgets/toolbars-{light,dark}.json. Pure constructors and
portable controls perform no I/O. Each accepted layout edit writes through a
0600 temporary file, sync, close and atomic rename; failures leave in-memory edits
usable and show a visible error. Loading is bounded to 64 KiB/four layouts/64
items. Retired and duplicate IDs are removed and new catalogue items use their
default order. Invalid files and unsupported schema versions are left untouched; persistence is
disabled for that session until the file is repaired/removed and the app relaunched.
Restore Defaults is persisted through the same Change path. Appearance packages
have separate files to avoid concurrently overwriting each other's settings.

ToolbarSpace and ToolbarFlexibleSpace are noninteractive leaf kinds, with default
16 pt fixed width and 8 pt minimum flexible width. Flexible space consumes spare
horizontal toolbar width and contracts before actions overflow. Both use stable
IDs and participate in customization ordering/hiding/persistence. Neither has a
focus stop, accessible action, command or overflow menu entry. An all-hidden action
catalogue produces an empty bar even if spaces remain configured. Mixed toolbar
demonstrates both types; simple Small/Regular/Large comparisons remain unchanged.

Toolbar.ContextMenu reuses existing managed-menu validation/lifetimes. Right/
Control-click over empty toolbar areas, spaces or ordinary actions routes to the
nearest toolbar menu; a child's explicit context menu takes precedence. Shift-F10
on a toolbar action opens the menu, ignoring CapsLock/NumLock flags. Menu opening
preserves the focused child. Disable, option replacement, removal and window
teardown revoke menu capabilities. Gallery menus expose checked per-toolbar
Icons Only, Text Only, Icons and Text, Icon Above Text, and Customize Toolbar.
Mixed also offers its original mixed presentation. Per-toolbar display overrides
and the common display choice are persisted in the version-1 preference document.
Common gallery display buttons clear overrides. Opening customization from a menu
uses the same retained customizer and external return-focus target.

Customization polish uses bundled SymbolGrip (six dots), a native grab cursor,
and a clipped blue insertion marker with a 100 ms ease-out/fade transition. A
capture retains at most one core scene Animation; retargeting reuses it and every
completion/cancellation synchronously removes animation, sprite and captured
references. Animation reads native scene time, never goroutine timers or model
callbacks. No backend pacing/rendering policy changes. Scoped, explicitly bordered
buttons/checkboxes paint focus only when they own actual focus, are enabled and
belong to the current modal/tab input projection. Resting/focused border allowance
is stable. Background openers and stale native Focused flags cannot paint extra
focus cues. Reduced-motion system preferences are not queried by this demo.

While list capture is active, LongHoverStart/LongPressStart help on every managed
source/target row control is consumed. The shared list capability is removed on
all terminal paths; normal help then resumes. This prevents tooltip-stage
retirement during a held drag. Native target-grip/checkbox/arrow regression
coverage is separate from the packaged held-pointer observation.
