# Framework implementation phases

This implementation proceeds alongside the unresolved foundation checks. Those
checks still prevent a production-ready claim. They do not prevent building and
assessing the workbench API. Each phase receives Prism review, relevant tests,
and a separate commit.

1. Application lifecycle and keyed UI API — implemented; native driver and executable example.
2. Commands and cancellable tasks — implemented; shared registry, menus, toolbar, palette, shortcuts and bounded owner-thread completions.
3. Docking, document tabs, inspector and layout persistence — implemented; menu-driven move/split, resizable native panes, transactional versioned layouts.
4. Virtualized tables and trees — implemented; 1,000,000 real records, 100,101 tree nodes, 20/24 native row pools, cancellable sorting and stable domain selection.
5. Shared-device GPU canvas with navigation and selection — implemented; 96 retained GPU objects, pointer-anchored zoom, pan, selection and inspector integration.

The native driver exposes multiple desktop windows in one process main loop. Root UI ownership and
logic tests are portable; native runtime support is not claimed across platforms.

Phase 5 has passing model/native regression tests and real GPU pixel verification.
The live pointer automation check remains unverified because the computer-use
window coordinate mapping targeted a different region; see the phase 5 review
evidence. This limits runtime certification, rather than the implemented API scope.

## Framework expansion

1. Keyed reconciliation — implemented; native identity and retained state survive
   insert/remove/reorder/reparent and docking moves. See the
   [contract](../specs/keyed-reconciliation.md). Portable/native checks pass;
   live focus/accessibility workflows remain acceptance checks.
2. Layout and theming — implemented; optional value styles, application-local
   themes, native layout/appearance mapping and reset semantics. See the
   [contract](../specs/layout-theme.md).
3. Standard controls and interaction — implemented; controlled values, queued events,
   keyed tabs/scroll/modal children, focus requests and semantic metadata. See
   the [contract](../specs/controls.md); platform screen-reader publication remains
   an acceptance gate.
4. Multiple windows and platform services — implemented; scoped callbacks,
   independent close lifetimes and macOS clipboard/file panels/application menus.
   See the [contract](../specs/windows-services.md).
5. Independently consumable SDK and benchmarks — implemented; driver-owned pinned
   bootstrap, standalone consumer/generator and repeatable host measurements. See
   the [contract](../specs/sdk.md). Publication and production performance remain
   release gates.

## Framework completion work

1. Workbench interaction — implemented; pointer docking/reordering, independent
   floating windows, redocking and transactional session persistence. Domain data
   survives native adapter remounts. See [contract](../specs/workbench-interaction.md)
   and [evidence](reports/workbench-interaction/README.md).
2. Native accessibility and keyboard workflows — implemented for macOS managed
   controls, active dock panels, modal scopes and bounded virtual views. Native
   selector probes and keyboard regression tests are separate from external
   screen-reader/IME acceptance. See [contract](../specs/accessibility.md).
3. Rendering hardening — implemented; demand painting, synchronized composition,
   live resource rebuilds and device-loss error/shutdown. Local race checks cover
   resize/minimize/restore and teardown. Automatic physical GPU recreation,
   cross-display DPI and sustained workload budgets remain open. See
   [contract](../specs/rendering.md) and [evidence](reports/rendering/README.md).
4. Windows/Linux services and qualification tooling — implemented; native-main
   clipboard, OS chooser transports, menu widgets, target smoke/CI harness and
   explicit strict qualification gates. This macOS host cannot establish target
   runtime, TSF/IBus/Fcitx, UIA/AT-SPI or native Wayland support. See
   [contract](../specs/platform-support.md) and [evidence](reports/platform-support/README.md).
5. Versioned private SDK — implemented; alpha version API, independent consumer
   scenarios, integrity-checked relocation, tutorial and private application
   packaging. The owner selected `github.com/dshills/vellum` and MIT for Vellum's
   own code. Distribution audit and actual target acceptance remain public-release
   gates. See [contract](../specs/release.md)
   [tutorial](tutorials/building-an-app.md) and [evidence](reports/sdk-release/README.md).

Per-phase Prism reviews and verification records accompany these implementations.
The private SDK evidence records a tested precommit snapshot; its clean committed
archive is generated after committing, so its future commit identity cannot be
embedded in the earlier evidence. These
implementations extend the experimental framework; they do not certify a public
or production-ready cross-platform release.

## macOS completion milestone

The current goal focuses framework completion and acceptance on the local Mac;
Windows/Linux runtime work is deferred. See the [macOS completion checklist](../specs/macos-completion.md)
for the complete remaining scope.

Document history is implemented with bounded sparse changes, undo/redo, save
revision validation, dirty status and focused native keyboard precedence. The
workbench records component edits and updates source revisions on replay. See
[document history](../specs/history.md) and [document persistence](../specs/document.md).
The portable session now saves/loads on a worker and guards unsaved native closes;
the workbench persists component data, geometry and ordered file references.
Actual file-panel selection remains an acceptance gate. Native command phases also run in focus-free windows;
startup shutdown and late native creation use stable stop/admission handling.
See [local evidence](reports/macos-history/README.md).

External file drops are implemented through `ui.Node.FileDrop`, bounded absolute
path payloads, synchronous owner admission, clipped target geometry, active
tab/workbench eligibility and modal host fallback. The workbench references
dropped files with undo/redo and a bounded status display. Portable/native race
checks and a synthetic cursor/Cocoa receive fixture pass; the physical Finder
drag attempt did not produce delivery and remains unqualified. See
[contract](../specs/file-drop.md) and [evidence](reports/macos-file-drop/README.md).

Native dispatch now acknowledges failure, preserves panic identity, closes
admission and avoids a terminal crash-dialog wait. Desktop wakes serialize with
GLFW termination. Reproducible patch22, twenty repeated race runs and real native
loop/windowed checks support the bounded contract. See
[dispatch](../specs/native-dispatch.md) and
[evidence](reports/macos-native-dispatch/README.md); broad panic resource recovery
and remaining macOS qualification stay open.

Retained custom drawing now adds immutable indexed 2D meshes, world strokes,
million-sample envelope reduction and paint-ordered rectangle spatial culling.
The independent drawing application exercises a flat shared-device native
renderer, explicit resource accounting, retention and snapshot color handling.
See [contract](../specs/drawing.md) and [bounded evidence](reports/macos-drawing/README.md);
the broader P4 workloads and Mac qualification gates remain open.
