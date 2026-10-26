
## Add controlled Mac checklist with inline names and category artwork

Prism: `mac-content-checklist.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism coverage is complete. Its per-keystroke validation concern is not reproduced: ordinary managed TextField Change commits on Return/blur, unlike the special live palette field. The native regression now clears a draft without an Invalid callback and commits a replacement name once. Names intentionally remain bounded nonblank single-line UTF-8. Native Space events include their real space rune; earlier malformed test events and fixture compilation failures remain local diagnostics. Root race/vet, checklist routing, gallery checks and unchanged v1 baseline comparison pass. No storage, clocks or physical input qualification is added.

Focused checks: `mac-content-checklist-ready-checks.json`.

## Add Mac file and folder picker field with native Browse and Reveal

Prism: `mac-content-path-picker.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism coverage is complete. Overlapping requests are now refused synchronously using the same completion Err-based active predicate as View; mode switches and late results are scoped to the captured kind. A regression proves overlap refusal and busy retirement without showing a chooser. The ticket-retention finding is bounded and not a stuck-busy defect: revoked completion Err automatically becomes nonnil and View/Request consistently derive availability from it. Native chooser recipes and real cancellation cleanup use a fake panel, with no OS presentation; no Finder action is executed by tests. Root/native/gallery/vet/API checks pass. Existing Services and accepted v1 archives remain unchanged.

Focused checks: `mac-content-path-picker-checks.json`, `mac-content-path-final-checks.json`, `mac-content-path-review-fixes-checks.json`.

## Add Mac date-range picker with shared calendar and civil-date presets

Prism: `mac-content-date-range.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Fixed civil-equivalent time representation resets and guarded callbacks against stale accepted values and calendar months. Hardened calendar decoration against short or unrelated node IDs. Regression tests cover these review findings; no physical pointer qualification is claimed.

Focused checks: `mac-content-date-range-checks.json`, `mac-content-range-final-checks.json`.

## Add embedded Mac PDF viewer with bounded native lifecycle

Prism: `mac-content-pdf-final.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Fixed Prism thumbnail constraint finding. Addressed the serialized-worker observation with two bounded jobs; two blocked OS reads/decoders can still delay loading, explicitly documented. The native run reproduced omitted optimized-wire metadata, repaired by canonical JSON fallback with copy/equality/schema regressions. Final actual-window check passed load/removal/recreation/removal with zero hosts/workers. This is automated native evidence, not physical pointer or human accessibility qualification.

Focused checks: `mac-content-pdf-ready-checks.json`, `mac-content-pdf-final-checks.json`.

## Add Mac media player with native transport and full-color posters

Prism: `mac-content-media-final.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Fixed the prior Prism smoke fixture branch finding; it constructs only the chosen widget. TiledImage is immutable and the model owner-only; a WithPatch replacement regression proves poster refresh without mutating prior artwork. Final Prism notes AVFoundation reopens a mutable path after initial admission: documented explicitly in the API and control contract, with application-owned snapshots required for stable bytes. This is not an untrusted-file sandbox or hard runtime size guarantee. Native source/observer retirement, raw clipping at 1x/2x and both axis directions, namespaced dock geometry, original-color posters, root API compatibility and focused native load/removal checks passed. Actual-window evidence is separate from unperformed physical playback/seek/volume and human accessibility checks.

Focused checks: `mac-content-media-ready-checks.json`, `mac-content-media-final-checks.json`, `mac-content-media-title-checks.json`.
