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
source-list checkpoint leaves rendering, pacing and release policy unchanged.
Managed input/button borders now reuse the same published activation state as
described below.

Activation binding is idempotent for the same event source. A changed publication
requests one originating composition; duplicate observations or a coalesced return
to the already-published state do not request another restyle or composition.

The standalone widget gallery returns focus to tag entry after accepted/rejected
entry and removal, so multiple tags can be entered in sequence. At MaxTokens the
entry is disabled and continuation moves to the color well. Disabling inputs
cancels pending root focus requests while preserving their monotonic sequence;
re-enabling does not replay them. This is application policy with existing APIs,
not a change to TokenField's reusable focus contract or ordinary text-field Return.

The demo displays tag validation feedback above the tag controls, keeping it
visible when entry focus scrolls near the viewport's bottom edge. Input validation
messages omit the internal ui: package prefix; diagnostic/configuration errors
and the reusable components' error values remain unchanged.

## Full token-toolbar keyboard regression

At the 32-token bound, the existing chip toolbar must reveal the last chip through
End navigation and send one removal proposal with its full stable ID and text.
The caller may then focus the newly enabled entry through a new FocusRequest.
Regression coverage uses long mixed-script labels, a 320-logical-pixel field,
light/dark themes and matched software window/DPI160 and DPI320 sizes. This is
routed offscreen geometry/owner evidence, not physical high-DPI qualification.

Known narrow-layout limitation: chip abbreviation is a fixed 24-rune budget, not
viewport-responsive. A long chip may exceed a roughly 200-logical-pixel viewport
and have clipped completed bounds. Preserve that finding in the integration
guide; responsive chip sizing is follow-up work, not part of the passing
matched-scale assertion. No public API, rendering policy or release gate changes.

## Oversized toolbar focus recovery

A focused chip larger than its toolbar viewport must not remain wholly offscreen.
The adapter adds a native Focus fallback using the current toolbar scope and
existing scroll APIs to reveal its leading edge, retaining the native scroll
margin. This applies to horizontal/vertical toolbar frames, managed focus requests
and ordinary keyboard focus; normal-sized controls retain native scrolling.
Bindings follow current ancestry and are cleared on reparenting; disabled or
retired scopes do not scroll. No owner-queue work, polling or public API is added.

The 32-token regression now includes 200/320 logical widths and checks unclipped
raw geometry for ordinary chips, visible leading geometry for oversized chips,
one full-value removal and an independent managed FocusRequest after removal.
A separate oversized vertical-action regression checks visibility, activation and
Home return. Narrow chips still have partial labels/trailing marks; responsive
truncation remains separate work, as disclosed in the integration guide.

## Conditional gallery feedback rows

The private gallery emits font/tag/date feedback labels only when their messages
are nonempty. Empty labels have no visible text bounds but still contributed
child layout/gap space: a completed-frame before/after comparison measured 28
pixels in the font group and 23 in the tag group at the pinned software font/DPI,
in both themes. Those measurements are fixture-specific, not universal widget
height promises. Shared control styling and public components are unchanged.

Actual tag entry/rejection/correction checks require visible feedback at the
focused entry and preserve that editor's native incarnation as the label is
inserted and removed. The mounted empty state contains no blank feedback labels;
font/date corrections and Reset clear their conditional rows as well.

## Retained entry visibility at the native initial window size

The standalone gallery starts at 660×740. Its Manual helper now uses that height
instead of 760. A duplicate-tag repro at matching dimensions retained the field
but clipped its shared outline from a 38-pixel initial row to 24 visible pixels.
The matched light/dark regression requires the unchanged row height in addition
to readable feedback, entry focus and native identity through correction.

The driver reuses accepted root FocusRequests to queue at most one main-scene
visibility intent per native window. A pre-child-paint check uses completed
positions and existing ScrollToThis, retaining native focus/selection. It cancels
on changed native identity, hidden/disabled/removed targets, emptied/failed
requests, a different scene or window closure. No continuous focus lock, owner
queue task, polling, renderer option, pacing or release-gate change is introduced.
Focus-only requests use ordinary NeedsRender so they settle even without layout
changes. Dialog-scene focus remains unchanged.

## Focus visibility cancellation and coalescing evidence

Additional routed Manual fixtures expand content while an accepted intent is
pending, then disable its ancestor, collapse its disclosure, replace its native
control under the same ID, clear the request or move native focus elsewhere.
Each clears the intent without scrolling the obsolete target. The replacement
is explicitly focused before paint, so identity cancellation is independent of
its focus flag. Invalid-request cleanup is tested at the locked request validator;
the runner's existing terminal invalid-descriptor policy is retained.

Three requests before paint replace the one pending intent and leave only the
latest target focused and fully visible. A second software window uses the same
local IDs; painting it cannot consume the origin's pending intent or alter its
scroll. The origin's later frame reveals its own field. These are automated
routing/lifetime checks, not physical-pointer or native multi-window qualification.
No runtime behavior or packaged binary changes in this coverage checkpoint.

## Inactive-window managed input borders

Explicitly styled TextField, NumberField, Button and MenuButton borders return
to their declared resting border while the originating native window is inactive.
Shared combo/secure-entry outlines follow the same rule, including accessory
button focus. An ordinary explicitly borderless field loses its native underline
while inactive and regains it on activation. Reserved maximum border geometry is
unchanged, so activation preserves control allocation, native focus, editing
selection and application values. No semantic edit or application mailbox drain
is caused by the appearance signal. Duplicate activation does not request paint.

Light/dark routed software regressions cover editor/accessory combo focus,
secure entry, ordinary bordered/borderless text, compact numeric stepping and
buttons. They compare painted border/margin pixels, excluding editable content
whose caret can blink independently. Synthetic selection is settled before frame
counting. Existing originating-window activation delivery/lifecycle coverage is
reused. These are automated routing/pixel checks; live inactive input-border
pixels have not been independently captured. Controls without explicit managed
borders, caret appearance and other native control classes keep their existing
policy. No public API, rendering backend, pacing or release gate changes.
