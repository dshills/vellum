# macOS v1 for personal use

The current source release is **v1.1.3**. The v1.0.0 stable API baseline and
recorded Apple Silicon macOS 26.6.2 qualification are historical anchors, not the
version of every installed app. The private build25 candidate requires macOS 27.0
and has scoped Documents checks on 27.2. That evidence does not qualify every
app, OS or host. See [current source/artifact status](status.md).

The [published build13 SDK and six apps](https://github.com/dshills/vellum/releases/tag/macos-v1.0.0-build13)
remain the prepared public downloads. They use source 64bf9cf and SDK v1.0.0,
while later root source tags and private builds retain distinct identities.
They are ad-hoc signed, unnotarized developer downloads. The
[release closeout](../specs/macos-v1-release.md) preserves the scope of the owner
choices; it does not authorize every future native payload or signing route.

## Supported API contract

| Area | v1 contract and limits |
| --- | --- |
| Application/lifecycle | `app.Application`, stable IDs and exact `WindowRef` lifetimes; owner-thread model changes, cancellable work and multiple native windows. Native main loop runs once per process. |
| Commands/history/documents | Public `command`, `history`, `document`, `editor` APIs; application-owned validation, history and persistence. Notes remains plain text/local storage with bounded recovery. |
| Mac appearance | Public `ui.MacTheme`, `MacTypography`, `MacProseStyle`; compact field insets, proportional prose and emoji fallback. Managed controls approximate macOS; AppKit owns titlebar toolbars and native services. Original-color toolbar PNG artwork is preserved. |
| Native shell | `Driver.TitlebarToolbars` factory binds each exact live window; immutable catalogue of up to16 actions, up to32 concurrent attachments. Legacy singular fields remain supported. Availability is owner-captured and rechecked on execution. Customization persists using explicit stable identifiers. |
| Menus/shortcuts | `Driver.CommandMenus` uses process-wide `Driver.Commands`. Independent document registries require application active-document routing. Native menu key equivalents cover single-character window shortcuts; focus-first text shortcuts stay managed. |
| Controls and collections | Existing frozen [catalogue](../release/personal-macos-v01.json); [control contracts](controls.md) document per-widget bounds, validation and keyboard alternatives. Virtual data views retain bounded/caller-owned data. |
| Canvas/placement | Existing public APIs retained; palette placement stays same-window, with application-owned command/history. Cross-window drag and arbitrary scene workloads are not qualified. |
| Experimental tools | Replay, semantic automation, diagnostics, timing probes, custom GPU paths and professional workbench qualification remain experimental. These exports remain outside the stable API promise. |

The named stable APIs preserve exported signatures and documented defaults throughout v1. Read the [explicit stable root/driver scope and experimental exclusions](../specs/stable-api.md) before relying on that guarantee. No API removals or import-path changes are part of this closeout. The Cogent/WebGPU backend is unchanged and experimental rendering/pacing remains disabled. Existing measured responsiveness policy is unchanged; no numeric optimization target was introduced as a new gate.

## Verification boundary

Portable race/vet, affected native offscreen tests, command/lifetime regressions and independently extracted SDK consumers provide automated evidence. Physical pointer/native-window observations are recorded separately in the [RC native report](reports/macos-v1/README.md). That report lists exact source identity, observed host and outcomes. Historical soak and catalogue evidence is reused for unchanged paths; it does not certify all future workloads.

Human IME candidate behavior and VoiceOver usability remain explicitly
**unqualified**, as chosen by the owner. Use the [human checklist](personal-macos-input-checklist.md)
before claiming those results. Other macOS versions, Intel, additional displays,
Windows and Linux remain unqualified; Wayland is unsupported.

The owner approved the exact [build13 publication](../release/macos-v1-build13-publication.json).
The [download receipt](reports/macos-v1-build13-publication/README.md) records
all eight actual public downloads, unchanged SDK/app hashes, the reviewed release
body, matching source tag, original notices and checksum filename adaptation.
The [build13 tests](reports/macos-v1-build13/README.md) retain their bounded native
render/close/exposure results and intermediate observations. Producer-time
unverified/private flags and older reports retain their historical meaning;
they are not blanket authorization to publish a future payload. Apps remain
ad-hoc signed and unnotarized; downloaded-quarantine launch is unqualified.

## Consume the published SDK

Download the prepared SDK and SHA256SUMS from the
[build13 release](https://github.com/dshills/vellum/releases/tag/macos-v1.0.0-build13).
Compare the SDK's digest against its named checksum row before extraction. If
you downloaded all seven payloads, `shasum -a 256 -c SHA256SUMS` verifies the set.
For a subset, compare only the downloaded files' named rows.

Extract the internal `vellum-v1.0.0` folder intact. Use Go1.25.6, Python3 and
Xcode Command Line Tools on Apple Silicon. The SDK already includes prepared
sources and the pinned native capsule; do not run prepare.sh over it. From the
SDK, `python3 scripts/new-app.py /absolute/new/app --module example.com/my-app`
generates an independent starter; build/run from that new app. Its replacements
must resolve into the extracted SDK. The included examples are separate modules.
See [app-shell APIs](macos-app-shell.md), [shared presentation](mac-presentation.md)
and [Notes storage/commands](../examples/macapps/notes/README.md).

Native apps use Archive Utility to extract their ZIPs. Notes keeps its existing
local plain-text notebook and recovery/history behavior; validation used only a
disposable notebook. Preserve backups of important personal data. The showcases
use original offline fixture data and reset on launch. Human/host qualification
limits above remain unchanged. Future payloads, native-module version tags,
Developer ID signing and notarization require separate owner decisions.
