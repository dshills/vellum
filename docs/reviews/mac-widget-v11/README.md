
## Add Mac-style tooltip help with keyboard and lifecycle routing

Prism: `mac-v11-help-tags.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism coverage is complete. Its lifecycle concern prompted clearer nil-check ordering and stronger tests that disable/remove actual open help tags. Record removal paths in reconcile.go (132/215), failed mounts and native window retirement all call disposeControl or disposeControlTree; no skipped removal path was identified. Root race/vet, native F1/Escape/focus/disable/removal routing, native vet, gallery compilation and unchanged stable API baseline comparison pass. Hover uses the existing delayed event source; this is offscreen evidence and does not claim a physical hover run. The v1 build9 archives remain unchanged.

Focused checks: `mac-v11-help-checks.json`, `mac-v11-help-final-checks.json`, `mac-v11-help-lifecycle-checks.json`.

## Add Mac link control and optional native URL opening

Prism: `mac-v11-links.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism coverage is complete. Its low-confidence dual-parser concern is addressed with shared Go/native HTTP(S), IPv6, encoded-path and rejection vectors. Both validation gates are required; native rejection remains authoritative and no guarantee of universal parser equivalence is made. Existing Services implementations are unchanged. Root race/vet, native recipe tests without browser opening, gallery tests/vet and stable API comparison pass. Contextual Open and Copy revalidate live identity, URL, disabled state and callback availability. Browser/clipboard actions were not invoked by tests; physical link targeting is not newly qualified.

Focused checks: `mac-v11-links-checks.json`, `mac-v11-links-parsers-checks.json`.

## Add native Mac sharing with bounded payloads and exact-window cleanup

Prism: `mac-v11-share-final.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Initial Prism findings addressed: regular-file validation is explicitly a worker-side ShareServices obligation and implemented in the supported driver; metadata validation performs no I/O. Gallery busy state derives from Completion lifetime, with revoked-window regression. Final findings: added nil gallery-binding guard; documented synchronous filesystem syscall cancellation limit rather than create detached unbounded workers; shutdown already marks finished synchronously, and expanded real-cleanup fixtures prove late finish is idempotent for selected/unselected sessions. Cleanup never closes committed service UI. All focused checks pass; initial icon catalogue failures remain diagnostic artifacts and the exact expected catalogue increases by one. Native recipes/cleanup tests use synthetic content/fake picker, with no OS presentation or transfer. Physical Share anchoring/service delivery remains unverified. Stable API comparison passes and accepted v1 archives remain unchanged.

Focused checks: `mac-v11-share-verified-checks.json`, `mac-v11-share-lifetime-checks.json`, `mac-v11-share-review-fixes-checks.json`.
