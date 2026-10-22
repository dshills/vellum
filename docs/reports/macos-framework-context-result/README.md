# Complete timed journal — phase admission rejected

The context-enabled consumer completed all 128 warmup and 120 measured routed
inputs, with native exit 0, empty stderr and zero foreground losses. The original
input/backend/display-clock validators pass. The stricter phase control rejects
this run: 155 of 248 actual receipts exceed the unchanged 2 ms phase bound
(74 warmup, 81 measured). The first violation is sequence 4, 2.365 ms; maximum
error is 4.183 ms. The rejection remains in force and every row is retained.

Preparation acec5a2 was reviewed with Prism using Anthropic Claude Sonnet 5.5.
The owner authorized screen use for remaining release tests. Total duration is
43.076 seconds, native 34.724 seconds, within 60/40-second limits. All nine
owned groups retire and are observed absent. Eighteen raw logs, exact bound
inputs, geometry/configuration and all 8321 SDK files verify unchanged. The
attempt is sealed; no unchanged retry or background native work remains.

Independent all-row accounting joins 248 requests/receipts to acquisitions,
drawable identities and callbacks, with all 258 attempts/1022 events retained.
All 252 scoped metrics samples join as well. Of all callbacks, 257 are positive
and one pre-observer callback has zero shown time. All 1987 display ticks remain.
Recorder scope closes, pending/acquiring zero. No failure-context observation
ran because no foreground check failed.

At receipt, presentation request and final GPU end, none of the 120 measured
inputs has an older recorded acquired drawable awaiting callback. This describes
the recorded lifecycle only; it does not prove Metal pool availability, native
object deallocation, compositor depth or a causal queue explanation.

The complete unfiltered measured population has descriptive receipt-to-shown
p95 37.362 ms and 30/120 observations above 33.3 ms. These are **unqualified**
statistics, not accepted phase calibration or release evidence. Measured wake
lateness p95 is 1.161 ms; foreground check p95 1.400 ms; post-check-to-receipt p95
0.535 ms. Percentiles must not be added. The same-row clocks show preparation
currently starts at the target, then performs synchronous foreground/snapshot
work before actual receipt. Next prepare those checks before the target and
keep the final native wait immediately before dispatch, preserving the same
actual receipt bound and all original route/identity/clock constraints. Do not
relax criteria, filter violations or compare this with an unrelated native run.

See [summary](result-summary.json), [all-row verifier](execution-sources/verify_result.py.txt)
and [indexed archive](evidence-index.json). `capture/REL` maps to local
`prepared/REL`; the verifier is a source snapshot from its parent, not a portable
standalone CLI. Hashes and committed-copy mappings can be checked from the
archive; full input revalidation requires the named external pinned workspace.
Internal verification and Prism cover this actual result before commit. No
feature, rendering/pacing default, SDK, API or 60 Hz/33.3 ms gate changes occur.
The first macOS release is still unqualified; representative responsiveness,
resource/endurance, physical/human, final SDK/rights and final audit remain open.
