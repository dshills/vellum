
## Add Mac-style tooltip help with keyboard and lifecycle routing

Prism: `mac-v11-help-tags.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism coverage is complete. Its lifecycle concern prompted clearer nil-check ordering and stronger tests that disable/remove actual open help tags. Record removal paths in reconcile.go (132/215), failed mounts and native window retirement all call disposeControl or disposeControlTree; no skipped removal path was identified. Root race/vet, native F1/Escape/focus/disable/removal routing, native vet, gallery compilation and unchanged stable API baseline comparison pass. Hover uses the existing delayed event source; this is offscreen evidence and does not claim a physical hover run. The v1 build9 archives remain unchanged.

Focused checks: `mac-v11-help-checks.json`, `mac-v11-help-final-checks.json`, `mac-v11-help-lifecycle-checks.json`.

## Add Mac link control and optional native URL opening

Prism: `mac-v11-links.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism coverage is complete. Its low-confidence dual-parser concern is addressed with shared Go/native HTTP(S), IPv6, encoded-path and rejection vectors. Both validation gates are required; native rejection remains authoritative and no guarantee of universal parser equivalence is made. Existing Services implementations are unchanged. Root race/vet, native recipe tests without browser opening, gallery tests/vet and stable API comparison pass. Contextual Open and Copy revalidate live identity, URL, disabled state and callback availability. Browser/clipboard actions were not invoked by tests; physical link targeting is not newly qualified.

Focused checks: `mac-v11-links-checks.json`, `mac-v11-links-parsers-checks.json`.
