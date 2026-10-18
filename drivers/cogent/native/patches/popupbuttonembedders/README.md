# Direct Button embeddings in popup inspection (stage 84)

The bounded actual-popup inspector now uses Core's existing `AsButton` embedding
capability, as the menu implementation already does. Vellum MenuButton items
embed Button for lifecycle ownership; an exact `*Button` assertion omitted them.
The base must be initialized and share the direct child's AsTree identity.
Forwarding a different button or wrapping a nested button is not a direct
embedding and does not create an inspection target.

All existing ownership, stage precedence, record/label/byte budgets, detached
copies, clipping, visibility/read-only state and monotonic lifetime IDs remain
in force. Every visited child is charged before conversion; passive separators
remain absent. No recursive traversal, registry or new public API is added.

The focused software regression uses real embedded and standard buttons,
duplicate labels, a disabled embedding, a separator and a foreign-button wrapper.
It checks actual bounds, distinct/stable IDs, exactly-once measured pointer input,
retirement and new IDs after reopening. The old implementation fails by omitting
the two embedded items; the corrected source passes. These software results are
separate from native/physical qualification. The native navigation catalogue
supplies the actual MenuButton integration record and its explicit limitations.
