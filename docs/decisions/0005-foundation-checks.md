# ADR 0005: retain the foundation hold after remaining local checks

Date: 9 October 2026
Status: Provisional; G1 remains open

## Context

The user authorized remaining foundation checks before expanding the workbench. A separate [prototype](../../prototypes/foundationchecks/README.md) reproduces multiline grapheme and RTL glyph lookup failures, measures verified font discovery/coverage and memory, stresses native adapter lifetimes, probes the actual event transport, checks development packaging and compiles portable tests for Windows/Linux. [Results](../reports/FOUNDATION_CHECKS_RESULTS.md) distinguish narrow passes, confirmed failures and unverified requirements.

## Decision

Retain D as the experimental candidate, A as control and B as fallback. Keep two small text fixes isolated: logical multiline grapheme navigation/deletion and direction-independent cluster glyph lookup. Preserve the four-patch nativeinput baseline. Seal an owned development bundle correctly; do not treat an ad-hoc signature as trusted distribution.

Do not expand the workbench or accept G1. Exact text preservation, visual bidi/native geometry, full screen-reader workflows, real candidate-panel placement, supported platform lifetime hooks, loaded-font memory bounds, bounded lossless transport, native platform validation and clean distribution remain requirements. No upstream submission or production API commitment is made by these local patches.

## Consequences

Race tests and narrow diagnostics now establish correct representative logical deletion and cluster lookup. Verified font shaping requires roughly 257 MiB live under the current fallback policy. Native input actions are lossless in the probe but queue without a limit. Correct autorelease-pool control removes most apparent lifetime growth; residual native retention remains unattributed and GPU lifetime is not covered.

Prioritize foundation contracts and evidence access, then reevaluate whether a modest upstream extension remains plausible. If supported hooks/cache ownership/transport require broad toolkit replacement, revisit B using the same acceptance criteria. Desktop automation and Windows/Linux test hosts are needed for outstanding native checks; portable cross-compilation does not retire those risks.
