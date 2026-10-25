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
| 3 Input/accessibility | Pending | Focused tests plus bounded native session; human IME/VoiceOver separate |
| 4 Integration | Pending | Stable-candidate suite and bounded native resource/lifecycle checks |
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
