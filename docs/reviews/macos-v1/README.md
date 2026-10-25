
## Share compact macOS typography and prose presentation

Prism: `mac-v1-shared-style.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage. Its low-confidence pixel/Dp concern is resolved by manualConfig fixing LogicalDPI=160 and providing all fonts explicitly; no ambient display scale enters this test. Real light/dark field and prose geometry passes, along with unchanged inset/focus and retained model tests. Portable race/vet and Notes offscreen race checks pass. Scope follows the owner-approved macOS-v1 candidate contract; no backend/performance/publication gate changes.

Focused checks: `mac-v1-style-acceptance-checks.json`.

## Add window-scoped native toolbars and command menus

Prism: `mac-v1-app-shell-final.json` (Anthropic Claude Sonnet 5.5, complete coverage).

Prism complete coverage. The initial real menu retry finding is fixed with one bounded same-state retry and an in-flight cancellation/join regression. Final findings adjudicated: error scopes compile and propagate owner/preparation errors separately; availability is captured during dirty reconciliation, not native paint callbacks; attachment is enqueued before any state flush on the same native main FIFO, and the attach acknowledgment invalidates on owner. Atomic clear-before-load publication is directly tested. Native window isolation, ABA refusal, unchanged customization/modal/teardown tests and Notes command history pass. No backend, gate or publication changes.

Focused checks: `mac-v1-shell-acceptance-checks.json`, `mac-v1-shell-review-fixes-checks.json`.
