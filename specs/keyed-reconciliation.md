# Keyed reconciliation

Each managed UI tree uses unique stable node IDs. Inserting, removing, reordering
or reparenting compatible nodes preserves their native widget identities. Nodes
are compatible when their ID and kind match. Extensions also require the same
concrete type, and pointer extensions require the same instance. Replacing an
extension instance is an explicit native-resource lifetime boundary.

Descriptors and extension support are validated before mutation. New native
widgets are allocated before retained widgets move; allocation failure removes
new widgets and keeps the old managed tree. Extension Sync failures propagate to
the driver and terminate the run; arbitrary extension side effects cannot be
rolled back. Extensions must return one attached root widget and own its native
subtree. Application views describe fresh nodes; callbacks are refreshed while
widgets retain focus, selection, scroll and renderer resources.

Each workbench panel is a managed tree with its own ID scope. Dock containers may
be rebuilt for topology changes, but panel roots are parked in the same native
scene, then reconciled into their new containers. Removed and incompatible roots
are destroyed using the native widget lifecycle, releasing their owned resources.
A failed dock rebuild clears its native contents and caches; a later direct Sync
can rebuild cleanly. It does not roll back arbitrary extension side effects.
A managed host contains exactly one described root; application code must not
append unrelated native children to that host.

Validation covers sibling insertion/reordering/removal, cross-parent moves and
ancestor inversion, replacement cleanup, allocation failure, refreshed callbacks,
retained state, and content identity during dock moves. These are headless native
adapter contracts; live focus/accessibility behavior still needs host validation.
