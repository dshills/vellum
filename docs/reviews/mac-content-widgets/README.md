
## Add controlled Mac checklist with inline names and category artwork

Prism: `mac-content-checklist.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism coverage is complete. Its per-keystroke validation concern is not reproduced: ordinary managed TextField Change commits on Return/blur, unlike the special live palette field. The native regression now clears a draft without an Invalid callback and commits a replacement name once. Names intentionally remain bounded nonblank single-line UTF-8. Native Space events include their real space rune; earlier malformed test events and fixture compilation failures remain local diagnostics. Root race/vet, checklist routing, gallery checks and unchanged v1 baseline comparison pass. No storage, clocks or physical input qualification is added.

Focused checks: `mac-content-checklist-ready-checks.json`.

## Add Mac file and folder picker field with native Browse and Reveal

Prism: `mac-content-path-picker.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism coverage is complete. Overlapping requests are now refused synchronously using the same completion Err-based active predicate as View; mode switches and late results are scoped to the captured kind. A regression proves overlap refusal and busy retirement without showing a chooser. The ticket-retention finding is bounded and not a stuck-busy defect: revoked completion Err automatically becomes nonnil and View/Request consistently derive availability from it. Native chooser recipes and real cancellation cleanup use a fake panel, with no OS presentation; no Finder action is executed by tests. Root/native/gallery/vet/API checks pass. Existing Services and accepted v1 archives remain unchanged.

Focused checks: `mac-content-path-picker-checks.json`, `mac-content-path-final-checks.json`, `mac-content-path-review-fixes-checks.json`.
