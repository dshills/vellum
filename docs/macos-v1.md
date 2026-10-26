# macOS v1 for personal use

`v1.0.0` adopts the [stable API contract](../specs/stable-api.md) for private personal use on the recorded Apple Silicon macOS26.6.2 configuration. The [release closeout](../specs/macos-v1-release.md) governs final CI and matching SDK/apps. The completed RC, public alpha.3 and older personal-v0.1 deliveries retain their original identities. The later owner-authorized [v1.0.0 source release](../release/notes/v1.0.0.md)
adds a public root tag and checkout source; current native SDK/app packages remain
private with their original identities. No native binaries accompany that tag.

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

Human IME candidate behavior and VoiceOver usability remain explicitly **unqualified**, as chosen by the owner. Follow the [human checklist](personal-macos-input-checklist.md) before claiming those results. Other macOS versions, Intel, multi-display changes, Windows and Linux remain unqualified; Wayland remains unsupported. No signing identity, notarization or public distribution approval is implied. Current payload distribution audit remains pending.

## Consume the private SDK

The matching delivery includes a private SDK archive, original-color Mac Widgets catalogue, Notes, a two-window shell fixture and a separate example-source archive. Verify SHA256SUMS before extraction. Extract the SDK and example source beside each other so the example's module replacements point into `vellum-v1.0.0`, never the repository checkout. Use Go1.25.6 and Xcode command-line development tools on Apple Silicon. The SDK contains the pinned native capsule; do not run prepare.sh over it.

From the example source: `go test -race ./...`, `go vet ./...`, then `go run ./cmd/notes` or `go run ./cmd/macshell`. From the extracted SDK, `python3 scripts/new-app.py /absolute/new/app --module example.com/my-app` generates an independent starter. See [app-shell APIs](macos-app-shell.md), [shared presentation](mac-presentation.md) and [Notes storage/commands](../examples/macapps/notes/README.md).

Apps are ad-hoc signed and unnotarized. The delivery manifest pins source commit, SDK/native hashes and exact ZIP bytes. Notes uses its existing local notebook location and preserves saved data; its toolbar does not change file format or history. The delivery README contains launch instructions, known limitations and current acceptance status. Public native release assets and native-module tags require a separate owner decision.

## Remaining qualification and distribution work

Human IME/VoiceOver usability and additional host/display qualification remain explicitly outside the current support claim. Obtain results before expanding those claims. Native SDK/app distribution, coordinated native-module tags, signing/notarization and changed-payload audit require a separate owner-authorized release. Cosmetic or broader-framework work remains in the backlog.
