# macOS widget expansion — personal-use phase

Owner direction, 23 October 2026: expand available widgets for personal use, with
macOS as the primary platform and macOS look and feel required. This starts a
post-preview development branch from 320376e. The published alpha.3 artifacts and
historical preview qualification plan remain unchanged. New widget work is
authorized; publishing another release, distribution work and changing rendering
or pacing are outside this checkpoint.

## First increment

Add reusable public Go components composed from the existing retained primitives:

- SearchBox: magnifying glass, editable query/hint and an accessible Clear action.
- FormField: visible label, required indication, help and caller-owned validation
  text associated with one focusable managed leaf's accessibility description.
- InlineNotice: information, success, warning or error message, optionally with
  one application-owned action.
- MacTheme: opt-in compact light/dark styling that applications can share.
  Reuse NumberField for numeric stepping rather than duplicating its API.

These use existing window ownership, native text ingress, focus, callback queues
and lifecycle guards. No global settings, input sources or backend changes.

## Contract and bounds

SearchBox and InlineNotice are owner-held components. Their View methods return
fresh descriptors or an error and allocate a fixed number of managed nodes.
Generated IDs occupy root/search/ and root/notice/. FormField occupies root/field/
for help/error labels; its Content keeps its caller-owned stable ID. Ordinary tree
validation catches collisions. SearchInputID and SearchClearID expose normal
managed focus/automation targets. Do not share a mutable component across owners.

Search edits and Clear propose SearchEvent through the current Change handler.
The component does not mutate accepted Text. It uses the existing TextField commit
semantics, not a new per-keystroke stream; no suggestions, debounce, recents or
background search service is added. Empty Clear is disabled and does not propose
another edit. Nil Change disables the component. Queued actions recheck current
component identity, Disabled and callback; native removal/ancestry guards remain
responsible for retired managed records.

Node.Placeholder is a literal TextField-only hint, never accepted text or an
accessible name. Empty hints preserve the existing no-hint behavior.
Hint updates must preserve accepted values, retained identity and private drafts.

FormField wraps a focusable leaf (including existing single-line/numeric/choice
controls). It adds no focus stop. A supplied accessible name is preserved;
otherwise Label names the content. Existing Help, required status, hint and error
are included in the content's accessible description. View copies, never mutates,
the caller's content. Required does not run a validation engine. Validation text
is descriptive, not a new native live-region announcement.

InlineNotice has at most three children and one optional action. Tone must be
known. ActionLabel and Activate must either both be supplied or both omitted.
Passive messages add no focus stop. Current callbacks are read on owner delivery.
Disabled and changed component IDs refuse obsolete proposals.

## Mac appearance and evidence

The first styling pass uses restrained neutral surfaces, compact 13-point
Helvetica Neue typography, rounded input borders and existing focus indicators.
The font comes from macOS; it is not bundled or claimed to be San Francisco.
Light/dark choices affect this app/window only. MacTheme returns a fresh map;
explicit node styles still win. Color values are fixed preview choices, not
dynamic NSColor, user accent/high-contrast tracking or an exact AppKit replica.

The components remain custom rendered by the existing Cogent/WebGPU/Metal path,
with the existing macOS native window, text and accessibility integration.
They are not NSSearchField/NSTextField/NSButton wrappers. Do not claim pixel parity,
human VoiceOver qualification or physical pointer acceptance from software tests.

Design references:
[Apple search fields](https://developer.apple.com/design/human-interface-guidelines/search-fields),
[text fields](https://developer.apple.com/design/human-interface-guidelines/text-fields)
and [focus/selection](https://developer.apple.com/design/human-interface-guidelines/focus-and-selection/).
The specific implementation metrics above are project choices.

## Acceptance

- Portable race tests cover acceptance/rejection, Unicode, clear/readonly/disabled,
  current callbacks, stale component IDs, bounds and generated-ID collision.
- Driver offscreen routing covers actual mouse/key events, frame change, retained
  identities, label/error metadata and removed/disabled queued callbacks.
- Gallery consumer uses only public Go APIs and exercises filtering, validation,
  reset, rejection and disabling. Runnable focused Mac view has light/dark options.
- Inspect the actual Mac view in both appearances and record what was exercised;
  preserve the distinction from human physical input or native-control identity.
- Focused native build and affected package vet; Prism Anthropic Claude Sonnet 5.5
  review with finding dispositions, then a separate development checkpoint commit.
  No full release campaign or new SDK distribution is required.

Later widget candidates: editable combo box, date/time picker, context menu/popover,
multi-selection list and richer inspector rows. Select subsequent work from actual
personal application needs; these are not part of this increment.

The Mac gallery uses the standard OS title bar; content must stay below its drag
region. Form groups have 12dp interior padding; full field outlines use the native
outlined-field layout so all sides reserve the focus-border space. Text inputs
grow within their group, while numeric entry remains compact.

NumberField optionally sets NumberButtons: NumberButtonsStepper for compact
up/down arrows at the right. Upper/lower halves increment/decrement; ordinary
keyboard stepping uses the same clamped bounds and does not wrap in this mode.
The existing numeric editing/owner callback and root stepper semantics are reused.
The private arrow control adds no Tab stop. Removing the option restores inline
minus/plus controls on the same retained numeric field.
