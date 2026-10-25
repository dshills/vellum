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

## Private SDK and app instructions

The source/API identity is `v0.1.0-personal.1`. The published alpha.3 tag/assets
remain historical. New public distribution is blocked (`third_party_distribution_audit`
is pending for this changed payload); no old exact-payload approval is reused.
The existing native archive and original notices remain hash-bound inputs.

Extract the SDK somewhere new, outside this checkout. It contains the prepared
native sources and pinned ARM64 static library; do not run prepare.sh over that
capsule. Source `drivers/cogent/env.sh`, then run the starter from
`examples/personal-mac`. Alternatively, from the extracted SDK:

```sh
python3 scripts/new-app.py /absolute/new/my-app --module example.com/my-app
cd /absolute/new/my-app
go test ./...
go run .
```

The new module's replacements must point inside the extracted SDK, never back to
the original checkout. `scripts/check-sdk.py` verifies archive contents and this
relocation/build route; `--runtime` additionally launches/closes its own native
consumer and belongs in the authorized desktop window.

Mac Widgets offers the frozen catalogue. Its light bundle follows system
preferences by default; Components/System restores following after a local
Light/Dark/accent override. The dark bundle is an explicit demonstration override.
Apps are ad-hoc signed and unnotarized. ZIPs made from clean temporary staging
avoid file-provider metadata being added to loose .app directories in Documents.
Verify signatures after extraction of the actual delivered ZIP.
