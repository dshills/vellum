# Mac personal-use widgets — second expansion

The owner approved ColorWell, TokenField, FontPicker, PathControl and secure text
entry. This supersedes the older internal-app API freeze in CONTRIBUTING.md for
these five capabilities only. Existing rendering, pacing, release gates and public
alpha.3 artifacts remain unchanged. Each phase receives focused tests, Prism with
Anthropic claude-sonnet-5-5, and its own commit; the gallery is packaged and
relaunched after integration.

All controls use public Go descriptors and the existing managed Mac driver.
Appearance targets compact macOS controls in explicit light/dark themes, including
spacing, focus, keyboard access and readable disabled states. These are custom
controls, not promises of pixel-identical AppKit wrappers or automatic OS accent
tracking. Native visual smoke checks are separate from automated routing evidence.

## Color well

Compact color swatch opens the existing RGBA/hex editor in a same-window anchored
popover. Accepted color stays application-owned. Optional recent colors are
caller-owned and bounded; opacity uses straight-alpha sRGB. No shared system color
panel or background sampling service.

## Token field

Stable application token identities, removable rounded chips, text entry and
bounded completion suggestions. Return commits typed text or a chosen suggestion;
Tab reaches chip removal actions and the suggestion menu. Application accepts or
rejects edits, assigns new token IDs and controls the draft. This increment does
not implement Mail's inline token text editing or chip drag rearrangement.

## Font picker

Compact family/typeface/size controls and live plain-text preview. Applications
supply a bounded catalog, using installed font names. Controlled values, validation
and stale-catalog checks reuse existing menu/input callbacks. No system font-panel
ownership, font installation or automatic enumeration.

## Path control

Compact path segments with folder/document icons, ancestor menus and bounded
middle overflow. Stable domain IDs and application-owned navigation; reuse the
breadcrumb/menu lifecycle. No filesystem access, file permissions or path dropping
is implied by rendering a path.

## Secure text entry

Masked native-backed field, optional explicit reveal control, protected Mac
accessibility subrole, and no clear text in AX values or native substring queries.
Reuse normal managed text input, selection and lifecycle. Values remain ordinary
application-owned Go strings, not zeroizable secrets or a credential store. This
increment does not claim OS secure-event-input mode or protection from other
processes with debugging permissions. Test fixtures use synthetic values only.

## Overnight interaction hardening

Suggestion proposals are checked against the current TokenField catalog before
owner delivery, even when catalog changes have not yet reached View. Removed,
disabled, relabeled or otherwise changed options revoke queued menu actions.
Changing a SecureField's ID resets private reveal state: a different field starts
masked and requires a new explicit reveal action. This changes no public API,
rendering option or release gate.

ColorWell cancels its private open session when View observes a different ID,
Disabled, or a removed Change handler. Re-enabling/restoring the handler leaves
it closed until explicit activation. Old editor callbacks remain revoked.

Mac multiple-selection lists use flat full-width rows with no checkbox glyphs or
push-button outlines. The keyboard-focused row retains its focus outline. Selected
rows change together between active blue and inactive gray when keyboard focus
enters or leaves the list. Retained rows restore their original single-selection
presentation when the owner changes modes. No new public API or theme tracking.

Source-row text paints its glyphs over the row selection once. Inactive translucent
selection must remain uniform beneath labels, including on dark surfaces.

## Bounded geometry and routing checks

The five composites are checked inside Mac-themed forms at logical widths 480,
660 and 960, each at 160 and 320 logical DPI with a proportionally sized software
viewport, in both themes. Checks require nonempty contained composites, contained
interactive children and no typeface/size-stepper overlap. At every combination,
actual Manual pointer routing opens the color well's editor, its popup stays
inside the viewport, and Escape closes it and restores anchor focus. These are
deterministic offscreen checks; native display-scale qualification and arbitrary
narrow/oversized content are not inferred from them.

Long token chip labels abbreviate centrally at 24 runes to keep the trailing
remove affordance visible in compact Mac forms. Accepted text, duplicate checks,
removal proposals and accessibility names retain the full value. Multiple chips
continue to use the existing horizontally scrollable toolbar.

Each composite also has a mounted disable/re-enable routing check: disabled
targeted pointer input is refused, keyboard input produces no accepted edit, and
re-enabling retains the native control and permits explicit refocus/activation.
The color well reopens and Escape cancels without creating a color edit.

## Managed input focus appearance

Explicit TextField backgrounds stay flat while focused instead of gaining a
backend state-layer fill. An explicit nonzero border receives a full focus outline;
an explicitly transparent, zero-border embedded editor suppresses its private
bottom underline only when its direct enclosing row actually owns a shared
outline. Ordinary fields, including transparent borderless fields, retain their
native keyboard focus indicator. Removing a composite's border restores its
editor's own indicator without replacing the editor or its selection. Bordered
rows with a directly contained transparent borderless TextField share one outer
focus outline across the editor and accessory controls, including combo boxes and
secure entry. The outline allowance stays fixed between focus states. Unstyled
fields retain backend defaults; removing the explicit styles restores them.
This uses existing managed styles and focus routing, without changing public APIs,
rendering backend or release policy.

Numeric fields also retain an explicit theme background under focus in both
compact-stepper and plus/minus modes. Their existing border, hit geometry, numeric
validation and stepping policy remain owned by the Spinner adapter.

Source-list focus appearance uses the existing accessibility focus invalidation.
The originating window refresh reconciles all selected row colors, including
rows whose native focus state did not change. Draining the shared application
model queue must not directly restyle a native list.

Mac multi-selection source lists also mute selected rows when their native window
loses activation, retaining the keyboard-focused row and accepted selection. The
existing native attachment focus callback carries only a bounded latest state;
appearance is published and restyled on that window's render owner. Repeated
identical notifications do not restyle, and retirement cancels queued work. This
increment does not change other controls' window-activation cues, rendering,
pacing or release policy.
