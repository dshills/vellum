
## Share compact macOS typography and prose presentation

Prism: `mac-v1-shared-style.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage. Its low-confidence pixel/Dp concern is resolved by manualConfig fixing LogicalDPI=160 and providing all fonts explicitly; no ambient display scale enters this test. Real light/dark field and prose geometry passes, along with unchanged inset/focus and retained model tests. Portable race/vet and Notes offscreen race checks pass. Scope follows the owner-approved macOS-v1 candidate contract; no backend/performance/publication gate changes.

Focused checks: `mac-v1-style-acceptance-checks.json`.
