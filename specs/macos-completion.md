# macOS framework completion

The active completion target is a professional Go-native UI framework on the local
macOS development machine. Windows/Linux qualification is deferred. Existing
portable packages and target adapter sources are retained, but cross-platform
acceptance does not define this milestone's success.

Completion requires evidence for the intended professional application class:
application ownership and cancellation; standard controls and keyboard workflows;
commands; document edits and undo/redo; docking and independent windows; scalable
data views; custom GPU drawing; native clipboard, menus, choosers and file drops;
high-DPI geometry; functioning native accessibility/input; deterministic tests;
measured sustained performance; independently consumable SDK and Mac packaging.
The research/prototype plan and detailed subsystem contracts remain authoritative
for their acceptance workloads. Historical claims are not current qualification.

## Current implementation and remaining work

Application ownership, cancellable tasks, commands, controlled standard controls,
keyed reconciliation, layout/theme, docking/window sessions, virtual tables/trees,
shared-device rectangle canvas, Cocoa clipboard/menus/file panels, demand painting,
resource lifecycle checks and private SDK/bundle tooling exist.

Native dispatch now has explicit synchronous failure and desktop wake termination
semantics. See [contract](native-dispatch.md) and its evidence; broad panic resource
recovery remains outside that contract.

Document history adds bounded sparse undo/redo, revision-aware dirty/save handling
and focused native text precedence. [Document persistence](document.md) adds asynchronous save/open and native
unsaved-close prompts with a working component-document consumer. Its bounded
acceptance evidence remains separate from actual file-panel selection.

External file-drop callbacks and Mac file-URL routing are implemented; physical
Finder workflow acceptance remains separate. See [contract](file-drop.md).

Remaining framework work includes richer reusable
custom drawing (beyond bounded rectangle scenes) and supported integration of the multiline editor
beyond historical prototypes. Semantic command/UI inspection exists through APIs
and native accessibility; application automation transport and deterministic UI
replay need an explicit public contract and working consumer evidence.

Remaining Mac acceptance includes external AX client/text ranges and human
screen-reader workflows; candidate-panel placement through scrolling/focus/scale;
cross-display DPI and monitor changes; physical GPU loss policy acceptance;
representative sustained frame/input latency, idle CPU and bounded font/native/GPU
memory; actual chooser selection, native menu activation, clipboard roundtrips and
external file-drop workflow. Automated selector/fixture tests cannot substitute
for broader user-facing acceptance. OS preference changes require specific user
authorization and restoration; routine checks do not change those preferences.

Public distribution also requires the third-party/native dependency audit.
The selected namespace and MIT license resolve owner choices but do not certify
that audit, signing/notarization or production qualification. Existing version is
v0.1.0-alpha.1 and macOS remains experimental until acceptance is complete.

## Review and stop points

Use internal agent review during implementation. Run Prism on significant completed
batches using the user's selected Anthropic Claude Sonnet 5.5, resolve findings,
run affected tests and commit stable checkpoints. Avoid repeated unchanged Prism
runs. Native acceptance must record source identity, command and exit status;
qualification remains open wherever evidence is absent.
