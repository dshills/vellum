
## Share compact macOS typography and prose presentation

Prism: `mac-v1-shared-style.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage. Its low-confidence pixel/Dp concern is resolved by manualConfig fixing LogicalDPI=160 and providing all fonts explicitly; no ambient display scale enters this test. Real light/dark field and prose geometry passes, along with unchanged inset/focus and retained model tests. Portable race/vet and Notes offscreen race checks pass. Scope follows the owner-approved macOS-v1 candidate contract; no backend/performance/publication gate changes.

Focused checks: `mac-v1-style-acceptance-checks.json`.

## Add window-scoped native toolbars and command menus

Prism: `mac-v1-app-shell-final.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage. The initial real menu retry finding is fixed with one bounded same-state retry and an in-flight cancellation/join regression. Final findings adjudicated: error scopes compile and propagate owner/preparation errors separately; availability is captured during dirty reconciliation, not native paint callbacks; attachment is enqueued before any state flush on the same native main FIFO, and the attach acknowledgment invalidates on owner. Atomic clear-before-load publication is directly tested. Native window isolation, ABA refusal, unchanged customization/modal/teardown tests and Notes command history pass. No backend, gate or publication changes.

Focused checks: `mac-v1-shell-acceptance-checks.json`, `mac-v1-shell-review-fixes-checks.json`.

## Observe canceled frame measurements before FIFO dispatch

Prism: `mac-v1-cancellation.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage with no findings. Deterministic delayed-notification test preserves the identical admitted event and earlier terminals; actual grid/tree metrics race checks pass 20 repetitions. Foreground first-paint observation remains in the upcoming native checkpoint; this commit does not claim it.

Focused checks: `mac-v1-cancellation-checks.json`.

## Define private macOS v1 candidate support and delivery

Prism: `mac-v1-candidate-contract-final.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage. Preserved the historical personal_use contract in response to its initial finding. Final stale-reference concern is not a defect: the frozen personal-v0.1 catalogue and immutable historical records retain their source identities. A search of active Go consumers found the SDK contract version assertion, now updated; root metadata consistency and independent SDK tests pass. The independent consumer intentionally pins the candidate API version instead of reading repository files. Native evidence remains pending until actual observation; public audit and human qualification remain unchanged.

Focused checks: `mac-v1-delivery-contract-pass-checks.json`.

## Return native file-panel focus to its live parent

Prism: `mac-v1-dialog-focus-final.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage. Added exact main-window identity for nil-key completion and tested production identity/state combinations in response to its first review. Final timing concern is reserved for actual packaged native verification: parent or nil-key completion is accepted only with active/visible/nonminimized/exact-main guards. Scalar fixture is explicitly policy-only; it does not establish responder delivery or closed-window behavior. The live repro and follow-up outcomes are recorded separately. Focused race/vet checks pass; backend and release gate unchanged.

Focused checks: `mac-v1-dialog-focus-final-checks.json`.

## Record private macOS v1 candidate delivery and native acceptance

Prism: `mac-v1-final-acceptance.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage with no findings. Final native cases pass on source48c57e8: file-panel focus handoff, exact note history/cleanup, independent window toolbars, disabled refusal, gallery pointer targeting and saved-notebook relaunch. Initial real focus failure and long service-call interruption remain recorded separately. Matching SDK, external module origins, starter, exact signed app ZIPs and checksums verified. Raw local journal is private. Human IME/VoiceOver and other hardware remain unqualified; public release and existing responsiveness policy unchanged. This report-only commit does not relabel runtime bytes.

Focused checks: `mac-v1-dialog-focus-final-checks.json`, `mac-v1-cancellation-checks.json`, `mac-v1-delivery-contract-pass-checks.json`.

## Refresh README for the merged macOS v1 candidate

Prism: `mac-v1-readme.json` (Anthropic Claude Sonnet 5.5, complete coverage, no findings). Current source and historical public downloads are distinguished; Mac presentation, native shell, examples, verification and support limits are described. All35 local links, four entry points, three gallery flags and release-version consistency checked; receipt: `mac-v1-readme-checks.json`. Documentation only; existing runtime verification reused.
