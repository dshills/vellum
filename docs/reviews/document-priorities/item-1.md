# Item 1 — focus and Undo/Redo diagnosis

The build24 Documents application was observed on macOS 27.2 with its isolated
A/B fixture. A fresh B note accepted `B café 日本語\nalpha beta alpha` through
paste after keyboard navigation verified editor focus. Command-F opened Find.
After committing the query, two matches were reported. Keyboard navigation to
the replacement field and Replace All produced `gamma beta gamma`.

Escape was sent separately, and the returned accessibility snapshot confirmed
Find dismissal and editor focus. Only then was Command-Z sent: both replacements
were undone in one transaction. A separate Command-Shift-Z restored both. The
previous failure batched Escape and Undo across asynchronous focus/owner updates;
that result is not evidence that settled keyboard Undo fails.

Text-field pointer clicks through the desktop tool did not establish replacement
focus, while keyboard navigation did. The regression in [find_routing_offscreen_test.go](../../../examples/macapps/notes/find_routing_offscreen_test.go), using [viewport_offscreen_test.go](../../../examples/macapps/notes/viewport_offscreen_test.go), exercises the
actual managed hit-testing route at measured replacement bounds, then Replace All,
Escape, Command-Z and Command-Shift-Z, with explicit settlement between phases.
It does not send descriptor callbacks in place of pointer input. Domain setup
and accepted query/replacement values remain controlled fixture initialization.
Software routing cannot qualify physical pointer behavior or prove the external
tool's coordinate/AX behavior. Do not infer a runtime fix from a tool refusal.

No runtime behavior or API is changed. Desktop automation must observe focus
before typing and settle scope transitions before sending a dependent shortcut.
The existing build24 package is not relabelled as a build of this test change.

Verification: focused Find/layout race checks pass with Go 1.25.6 and offscreen routing. Prism staged review with unchanged defaults completed in full (Wafer/GLM-5.3), exit 0, with two low documentation findings. Both are corrected above; no runtime/test source changed after review. The original [review receipt](item-1-prism.json) is retained.
