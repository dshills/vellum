# Mac widget expansion review and checks

This development checkpoint adds SearchBox, FormField, InlineNotice, the opt-in
MacTheme and compact NumberField up/down buttons. The published alpha.3 tag and
artifacts remain unchanged. The owner requested personal-use widget expansion,
primarily macOS appearance, then supplied screenshots identifying title-bar overlap,
missing insets, malformed fields and the awkward horizontal counter.

## Review dispositions

Prism used Anthropic claude-sonnet-5-5, baseline none and complete coverage for
all retained review inputs. Raw outputs/receipts are preserved. These reviews
returned medium findings (exit 1); this is not a zero-finding review.

- ceea13a4bf130e47: gallery setup now uses explicit owner setters; its view does
  not mutate retained component configuration.
- 84eb00d782ff15eb: full-tree and nested-node tests verify Placeholder rejection
  on unsupported kinds. Existing validation already visits all nodes.
- 1f9d4cdef8aa673f: gallery returns a visible error descriptor instead of panic;
  Reset fields now resets name, search and quantity.
- 39135d34a41381bf: outlined-field changes propagate errors and call Update only
  when native Type changes. A retained-field test verifies immediate restyling,
  native identity and draft preservation when adding/removing the outline.
- 7f42d5ceeddeaf00: not applicable. The prepared core/typegen.go SetPlaceholder
  implementation is a plain field assignment and return; it performs no update,
  layout, draft replacement or notification. Unconditional assignment therefore
  does not have the hypothesized side effects. Draft/identity regressions pass.
- 1d28c78023437a74: not applicable to root keyboard stepping. The private arrow
  button is not a Tab target; root Up/Down use Spinner.StepFunc directly. Pointer
  presses record current direction for the existing repeat timer's positionless
  Click events. State is a live pointer, not a captured direction snapshot. Actual
  manual-clock lower holds repeat downward twice, then release without an extra
  step. Resetting direction after every activation would break that behavior.
- 737330344c48ebc4: verified and guarded. Inline mode bypasses custom MouseDown and
  stylers, restores callbacks/icons/StepFunc/tooltip, clears Invisible on the lead
  button and updates the native Maker. Tests verify identity and visible inline
  restoration after disabling/re-enabling and switching modes.

## Checks and observed behavior

Root portable race/vet passed; focused ui race/vet was renewed after API changes.
Native and offscreen gallery/control tests pass within their recorded scopes.
Final stepper checks include real routed upper/lower clicks, two held lower ticks,
keyboard stepping, bounded values, rejected/disabled proposals and presentation
restoration. Field geometry passes at 96/144/192 DPI, light/dark, with inset labels
and controls, full focus-border allocation and retained draft/identity. Relevant
adapter/gallery vet and native build pass. Original test-setup failures remain
local: missing deterministic colors, explicit-frame requirement, unsupported Node
JSON assumption and native test accessor/disabled-part assumptions were corrected;
no old failed result is reported as passing.

The standard OS title bar replaces overlapping unified content; the owner confirmed
it is draggable. Native light screenshots show padded groups, a wider outlined name
field, a compact up/down Copies field and an unwrapped notice/action row. CUA
accessibility exposes the ordinary labeled fields and root numeric stepper. This
is custom rendering, not an AppKit control or human VoiceOver qualification.

The final private apps use direct Go executables rather than shell launchers.
Signatures pass after clean temporary staging. Source binary and signed app hashes
are retained locally. No new release/tag/assets or public source push occurred.

## Remaining limits

Dark native visual inspection is still pending; its geometry/routing has automated
coverage. New native pointer verification of the final arrow hit regions is not
claimed: live tool actions encountered user-state guards; real routed pointer and
hold tests are offscreen evidence. Tool launch calls sometimes exceeded requested
timeouts, and initial blank views were observed before later populated screenshots.
Their cause is unresolved; no startup-latency guarantee or backend fix is claimed.
The final updated app is visibly populated and running. Further visual refinement
can follow personal use without restarting framework qualification.
