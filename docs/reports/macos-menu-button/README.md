# Mac MenuButton checkpoint

Vellum adds `ui.MenuButton`, a public flat action menu with stable option IDs,
duplicate labels, disabled actions, owner callbacks, button artwork/appearance,
managed keyboard traversal and Cocoa semantics. The gallery demonstrates it on
Icons and actions. See [contract](../../../specs/menu-button.md) and the
[control catalog](../../controls.md).

Internal review caught destructive popup closure from a peer window's owner
drain and missing custom-wake forwarding from the modal scene. Closure now
revokes owner state first and destroys native resources only in originating
reconciliation/teardown. The menu forwards custom wakes. Real offscreen
regressions cover pointer selection, Tab/Shift-Tab/Escape, per-popup manual
incarnations, stale handler rejection, replaced options, peer-drained retirement
and owner updates delivered through the actual popup scene.

The first two native runs failed scoped Cocoa publication. The second captured
Cocoa still at initial generation: disabled items encoded a null action list,
which the production parser rejected atomically. Disabled items now encode an
empty array. A Darwin regression checks both the accepted payload and null-action
negative control. The failed runs remain in the archive.

The final bounded Cocoa run passed in 2.310 seconds: 20 selection cycles chose
`second` despite duplicate labels, retained one anchor object, created 20 distinct
item object lifetimes, refused disabled and dismissed item actions, and matched
positive item frames to Cocoa hit tests. An owner-posted disabling update closed
the modal menu; re-enabling and Cocoa collapse also completed. These are actual
in-process native selectors, not human screen-reader/physical-input acceptance.

Full portable race tests, full adapter race/offscreen tests and adapter vet
passed on the initial reviewed source. The final two-file delta (active-menu
registration ordering and fixture geometry readiness) passed focused race tests
and a rebuilt native run; unchanged files inherit the full-suite result. Earlier focused race repetitions
also passed; transient fixture mistakes and source-stage outcomes remain recorded
in [qualification](records/qualification.json). Prepared dependencies remained
unchanged (6202 files); this phase adds no upstream patch.

The [evidence index](records/evidence-index.json) binds the archived raw check
receipts, logs, producers, source bytes and preserved internal review. Prism
review output and the final checkpoint disposition are stored beside the archive
when the checkpoint is finalized. The archive excludes executables. A correction archive/index records the final
delta checks, original Prism findings and revised source. The anchor identity
field is an object token; all 20 samples contain the same token. The final native
recheck waits for item layout, preserving the earlier cycle14 premature-readiness
failure.

This control checkpoint does not close P6 timing, the primary/all-components
P11 workload, composed 1000 lifecycles, overnight behavior, cross-display/human
input/accessibility or SDK distribution gates. Nested action menus and shortcuts
are future API work.
