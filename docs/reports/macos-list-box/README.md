# ListBox checkpoint

Public `ui.ListBox` fills the simple always-visible selection gap between a
dropdown and the indexed data views. It is a bounded single-selection control
with up to 256 stable-ID options, optional empty selection, disabled/read-only
states and controlled application callbacks. The gallery's Lists page and an
independent root-module consumer demonstrate public consumption. See the
[contract](../../../specs/list-box.md).

Native row widgets retain identity across reorder/relabel. Arrows clamp and
skip disabled rows, Home/End and measured page movement reveal their targets,
and initial-letter cycling needs no timer. Pointer/Space/Return/semantic
selection explicitly propose stable IDs; admitted native proposals preserve
FIFO even when the rendered descriptor still holds an earlier value. Rejected
changes and queue refusal restore the accepted model. Programmatic updates emit
no callback. The list occupies one managed Tab stop when enabled rows exist.

Origin guards revoke queued input across row removal/disabled cycles and
whole-list visibility, disabled, modal and window lifetime changes. Host AX set
captures an independent immutable option eligibility snapshot. Final modal
projection preserves the originating inner dialog's native closure lease.
Mac semantics use AXList/AXRow, all/visible/selected rows and children, accepted
label value and actual clipped frames. Offscreen routing verifies that a
semantic focus request reveals an already-focused row after wheel scrolling
moves it offscreen. Native focus/reveal qualification remains open.

Internal review requires source fixes for held-input ordering, catalog-slice
aliasing, same-focus reveal and in-place reorder invalidation. Regression
coverage includes accepted/rejected held actual keys at owner batches 1/7/64;
DPI 96/120/160/240 scrolling, pages, pointer/key input, raster differences,
large fonts, narrow layouts and style removal; real offscreen Workbench/modal
focus and native closure scope; stale row/host actions, saturation, read-only and
window retirement; inert Cocoa all/visible/selected row selectors and selected-row
setter delivery through the real owner bridge; and the
public gallery/SDK contracts. Acceptance requires preserving failed fixture
iterations in the evidence archive and excluding them from passing results.

Checkpoint acceptance requires root and full adapter offscreen race/vet checks,
native gallery compilation, private SDK packaging/relocation and pinned source
identity, plus internal source/evidence review and one Prism review using
Anthropic Claude Sonnet 5.5. Acceptance also requires retaining exact commands,
source bindings, outputs, provider findings and dispositions in
`records/qualification.json` and its indexed evidence archive. Offscreen test
executables run without a native desktop window. Live checks run only after the
human explicitly made the desktop available.

The external Mac AX inspector verifies AXList, 24 AXRows, seven initially
visible rows, selected-row setter delivery, owner rejection, read-only protection
against model changes, programmatic clear and stable-ID selection of an
offscreen row. Two native gaps remain **OPEN**. macOS reports `AXSelectedRows`
settable in read-only mode even though an attempted set emits no owner callback
and leaves selection unchanged. Foreground activation could not be verified;
one targeted pointer probe did not select its row, subsequent guarded OS-input
probes refused, and an external offscreen-row focus request did not produce
visible bounds in that session. These observations do not establish a cause.
Native pointer/key delivery and native focus/reveal remain unqualified; all
failed attempts and process cleanup records are preserved.

This is a materialized single-selection list, with count bounds rather than
application payload byte bounds. There is no multi-selection, virtual list,
drag reordering or timed typeahead claim. Physical Mac pointer/keyboard input,
human VoiceOver, native/GPU memory, sustained performance, endurance and Mac
distribution gates remain open. This checkpoint does not complete the framework
or qualify the existing full Mac/P6/P9/P11 acceptance gates.
