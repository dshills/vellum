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
