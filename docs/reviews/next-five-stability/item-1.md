# Item 1: native palette acceptance

Target: corrected private demo suite revision 2, source
`9dc6012e0b1f765c4272182db6d334777c8183d2`, based on v1.1.0 and containing the
palette capture fix subsequently released in v1.1.1. The source tag's metadata
commit does not relabel this existing binary. Host: macOS 27.2.

The ordinary application was used, without native probe/fault-injection flags.
The selected app is Vellum Placement.app from the revision-2 suite. Its archive
SHA256 is `161709b0964eed836dbcc64519f979e0bd4da7ab33dfb5aab9e23bb8c98a199f`;
all 19 ZIP-extracted app bundles previously passed strict signature checks.

## Procedure and evidence boundary

The desktop tool inspected the native window, focused Block, and pressed Return.
The app showed block-1, one object, and Undo 1 / Redo 0. Native Undo restored zero
objects, Undo 0 / Redo 1 and a clean Untitled drawing. This establishes keyboard
placement and Undo, not a successful mouse drag.

Desktop-tool drag attempts produced no object change even after the user brought
the window forward. They are recorded as unsuccessful input delivery/observation,
not passed placement acceptance and not proof of a framework drag defect.

The operator checklist is:

1. Hold Block without movement; surrounding controls and canvas stay stationary.
2. Drag Block onto the canvas; exactly one object appears.
3. Undo then Redo; object count changes from one to zero to one.
4. Enable Reject all drops and drag Note onto the canvas; count stays one.
5. Disable rejection, begin dragging Connector, then Escape; count stays one.

Operator results and final native inspection are recorded in item-1-result.json.
The held-state geometry judgment is an operator observation; it is separate from
the measured software-routing regression in placement-capture-layout. This bounded
check does not establish other display scales, hardware, long-duration behavior,
IME or VoiceOver qualification.

The user reported “All five checks passed”. Final native inspection independently
showed exactly one selected Block (block-2, 80 × 50), Undo 1 / Redo 0, rejection
off and Connector focused after cancellation. The drawing is unsaved and was left
open; acceptance did not save or discard user-operated work. The operator reported
the intermediate zero/one Undo/Redo states and stationary held geometry; the tool
independently confirmed the final state only.

Default Prism (Wafer / GLM-5.3) completed with full coverage, no medium/high
findings and three low record-clarity findings. Added a record-finalization time,
explicit bounded outcome, and structured tool-attempt verdict; failed tool input
was not converted into a pass. The original review receipt is retained.
