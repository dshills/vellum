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
