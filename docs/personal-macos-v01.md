# Personal-use macOS v0.1

This is the delivery ledger for the [frozen finish contract](../specs/personal-macos-v01.md).
It replaces repeated widget expansion with five finite integration checkpoints.
The [frozen catalogue](../release/personal-macos-v01.json) identifies existing
controls; [control APIs](controls.md) and [gallery instructions](mac-widget-demo.md)
remain the usage references. Public alpha.3 assets are historical and do not
contain the current widget work.

| Phase | Status | Evidence |
| --- | --- | --- |
| 1 Scope | Contract and catalogue frozen | This contract; baseline 929541a |
| 2 Mac behavior | Implemented; focused race/owner/theme checks pass, native initial sample pending | Notification-driven system preferences, reduced-motion insertion, compatible TitlebarToolbar entry point |
| 3 Input/accessibility | Automated renewal prepared; native/human outcomes pending | [Input acceptance script](personal-macos-input-checklist.md), public Unicode/composition/history consumer, existing native AX/FIFO regressions |
| 4 Integration | Prepared; final native outcome pending | Existing cycle fixture now optionally exercises system preferences and requires zero terminal native observers; one full offscreen suite and bounded native checks |
| 5 Delivery | Pending | Matching private SDK/app, external consumer and final extracted bytes |

## Honest support boundary

Managed controls use Mac-style rendering with native windowing, text and
accessibility integration. AppKit owns the native titlebar toolbar, Quick Look,
Colors, Fonts and alert services. Pixel-identical AppKit controls, blanket human
accessibility, production performance, every display/OS and secure-event-input
protection are not claimed. Same-window placement, bounded collections and
caller-owned gradient/range history retain their documented limits.

Human-only input/accessibility cases and unsupported hardware need explicit
results. They are not inferred from successful semantic or routed event tests.
The final package is a private personal-use candidate, not a new public release.

## Bounded final native script

Use the existing native cycle runner with `-system-appearance -composed
-same-name -cycles=10 -sample-every=1 -timeout=4m`, then the existing workbench
`-endurance=10m` activity mode and document-safety fixture. The personal scope
uses these bounded checks, not renewed 1,000-cycle/eight-hour production
certification. Record actual observer delivery and terminal count, window/service
retirement, source identity and any sampled memory growth. An unexplained failure
stops qualification of that scope; do not dilute its assertion to obtain a pass.

Within the owner-approved 30-minute desktop window, also check the final gallery's
system observation, explicit theme override, new widget keys/drag, native panels,
window minimize/restore and resize. Use the current display configuration only;
physical scale/display changes and human IME/VoiceOver stay unqualified unless
separately performed. Preserve global settings and normal clipboard contents.

The native observer census counts each retained notification receipt and KVO
registration, rather than just the observer object. Retirement attempts each
unregistration independently, retains failed registrations for terminal failure
and retry, and folds cleanup errors into existing service-shutdown status.
