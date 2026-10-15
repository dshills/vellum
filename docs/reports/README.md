# Foundation reports

These reports preserve investigation outcomes and their original validation scope.
The [implementation status](../implementation.md) describes the current framework.

- [Bounded ListBox control](macos-list-box/README.md)
- [ListBox native writability metadata](macos-listbox-metadata/README.md)
- [Composed ListBox and Canvas2D coverage](macos-composed-controls/README.md)

- [Tree and retained-view geometry publication](macos-tree-publication/README.md)
- [Stored-tree CPU/GPU route investigation](macos-tree-gpu/README.md)
- [Large stored-tree adapter](macos-large-tree/README.md)
- [Mac clipboard data (offscreen)](macos-clipboard-data/README.md)
- [Mac tray and native service lifetime (offscreen)](macos-tray/README.md)
- [Mac local notifications (offscreen)](macos-notifications/README.md)
- [Native grid timing foundation](macos-grid-performance/README.md)
- [Stored-data worker grid](macos-grid-loaded/README.md)
- [Managed native grid adapter](macos-grid-native/README.md)
- [Independent managed grid example](macos-grid-example/README.md)
- [Indexed portable grid foundation](macos-grid-model/README.md)
- [Managed software UI replay](macos-replay/README.md)
- [macOS retained image drawing](macos-images/README.md)
- [macOS retained GPU meshes and scalable workloads](macos-drawing/README.md)
- [macOS document persistence and unsaved close](macos-document/README.md)
- [macOS native dispatch and termination](macos-native-dispatch/README.md)
- [macOS native external file drops](macos-file-drop/README.md)
- [macOS document history and native command lifecycle](macos-history/README.md)
- [Workbench interaction](workbench-interaction/README.md)
- [Rendering hardening](rendering/README.md)
- [Windows/Linux services](platform-support/README.md)
- [Private alpha SDK packaging](sdk-release/README.md)
- [Local SDK checks and performance](sdk/README.md), including [final integrated evidence](sdk/final/README.md)
- [P0 feasibility](P0_RESULTS.md)
- [Initial foundation](FOUNDATION_RESULTS.md) and [remediation](REMEDIATION_RESULTS.md)
- [Native input](NATIVE_INPUT_RESULTS.md)
- [Foundation checks](FOUNDATION_CHECKS_RESULTS.md), [contracts](FOUNDATION_CONTRACTS_RESULTS.md), and [integration](FOUNDATION_INTEGRATION_RESULTS.md)
- [Acceptance audit](FOUNDATION_ACCEPTANCE_AUDIT.md) and [regression fix](FOUNDATION_REGRESSION_FIX.md)
- [Input ordering fix](INPUT_ORDERING_FIX.md)
- [Visual text geometry fix](VISUAL_TEXT_GEOMETRY_FIX.md)
- [Multiline native adapter](MULTILINE_NATIVE_ADAPTER.md)
- [Large-document input](LARGE_DOCUMENT_NATIVE_INPUT.md)
- [Editor visual navigation](EDITOR_VISUAL_NAVIGATION.md)

The [validation plan](../../specs/prototype-plan.md) defines the research gates.
A passing local check does not imply production or cross-platform certification.

- [Completed-outline composed retention](macos-tree-retention/README.md): 1000 native cycles on committed source with strict native tree completion and sampled native memory; broader P6/P11 remains open.

- [Composed native actions](macos-composed-actions/README.md): 500 acknowledged Cocoa action steps across 25 windows, including worker-built tree disclosure; physical and sustained acceptance remain open.

- [Inline color picker](macos-color-picker/README.md): reusable RGBA/hex fields, retained preview and scoped Mac action/lifetime checks.

- [Interacting lifecycle fixture](macos-interacting-retention/README.md): 24-step color-inclusive preflight passes; the 1000-window campaign fails after 343 completed windows, with preserved attachment-close evidence and clean teardown.

- [Accessibility retirement](macos-ax-retirement/README.md): late-refresh registry retirement fix and copied first-close diagnostics, verified with offscreen race checks; prior endurance cause remains open.

- [Palette placement](macos-palette-placement/README.md): typed public API, same-window capture, independent demo and code-only routing evidence.

- [Owned private Mac build](macos-owned-build/README.md): observed compiler pairing for executable/map, sampled input stability and bounded logs; runtime and distribution rights remain separate.

- [One-pass glyph fingerprint](macos-glyph-fingerprint/README.md): fused geometry validation/internal hashing with exact normalized path equality and software capture parity.
