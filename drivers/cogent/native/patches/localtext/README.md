# Passive local text layout

Patch 0055 lets an admitted fixed-allocation local root update passive Text
without the allocation shape that its required sizing pass immediately replaces.
Value/updater callbacks and ordinary styles remain active; old shaped runs are
cleared before new rich styles. The scoped updater flag restores on return.
Sizing clears pending state; Render restores ordinary allocation shaping if the
request is canceled before sizing. Normal/full-layout updates stay immediate.

The helper refuses interactive text, foreign/detached/retired roots and children,
unrequested roots, wrong ancestors and full-layout fences without mutation.
Caller must retain passive styles and size every updated label before paint.
Vellum grid's existing styleDirty-to-layoutDirty relationship provides this rule.
The inline editor takes the ordinary full path.

Actual-shaper regressions cover one scheduled shape, interactive refusal and
immediate input-safe fallback, cancellation recovery and ordinary updates.
Existing grid pixel/focus/editing tests and native paired measurements qualify
this bounded change; no full frame, physical GPU or Mac acceptance is implied.
New helpers/tests are MIT; upstream text.go preserves its existing notice.
Preparation verifies original/output hashes and the patch SHA256.
