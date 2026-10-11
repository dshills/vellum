# Independent managed grid example

The separate [Grid application](../../../examples/grid/README.md) uses the public
portable model and managed native adapter in two independent windows: 10,000 and
one million procedural rows, each with 20 columns. Stable domain keys survive
constant-time immutable reverse/even-row projections. The source computes cells
and key inverses; it does not allocate a million-row dataset or key index.

At the same 960×1228 logical viewport, both source sizes require exactly 400
visible cells, 520 materialized cells, 52 RowKey calls, 52 IndexOf calls and 520
Cell calls. These are bounded access counts, not storage memory or latency
measurements. Geometry fixtures include fractional scales at 96/120/144/192 DPI
with odd logical row/header/column sizes.

The offscreen application routes actual primitive mouse, wheel and keyboard
events through managed widgets. It covers far scrolling, frozen regions, stable
selection after source replacement, stale presses, disabled/modal admission and
window retirement. Completed software frames, model state and measured geometry
match at owner drain batches 1, 7 and 64. No direct callback invocation stands
in for routed input. The standard example check builds and vets the module; the
full repository script also runs its offscreen race tests and vet.

The separate desktop smoke waits for actual Cocoa Grid frames, checks bounded
pools and frozen geometry, jumps to the last million-row cell, preserves a stable
key across an immutable source change and verifies inspection refusal after
logical close. It records native runtime return and process exit. Its operations
are explicit owner commands, so it does not establish physical input or desktop
compositor teardown. The unchanged editor desktop smoke is a regression check
for composition, field handoff, range geometry and document workflows.

The first desktop attempt exposed a smoke assumption: retained selection is
separate from viewport visibility after replacing a source. The corrected check
asserts the selected key and revision under the owner, then explicitly reveals
that key before verifying its rendered cell. A regression covers the removed
scroll anchor and retained offscreen selection. The initial failed desktop run
remains historical evidence. The status label also has a bounded application
width so long selection text does not expand the window's minimum width.

The bounded [live input observation](live-input.json) uses the exact integrated
race binary in a private unsigned app. An externally observed cell activation
selects row 2/column 2; native Down advances to row 3 and PageDown to row 17.
The coordinate-click attempt produced no selection change, so physical pointer
acceptance remains open. Native Quit leaves no matching private process. These
observations do not establish compositor teardown, broad AX or human workflows.

Qualification binds production sources, the dependency graph and logs before and
after execution. Earlier v2 passes against an intermediate adapter remain
historical; final acceptance uses the corrected native source map. The combined
[Prism review](../../reviews/macos-grid-native/README.md) covers this consumer and
its integration wiring along with the adapter, with a separate consumer commit.

[qualification.json](qualification.json) and [handoff.json](handoff.json) bind
the corrected consumer to four passing portable/router/vet/build checks and the
approved adapter. The parent [integrated qualification](../macos-grid-native/qualification.json)
records both real Cocoa smoke executions. Earlier proof, source and logs remain
in [evidence.json](evidence.json).

Loaded-data storage, variable heights, range/all selection, cell editing,
tree expansion, native service workflows, human input/accessibility and sustained
performance remain outside this bounded example. It does not complete P6 or the
macOS framework milestone. No OS settings changed.
