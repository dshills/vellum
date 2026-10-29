# Consumer closeout — 2026-10-29

Feature-frozen maintenance and developer handoff; no release gate/backend change.
All checkpoints use internal review under the owner's current instruction.
Historical Prism reports remain historical and were not rerun or transmitted.

| Phase | Result | Record |
| --- | --- | --- |
| 1 | Native starter Cancel/Redo, clean shutdown and saved relaunch passed | [Lifecycle](01-lifecycle.md) |
| 2 | Exact previously pushed main e932a7b green across all three workflows after one runner-capacity retry | [CI](02-ci.md) |
| 3 | Added offscreen hosted SDK consumer job; full local command passed | [SDK CI](03-sdk-ci.md) |
| 4 | Generated document apps use stable app-owned default private storage | [First launch](04-first-launch.md) |
| 5 | Updated guides, review policy and private build20 packaging instructions | [Handoff review](05-handoff.md) |

The native lifecycle observations supersede the incomplete check in the older
consumer-private-20261028 delivery. Its five implementation commits were already
pushed in main e932a7b. These closeout commits and the newly added hosted SDK job
remain local until a later push; prior green CI does not certify later commits.

Generate from the [application guide](../../building-macos-apps.md#complete-document-starter),
verify and package using the [private bundle recipe](../../private-consumer-bundle.md).
Final archive identities and live observations belong to the delivery's own
BUNDLE_MANIFEST.json, SHA256SUMS and verification.json. Never present a failed
producer attempt or intermediate build directory as the finished payload.

Native AX menu/window checks are distinct from physical pointer verification.
Human IME/VoiceOver and other hosts remain unqualified. Apps remain private,
ad-hoc signed and unnotarized; no new public release is authorized here.
