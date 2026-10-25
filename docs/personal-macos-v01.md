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
| 2 Mac behavior | Complete; focused race/owner/theme checks and native initial delivery pass | Notification-driven system preferences, reduced-motion insertion, compatible TitlebarToolbar entry point |
| 3 Input/accessibility | Complete within named automated/native scope; human usability explicitly unqualified | [Input acceptance script](personal-macos-input-checklist.md), public Unicode/composition/history consumer, existing native AX/FIFO regressions |
| 4 Integration | Complete within bounded scope: offscreen suite, 10 native cycles, four-state soak and 10-minute activity pass | Existing cycle fixture now optionally exercises system preferences and requires zero terminal native observers; one full offscreen suite and bounded native checks |
| 5 Delivery | Private delivery ready; exact light ZIP launched, native interactions checked, SDK native relocation passes | Matching private SDK/app, external consumer and final extracted bytes |

## Honest support boundary

Managed controls use Mac-style rendering with native windowing, text and
accessibility integration. AppKit owns the native titlebar toolbar, Quick Look,
Colors, Fonts and alert services. Pixel-identical AppKit controls, blanket human
accessibility, production performance, every display/OS and secure-event-input
protection are not claimed. Same-window placement, bounded collections and
caller-owned gradient/range history retain their documented limits.

Human-only input/accessibility cases and unsupported hardware need explicit
results. They are not inferred from successful semantic or routed event tests.
The final package is ready for the named private personal-use scope, not a new public release.

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

## Candidate identity and bounded results

The delivered runtime/SDK build checkpoint is
`a693e711821d1829d44d057bed6482ab04c1f6ea`. Subsequent acceptance documentation
commits do not relabel these built bytes. The SDK SHA-256 is
`46d8bca899a8b6f543052bd101c79ef582271e174fa68d1176be82ef381ce160`.
Both app variants are built from that exact extracted SDK and selected native
archive, then independently ad-hoc signed. Clean ZIP extractions verify with
`codesign --verify --deep --strict`.

All 789 native root test methods passed in five bounded groups after an original
monolithic 240-second package timeout. The original invocation remains a timeout;
no assertions were weakened. Other native packages, portable race/vet, placement,
SDK consumer, native vet and focused fixture repair checks pass.

The first final-SDK cycle attempt failed because the diagnostic's simple survivor
created a new canvas extension wrapper on each View. Preference delivery caused a
refresh and correctly refused a second mount of its still-live GPUCanvas. The
fixture now retains one wrapper per opening; the repair has focused race/vet
coverage for both simple and composed panel views, Prism review and its own
commit. This diagnostic failure is retained in the evidence.

The corrected 10-cycle run passes all composed native actions, nine same-name
replacement checks, stale-reference refusal and terminal resource retirement.
Preference delivery reached the owner; terminal preference registration count,
pending callbacks, native outstanding operations and managed AX windows are zero.
The bounded post-GC sample endpoints were 20.7/68.5 MB Go heap and 151.8/415.8 MB
RSS (decimal); final samples were nearly unchanged. This is not a native/GPU heap
leak proof or long-duration qualification.

The four-state soak passes actual visible/occluded/minimized/background state
verification and terminal cleanup. Each phase has one eligible 60-second sample:
0.022–0.025 percent of one CPU and zero primary paint wakes. A background worker
update committed. These are bounded observations, not a new performance gate.

The ten-minute composed workbench activity passes: 2,400 steps, 600 native edits,
200 resizes and 1,800 workers (600 committed, 1,200 stale results discarded), with
zero active workers at completion. The native temporary-document save/open and
unsaved-close fixture also passes. This activity fixture does not establish
terminal native/GPU allocation attribution.

The external native editor AX, native text/modal FIFO and public Mac starter
launch/close pass. A fresh external module extracted from the delivered SDK
passes origin/module verification, race/vet/build and native launch/GPU-ready
report. The exact light app renders with original-color icons; physical Quick
Look open/close/focus return and range/gradient drag/key checks pass. See the
[final acceptance record](reports/personal-macos-v01/README.md) for scope and
remaining unqualified checks. Find shortcut requests were posted at the end of
the desktop window but their final AX state was not observed; that final physical
case is not marked passed. Separate Dark-bundle launch was not repeated.

The owner explicitly chose to leave human real-provider IME and VoiceOver
unqualified for this personal v0.1. Physical scale/display changes and other
hardware also remain unqualified. No global settings were changed.

The terminal census also lists one invisible, untitled AppKit TUINSWindow record;
its attribution/deallocation is unqualified. Zero managed AX/tracked-operation
counts are distinct from zero AppKit allocations. The activity fixture reports
from inside an owner callback (`Dispatching: 1`), before native shutdown; its zero
worker count is the completed activity assertion. See the acceptance record for
these exact inventories rather than inferring a full native heap qualification.
