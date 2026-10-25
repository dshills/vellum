# Personal macOS app stabilization and Notes visual correction

The current private Notes build6 is independently consumable from a reviewed
SDK snapshot, tested, ad-hoc signed, signature-verified and left running.
Package/source identities are in [delivery.json](delivery.json). The source
checkpoint is `0402ec8e60f7a2b77005241cc6d94a527a94e4b3`; both
[Framework CI](https://github.com/dshills/vellum/actions/runs/40212664189) and
[Platform adapters CI](https://github.com/dshills/vellum/actions/runs/40212664232)
passed for that exact commit. No new public release, downloadable assets or tag
was published. These are private delivery observations, not full Mac acceptance.

## Completed checkpoints

- Repair older-macOS-SDK toolbar compilation and Windows path/permission tests.
  Preserve immutable replay theme fingerprints and route the legacy visual test's
  paints through its owner loop instead of concurrent direct rendering.
- Update vulnerable maintained Gio prototype pins and classify the copied-native
  notice alert without changing original upstream notice contents. The default
  branch alerts do not close until that branch receives these changes.
- Verify Dark bundle launch and physical Find shortcut routing in the existing
  widget bundle, separately from Notes input evidence. Global settings unchanged.
- Make Notes a public-API consumer with guarded local files, strict versioned
  decoding, atomic save/open, private storage locking, interrupted-session
  recovery, input settlement and shared application Undo/Redo. Fix duplicate
  launch control IDs and obsolete recovery after clean Redo.
- Document independent application construction and package its source beside an
  exact private SDK archive with all native linker replacements verified.

The owner's later Apple/Vellum comparison exposed presentation gaps. Personal
Apple Notes screenshot contents remain outside repository source and reviews.
The app now uses compact rows, trailing folder counts, correct folder symbols,
a small search field, pale amber selected notes, fine dividers, a smaller editable
heading and less chrome whitespace. Standard native chrome owns dragging and
window buttons. The owner explicitly approved one focused framework exception:
`MultilineText` may select proportional `Prose` presentation with wrapping and
no line-number gutter. Existing code editors keep their default behavior.

The writing surface retains the same exact buffer, public model, native input
and history path. Wrapping uses actual pixel widths and source-preserving shape
boundaries. Carets, selection and drawing share proportional glyph advances.
Resize invalidates obsolete visual-line affinity. Native checks found missing
emoji and an inappropriate outer focus box; the final app explicitly includes
emoji fallback, and frameless prose uses its caret/selection for focus feedback.
Code and other framed controls retain their native focus indicators.

## Evidence and concrete repairs

Every implementation checkpoint has a Prism/Anthropic Claude Sonnet5.5 review
and focused checks under [the review index](../../reviews/mac-personal-widgets/README.md).
Required portable race/vet and all Mac example model checks passed. Focused
offscreen tests verify real Notes geometry at1100×740 and1380×900, variable
glyph advances, exact source fragments, grapheme-safe wrap boundaries, pointer
geometry, width/160–200 DPI reflow, retained binding/composition/CRLF/history,
legacy code typography and focus appearance transitions. Clean preparation
checks patch input/output digests and the contiguous stage sequence.

The independent build verifies the complete module origins, module integrity,
consumer race/vet/build, selected native capsule identity, binary-bound private
notice presence and strict signatures after exact ZIP extraction and installation.
The new prose option is absent from the earlier frozen SDK, so this delivery
includes a private source-identified snapshot rather than relabeling that old
archive. Its base version/folder name remains personal-v0.1; it is not a newly
published SDK release or a new distribution-audit decision.

[Build5 observations](native-build5.json) verified native multiline paste, app
Save/Undo/Redo and text-range selection, then exposed the two visual defects.
[Build6 observations](native-build6.json) verified their correction, including
color emoji in a semibold title, list row and normal body, focused caret/selection
without a surrounding box, exact Unicode Save on disk and shared Undo/Redo on
the same note ID. Both disposable notes were removed through app actions. The
original note selection/body was restored and saved; final disk checks confirmed
seven original notes and no disposable n11/n12. The final bundle remains running.
Native Open/recovery/close checks from build3 are reused because those paths
were unchanged by this presentation work. No Apple Notes data was accessed.

## Remaining limitations and backlog

- Physical Notes body-click focus remains unqualified: computer-use coordinate
  clicks missed their intended focus target. Keyboard Tab/native input and
  accessibility range selection passed. Offscreen pixel geometry is separate
  evidence and does not establish physical pointer behavior.
- Human IME and VoiceOver assessment remains explicitly unqualified by owner choice.
- Background computer-use launch initially captured a blank client; raising the
  window painted the expected content. User-visible first-frame behavior is not
  independently established by that background-capture observation.
- Plain text has a separate title; fixed sample folders, title-button-only card
  selection, no rich formatting/attachments/accounts/cloud/translucent materials
  and no exact Apple toolbar reproduction. Prose does not supply fixed-column tabs.
- Earlier Framework run40211682154 failed
  `TestTreeFrameMetricsAdmissionAndCanceledDelivery` at its cancellation-metrics
  assertion. The subsequent package-source run passed without a related tree
  change. Treat this as an unresolved intermittent CI assertion, not evidence
  that Notes rendering failed or that the underlying cancellation case is fixed.
  Retain the failure log in private task artifacts for focused follow-up.
- Packaging is private, ad-hoc signed and unnotarized; no expanded public native
  distribution audit, human qualification or release performance claim is implied.
