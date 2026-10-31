# Item 1: editor click focus

Resolved within the recorded scope on 2026-10-31. The operator physically clicked
and typed in both the body and title of the isolated Vellum Suite Documents
window and reported: “Both the body and title receive text correctly when clicked.”
The app uses v1.1.2 source 121dde859369b3ce091fe66cdfa2d34367e23be1 with a temporary synthetic data profile at
/private/tmp/vellum-v112-focus-check (may expire), not an existing daily notebook.

The prior desktop-tool observation did not reproduce with physical input. No
runtime fix is justified by this evidence; this is not a blanket claim about all
focus, IME or accessibility behavior. The earlier automation click remains a
recorded delivery limitation, not a confirmed product defect.

The new offscreen regression routes down/up events between a title field and
document body in code and prose modes across three cycles. It checks focus both
before and after settled view invalidations. The focused race test passed after
final diagnostic changes (1.930s).

Default Prism completed full coverage on the initial and revised test. Its medium
finding (ignored Invalidate error) was fixed. Two low missing-window/snapshot
length diagnostics were added. The low loop-capture concern does not apply:
drivers/cogent/go.mod requires Go 1.25.6 with per-iteration loop variables. The
receipts are preserved alongside this record. Physical acceptance comes from the
operator report, independently of the routed offscreen test.

Regression: drivers/cogent/editor_pointer_focus_offscreen_test.go, TestPointerFocusMovesBetweenTitleAndDocument. From drivers/cogent with prepared dependencies, run `go test -race -tags offscreen -run ^TestPointerFocusMovesBetweenTitleAndDocument$ .`.

Closing default Prism review completed full coverage with three low traceability findings. The log now distinguishes the historical build25 observation from the later v1.1.2 physical check; the temporary path is explicitly a data profile that may expire, not a source checkout; the test command is identified above. Build25 bytes were not relabeled or physically requalified by this later check.
