# Documentation

Start with the [project overview](../README.md), [development guide](development.md)
and [implementation status](implementation.md). For current Mac use, follow the
[v1 support guide](macos-v1.md), [stable API contract](../specs/stable-api.md)
and [release closeout](../specs/macos-v1-release.md). The older delivery plan
retains broader qualification work.

- [Mac input widgets](mac-widget-integration.md): public API setup, themes, owner values and focus helpers.
- [Mac widget demo](mac-widget-demo.md): run the private light/dark gallery.
- [Repository layout](repository-layout.md): package, module, command and archive boundaries.
- [Research and recommendations](research/README.md): source-backed investigation and architecture alternatives.
- [Architecture decisions](decisions/): recorded choices and the evidence available when each was made.
- [Foundation and SDK reports](reports/README.md): investigation outcomes, fixes and remaining validation gaps.
- [Current macOS public packaging](reports/macos-v1-public-packaging/README.md): approved audit, final local candidates, launch results and separate publication decision.
- [Historical delta-audit packet](reports/macos-v1-distribution/README.md): exact private inputs bound by the current approval.
- [Distribution checks](distribution.md): packaging the native foundation fixture.
- [Review records](reviews/): committed reviews and verification evidence.
- [Architecture diagrams](assets/): research illustrations.
- [Specifications](../specs/README.md): implemented framework contract and validation plan.

Research, decisions, reports and review records preserve historical context. Their
past next-step instructions describe that stage of the investigation; consult
implementation status and the framework contract for the current implementation.
Recorded evidence may contain original paths and hashes. Those capture-time values
are retained; current document links and tooling use the reorganized paths.

- [Named canvas objects](reports/macos-canvas-accessibility/README.md): opt-in semantics, exact selection and completed geometry for retained Canvas2D.

- [Span header construction](reports/macos-span-headers/README.md): bounded scene-builder scratch and source-capture allocation evidence.
- [Packed glyph coverage](reports/macos-packed-coverage/README.md): owned cached-span replay with original admission and exact output checks.

- [Native integration session usage](desktop-integration.md).
- [Full native workspace checkpoint](reports/macos-native-workspace/README.md).

- [Native macOS application examples](../examples/macapps/README.md): independently runnable Go consumers and their working interactions.
- [Native application checkpoint](reports/macos-app-examples/README.md): verification and remaining visual acceptance.

- [Build a personal macOS application](building-macos-apps.md): independent SDK consumer, controlled state, commands, durable Notes files/recovery and packaging.
