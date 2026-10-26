
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

## Add reusable Mac file drop zone with shared chooser validation

Prism: `mac-v11-drop-zone.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism coverage is complete. Chooser cancellation now skips admission and preserves existing feedback, with regression tests. Its UTF-8 growth concern is not reproduced: replacement is the one-byte ASCII question mark and cannot grow a truncated byte buffer; split/invalid UTF-8 and NUL vectors now verify the 512-byte valid-text bound. Root race/vet, existing transformed native file-drop routing, rejected-drop consumption/ancestor protection, gallery checks and stable API comparison pass. The component reads no files and retains no paths. Finder pre-drop acceptance remains a documented existing backend limit; actual physical Finder behavior is not newly verified.

Focused checks: `mac-v11-drop-zone-checks.json`, `mac-v11-drop-review-fixes-checks.json`.
