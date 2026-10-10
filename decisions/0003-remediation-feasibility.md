# ADR 0003: continue bounded native-input work after local remediation

Date: 9 October 2026
Status: Provisional; representative field remediation demonstrated, G1 remains open

## Context

[ADR 0002](0002-foundation-gaps.md) held acceptance for text, scheduling and memory. The separate remediation fixture now verifies whole-grapheme single-line deletion, navigation, Undo and actual native UTF-16 selection. Three reviewable local Cogent patches add editing-state access, lazy spelling and opt-in demand painting. Matched 60-second diagnostics measured 0.284% of one CPU for demand versus 1.804% for periodic scheduling; dictionary deferral reduced initial retained heap from about 403 to 277 MiB. See [results and evidence](../REMEDIATION_RESULTS.md).

The adapter's former use of the committing `Text()` getter was corrected. Inspection now preserves live editing state. Seventeen top-level race-enabled tests, vet, native build, six resizes and a limited window-transition smoke passed. These remain small local diagnostics. P2/P3, font-memory budgets and the platform/distribution matrix have not passed.

## Decision

Keep D as the current experimental candidate, A as control and B as fallback. Continue one bounded native-input feasibility step before P4–P8 expansion. Use the new editing-state seam to investigate marked text, native selection and screen-coordinate caret delivery for one field; exercise a real Japanese/Chinese IME and VoiceOver workflow when that bridge exists. Secure native creation/focus/destruction and bounded lossless action transport. Investigate font/cache bounds with the full required corpus, then extend lifecycle and idle coverage.

Keep the original foundation intact and the three patches independently reviewable. Demand painting remains opt-in. Seek supported or upstream-accepted hooks before committing to production reuse; no upstream issue or PR has been posted by this trial. If native input requires substantial platform/text replacement or an enduring unsupported backend fork, run B against the same native editing/accessibility requirements before deciding.

## Consequences

The reproduced single-line failure is locally resolved and idle scheduling has a plausible low-cost path. Optional spelling still costs the same when initialized; font memory remains above budget. Native text geometry, multiline/bidi editing, IME composition, assistive-technology completion, robust animation and cross-platform packaging remain gates. This ADR does not accept G1, promise production support or commit the prototype API.

## Follow-up

[ADR 0004](0004-native-input-prototype.md) records the implemented Cocoa composition client and its local checks. Live CJK/VoiceOver and the remaining foundation gates are still open.
