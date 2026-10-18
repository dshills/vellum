# Documentation

Start with the [project overview](../README.md), [development guide](development.md)
and [implementation status](implementation.md). For current Mac work, follow the
[delivery plan](../specs/macos-delivery.md).

- [Research and recommendations](research/README.md): source-backed investigation and architecture alternatives.
- [Architecture decisions](decisions/): recorded choices and the evidence available when each was made.
- [Foundation and SDK reports](reports/README.md): investigation outcomes, fixes and remaining validation gaps.
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
