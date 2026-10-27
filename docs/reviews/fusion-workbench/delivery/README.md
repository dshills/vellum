# Forge private Mac delivery — build 3

The original Fusion-inspired workbench and reusable public additions are complete
for the bounded demonstration in [the spec](../../../../specs/fusion-workbench.md).
The owner requested replacement of the initial icon buttons; the final app uses
distinct original antialiased full-color silhouettes, transparent backgrounds and
flat tool buttons. Published releases, backend/pacing and release gates are unchanged.

Final app source: `b41f08fc5bcc97eb5df8a43005fbc23f1675da3f`. The private producer ran from that clean
committed checkout. Archive: `Vellum-Forge-macOS-arm64.zip`; SHA256 `1749fcc2d2f4321efb3da38ffde317c2cc562325f9e52c42ae5a1f0c7abf730c`.
[Manifest](package-manifest.json) and [checksum](SHA256SUMS) bind the exact ZIP.
The producer verified the signature after exact ZIP extraction. The exact extracted
bundle was launched, its new ribbon icon created one body, and native Undo restored
the original seven-body study. The app remains running from the nonsynchronized
private build3 extraction. Durable ZIP/README/checksum/manifest are in the owner's
local `deliveries/forge-20261027-build3` folder outside the repository.

## Verification

Focused race tests, native routing/geometry/overflow checks at160/240/320 DPI,
vet and the unchanged v1 API baseline check passed; checkpoint directories record
exact commands and original Prism findings/dispositions. All implementation and
visual checkpoints used Prism with Anthropic Claude Sonnet5.5 and were committed.
No broad performance/qualification campaign was started.

The [bounded native smoke](phase6-reviewed-smoke.stdout) sampled62 colors from a
779x502 real GPU readback; subsequent mesh growth advanced frames1->2 and tracked
buffer bytes48000->48816. Orbit/pan/zoom/history in that smoke are synthetic owner
callbacks. They are explicitly separate from physical native evidence.

[Native interaction records](native-checks.json) cover real pointer orbit, Shift-pan,
wheel zoom with cleared modifiers, cube face picking, Escape cancellation during a
drag, nearest visible body selection, browser selection/visibility, body deletion,
Add/Undo/Redo/Fit, feature earlier reorder and marker visibility with Undo. The
known source comparisons permit reuse after the artwork-only correction; the
final build3 ribbon icon and native Undo were physically rechecked. CUA native
keyboard verified quantity Apply, Undo/Redo, action-panel Escape and viewport focus
restoration. Initial tool/helper control failures are explicitly excluded in the
record rather than presented as application passes. No user notebook was changed.

The [raw final app capture](../../../screenshots/forge.jpg) was manually inspected;
[capture metadata](../../../screenshots/forge-capture.json) records source and bytes.
It is an unmodified running-app image, including macOS capture/pointer indicators.

## Limits

Opaque box proxies and a48-body demo cap, no CAD kernel/constraints/B-rep operations,
no Fusion-scale claim. Feature history reorders creation/visibility presentation;
it does not replay dimensions or solve dependencies. The directional action panel
has visible buttons, not invisible flick/nested radial tracking or pointer anchoring.
Documents are in memory; closing discards them. Primary capture is same-window;
alternate buttons do not capture outside the view. Human VoiceOver/IME, other
hardware, downloaded-quarantine launch and performance qualification are unchanged.
The private app is ad-hoc signed and unnotarized. Nothing was published or tagged.
