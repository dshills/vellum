# Mac field semantic cache correction

A clean relocated SDK from commit `03c3941` intermittently timed out before
admitting Undo. The owner held the committed Find value `日本`, while its actual
Cocoa accessibility tree still exposed the earlier preview `に`. The document,
selection, status and action availability matched. No Undo action was sent.
A bounded pristine reproduction passed once and then failed. A failure-only
probe retained generation 10 after 1,998 successful inspections and identified
the mismatching Find value. A final query returning a zero tree was not evidence
that native publication had never occurred.

The owner now copies the field's accepted editing state after outgoing native
focus-loss finalization and after controlled text/selection reconciliation.
Native publication shares that helper. An unchanged descriptor keeps live
preedit; a completed handoff keeps final value and directional selection without
reactivating the outgoing client or emitting a second Change. The editor smoke
retains its last successful inspection when reporting readiness failure.

The new semantic regressions failed before the correction for outgoing focus,
refocus and inactive controlled updates. Final focused field/focus-loss race20,
full driver race tests and vet, and editor race tests, vet and build pass. Three
complete actual native editor protocol runs pass on the same race binary,
SHA256 `b22dfa1af3c4302dc6d094df4e377b7cac2de261d4e1c6021e56e3eb5868b1d5`.
The six source/spec inputs stayed unchanged during qualification.

The first full driver gate exposed an older test that expected a stale semantic
cache after oversized input restoration. Its corrected assertion requires the
cache and published semantic value/UTF-16 selection to match the restored widget,
while retaining refusal, no-Change, history and valid-edit recovery checks.
That failed gate is preserved separately from final qualification.

[qualification.json](qualification.json) records exact inputs, commands, exits
and log hashes. [evidence.json](evidence.json) preserves original and final logs
as round-trippable UTF-8 JSON strings, including original CRLF bytes. These are
owned native protocol and semantic tests. They do not qualify human IME use,
VoiceOver, external AX, performance, or a public SDK release. macOS input sources
and accessibility settings were not changed.
