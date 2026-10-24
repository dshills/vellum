# macOS collection view

The public `ui.CollectionView` renders a bounded Finder-style icon grid using the
existing managed multiple ListBox routing, selection, accessibility and lifecycle
infrastructure. It supports at most 256 stable items. Item labels and eligibility
are separate from IDs. Immutable shared thumbnails are optional; bundled symbols
provide the fallback. No paths, filesystem actions or image decoding run inside
rendering.

`View` detaches item/selection slices and emits native `CollectionLayout` metadata
on a multiple ListBox. Width defaults to 112 logical pixels (configurable 80–256).
The native wrapping layout computes columns from available width, padding and
scrollbars. Keyboard navigation counts actual row origins in scene pixels, so
DPI and clipping do not need a guessed conversion. Tile icons/thumbnails are 48
logical pixels, with centered labels (abbreviated after 28 runes; accessibility keeps full names), a rounded selection background and the
existing active/inactive selection/focus policy.

Pointer selection supports Command toggle and Shift ranges. The owner accepts
`ListSelectionEvent` using `Apply` against current accepted membership, preserving
queued independent toggles. Arrow keys move through the measured grid; vertical arrows skip disabled cells
in the same column and stay put when the final row lacks that column. Shift
extends ranges; Command-arrow moves focus without replacing selection;
Command-A selects enabled items. Return and unmodified double-click request one
activation independently of selection. Space selects/toggles through the same
selection path. Tab follows existing native focus navigation.

Native callbacks use the existing bounded originating-window queue, row epochs,
catalog lease and visibility/modal scopes. Each detached presentation carries
immutable retained-component identity. Replacing that component or switching
between ordinary ListBox and collection presentation renews row/catalog leases,
including an intervening ordinary-list ABA transition. Label, thumbnail and
accepted-selection changes on the same retained component preserve eligibility. Disabled, removed, replaced, hidden or
retired controls cannot deliver stale queued actions. Programmatic updates do not
emit callbacks. Selected state stays application-owned even if proposals are
rejected. Label markup is escaped, and existing ListBox accessibility publishes
named selected options.

Focused checks cover detached snapshots, validation, controlled proposal
composition, grid movement, actual responsive row layout, routed arrow/Shift/
Return behavior, rejection and stale activation. Native desktop appearance and
physical pointer checks are recorded separately by the containing demo checkpoint.

This is a custom macOS-style collection, not NSCollectionView. It does not include
virtualization, item dragging, marquee selection, filesystem operations, automatic
OS appearance tracking or Quick Look. The existing 256-item bound applies.
