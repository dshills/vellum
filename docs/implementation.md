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

The first driver deliberately exposes one native window. Root UI ownership and
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
4. Multiple windows and platform services — in progress.
5. Independently consumable SDK and benchmarks — in progress.
