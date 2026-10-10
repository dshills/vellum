# Native accessibility acceptance

The standalone macOS accessibility bridge passes fresh portable/native race tests, vet, backend ownership regressions, two consecutive live Cocoa acceptance runs and an AX-enabled dynamic docking run. These are in-process native assertions; human VoiceOver and external AX client acceptance remain release gates. No VoiceOver, keyboard input source or operating-system preferences were changed.

Run on 2026-10-10, macOS 26.6.2 (25G83), arm64, Go 1.27.2. Final checks used the independent `/private/tmp/vellum-phase2-final-isolated` checkout with driver dependencies through patch 0018. Phase-four patch 0019 was excluded. The candidate uses synthetic Git baseline `5a7667995982006ea5799c3b78cb44e0d24fbe26`, prepared from canonical phase-one commit `89cb930`. [Source manifest](source-manifest.json) records SHA256 hashes of driver/portable inputs, preparation scripts/tests, dependency patches/manifests and prepared backend regression files. All final passing logs reference SHA256 `c8ac680cb1f580af127633934a0e46b5c1e6363ab1a988042eaffbf336e80e58`; the [native input manifest](native-source-manifest.json) preserves the same tested snapshot. Each log identifies the checkout, baseline, exact command, source-manifest hash and exit status. Race tests use `-count=1`; vet records its successful exit even with no diagnostics.

## Native selector evidence

The fixture walks the actual `NSWindow.contentView.accessibilityChildren` tree and reads roles, labels, values, enabled/selected state, identifiers and frames. It reads focus through `NSWindow.accessibilityFocusedUIElement`, verifies the native object remains an `NSWindow` with `AXWindow` role, and calls content-view `accessibilityHitTest` at published element centers.

Actual `accessibilityPerformPress`, `setAccessibilityValue:`, `setAccessibilityFocused:` and `setAccessibilityExpanded:` methods admit queued owner actions. The fixture separately verifies accepted model state; admission alone is not the acceptance assertion. It reads `isAccessibilityExpanded` after disclosure and verifies selection remains independent. Equal-frame selector parent/option probes resolve repeatedly to the later retained option. Hit testing visits descendants before their parent and later siblings first, traversing zero-frame groups. Tab and Escape enter the mutex-protected native system event queue and run through the actual window event loop.

Both live runs verify managed roles/values, disabled refusal, controlled button/text/numeric changes, numeric clamp, native focus and Tab traversal, selected-tab visibility, a 100000-row table with four materialized rows and stable selection/paging. Outline checks verify real pager geometry/hit testing, paging/recycled identity, expansion exposing a child without selecting its parent, press selecting without collapsing, collapse preserving selection, and leaf disclosure refusal. Modal checks verify isolation, nonzero geometry, focus, Tab wrapping and Escape dismissal.

A dynamically opened workbench window verifies duplicate local IDs in distinct panel namespaces, activation, deferred focus and Tab traversal, an initially-open dock modal, modal geometry/hit testing/isolation, owner wake for edits, accepted values, Escape and callback refusal after native retirement. The separate integrated docking fixture exercises floating, cross-window redocking, native-close return, shared document state, focus transfer and persistence with this accessibility bridge enabled.

## Commands and logs

Native commands source `drivers/cogent/env.sh` with `VELLUM_CACHE_DIR=/tmp/vellum-p0-cache`. The commands below ran in the isolated checkout, not the canonical phase-four build.

| Check | Command | Evidence |
| --- | --- | --- |
| Dependency preparation contracts | `cd drivers/cogent && python3 prepare_test.py` | [PASS, 11 tests](preparation-tests.txt) |
| Mandatory numeric bounds | `go run /private/tmp/vellum-ax-numeric-contract.go` with [recorded probe source](numeric-contract-source.txt) | [PASS](numeric-contract.txt) |
| Portable focus race | `go test -race -count=1 ./ui` | [PASS](portable-race.txt) |
| All native driver packages | `cd drivers/cogent && go test -race -count=1 ./...` | [PASS](native-race.txt) |
| Modal initial focus | `go test -race -count=1 -run TestAccessibilityDialogInitialFocusBeforeDeferredShow -v .` in the native driver | [PASS](modal-start-focus.txt) |
| Native retained-tree hit regressions | `go test -race -count=1 -run TestNativeAccessibilityHitTestRetainedTreeOrder -v .` in the native driver | [PASS](native-hit-regression.txt) |
| Native vet | `cd drivers/cogent && go vet ./...` | [PASS, exit 0](native-vet.txt) |
| Backend screen ownership/polling | `go test -race -count=1 ./system/driver/desktop -run TestVellumScreen` in prepared upstream | [PASS](backend-screen-race.txt) |
| Current-window publication/targeting | `go test -race -count=1 ./core -run TestVellumCurrentWindowSnapshotDuringPublication` in prepared upstream | [PASS](backend-current-window-race.txt) |
| Race-built AX fixture | `go build -race -o /private/tmp/vellum-accessibility-smoke-phase2-final ./cmd/accessibility-smoke` | [PASS](native-build.txt) |
| Native AX acceptance, run 1 | `/private/tmp/vellum-accessibility-smoke-phase2-final` | [PASS, exit 0](native-ax-1.txt) |
| Native AX acceptance, run 2 | `/private/tmp/vellum-accessibility-smoke-phase2-final` | [PASS, exit 0](native-ax-2.txt) |
| Race-built integrated docking fixture | `go build -race -o /private/tmp/vellum-accessibility-docking-smoke-phase2-final ./cmd/docking-smoke` | [PASS](docking-build.txt) |
| First-Escape event/focus trace | `env VELLUM_AX_EVENT_TRACE=1 /private/tmp/vellum-accessibility-smoke-phase2-final` | [PASS, exit 0](modal-event-trace.txt) |
| AX-enabled integrated docking | `/private/tmp/vellum-accessibility-docking-smoke-phase2-final` | [PASS, exit 0](native-docking.txt) |

Builds emit the existing duplicate `-lobjc` linker warning. Final live runs emit no race report.

## Regressions and review

The phase-two preparation-test snapshot asserts patches 001..0018 in order and verifies pinned input hashes, rollback preservation and failure handling. Canonical phase-four inputs additionally contain patch 0019; [phase input snapshots](phase-input-snapshots.json) explicitly record the driver, dependency manifest and preparation-test differences. Numeric bounds are mandatory for both NumberField and Slider; zero/omitted bounds and zero Step are refused before native publication. Native registry lookup is performed only after a stable live AX probe; the registry mutex alone does not protect Name/SetName during creation.

Focused tests cover bounded/stale/closed admission, replacement callbacks without unnecessary generation changes, finite numeric validation, modal/tab scopes, disabled ancestors, dock namespaces, typed-nil/retired nodes and bounded virtual identity. Requests cancel on attachment retirement and time out after five seconds with Background contexts without leaking registrations. Native payload tests reject malformed later nodes, duplicate identifiers, missing/cyclic parents and invalid metadata before AppKit elements mutate. Native hit fixtures check equal parent/child and sibling frames, reordered siblings, zero-frame ancestors, later-root overlay, non-containing later siblings and misses with 32 repeated queries per case. Namespaced focus preserves advertised permissions and original generation validation.

Table paging tests reach the trailing row of a six-row/four-slot viewport and reject stale counts. Outline tests separate selection/disclosure callbacks, check current expansion/child state and refuse replaced leaves or offscreen rows. Live startup reproduced geometry arriving before semantic elements under demand rendering; independent geometry caching now covers that ordering. Identical semantic refresh no longer retires admitted focus unnecessarily, and delivery resolves the latest callback.

An [earlier live failure](initial-modal-escape-failure.txt) left an initially-open dock modal visible after its first queued Escape. Source inspection identified a readiness gap: Scene.OnFirst keyboard handlers require an event-focus ancestor chain, while default dialog focus can wait for deferred Shown. Modal binding now seeds explicit start focus before Run, skips nested modal Scenes and supplies a body anchor when no managed target exists. Focused tests pass, followed by two normal live runs and one traced run, each sending initial Escape once with no prior focus action or retry. Trace lines 276–299 show focus, Escape through modal ancestry, close and focus retirement. This fixes the identified startup contract; the passing assertions remain an in-process acceptance scope.

Multiwindow checks reproduced backend screen-cache/polling and current-window publication races. Patch 0018 provides per-app poll admission, copied screen/window snapshots, synchronized accessors/DPI updates and current-window reads under the existing publication mutex. Fresh backend regressions and integrated docking pass against the repaired phase-two dependency.

Prism Claude Sonnet 5.5 findings were fixed or checked against concrete cross-file ownership rules. [Resolution facts](prism-resolution-facts.json) record the evidence; the final review disposition is tracked with the phase commit.

## Remaining gates

Human VoiceOver announcements/navigation, external-process AX queries/actions, Windows UI Automation, Linux AT-SPI, native detailed text ranges and multiline editor accessibility remain unverified. No physical pointer, external keyboard injection or screen-reader certification is claimed. Custom extensions require their own semantic adapter; offscreen rows are reached through paging. Multiple-monitor/DPI and GPU reset acceptance are qualified by the rendering report.
