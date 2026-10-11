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

[Retained GPU drawing](drawing.md) adds arbitrary indexed 2D meshes, strokes,
waveform envelopes and rectangle spatial culling, with an independent consumer.
[Retained images](images.md) add a real 4096² native texture workload with changed
tile uploads, alpha-safe filtering and independent-window retention evidence.
[Advanced GPU views](advanced-gpu.md) add custom shaders and a retained 3D
viewport with optional depth and native fault/resize/closure evidence.
An independent imaging consumer adds ordered compute/float-target preparation
and bounded asynchronous captures. P4 still needs broader performance qualification. The public [multiline editor](editor.md) now provides a supported model, managed
control and independent document application beyond the historical prototypes;
its native input/geometry evidence and separate read-only external AX check pass
against frozen sources. Human IME and VoiceOver acceptance remain separate. Semantic command/UI inspection exists through APIs
and native accessibility. Portable execution now has injected real/manual
workers and clocks, reserved owner results, bounded drain batches and separate
worker/delivery accounting.

The [managed replay checkpoint](replay.md) is internally qualified on the
offscreen software backend with supplied, digest-checked fonts and a declared
theme. It uses actual retained-widget pointer/key routing, focus/modal behavior,
measured geometry and completed software frames. One fixed 51-operation record
with nine captures matches exact model, semantics, focus, UTF-16 selection,
geometry and image identities at batches 1, 7 and 64. Explicit worker selection,
clock advancement and originating timer dispatch preserve observable pending
work; future-only timers do not force an implicit advance. The scope is one
active managed application per process, with multiple windows sharing its clock.
See the [checkpoint report](../docs/reports/macos-replay/README.md).

Full P9 professional workload and large-data replay acceptance, external
application automation transport, native services/input/accessibility, GPU
presentation and sustained performance qualification remain open. This bounded
software fixture does not complete the macOS milestone.

Remaining Mac acceptance includes broader external AX and human
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
