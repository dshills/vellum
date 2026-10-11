# Managed native grid checkpoint

`ui.DataGrid` connects the public indexed `virtual.Grid` model to retained native
widgets. Four clipped panes keep headers and leading columns frozen while
bounded text pools render the current two-axis viewport. The host owns keyboard
focus; pointer, wheel, arrow and page events update the caller-owned model before
the `GridChange` notification. Rendering uses retained projection data and never
fetches source cells or calls application callbacks. See the
[native contract](../../../specs/grid-native.md).

Completed native frames publish clipped cell/header geometry through
`cogent.InspectGrid`. The retained accessibility tree exposes visible table rows,
cells and headers with absolute indices, full source counts and stable identities.
Cocoa selector fixtures check those table attributes. Hidden cells are not
materialized for accessibility. Generated identifiers reserve the `grid:` prefix;
ordinary descriptors cannot alias dynamic grid targets.

Internal review reproduced nine defects and verified their corrections: unchecked
local viewport bounds, stale scene geometry on pointer release, fractional shared
edges, finite wheel overflow, cell and row identity collisions, inactive native
tab admission, interrupted mount registry retention and Workbench cell-press
focus namespacing. Original red test bodies and source-bound results remain in the
evidence; earlier passing versions are historical and do not qualify the final
corrections. Interrupted mounts release their native claim and restore a previous
valid remount binding. The adapter does not recover source panic or Goexit;
registry cleanup precedes private widget destruction. Arbitrary failures in
private destruction callbacks remain outside this contract.

The repository checks cover the portable and managed contracts, actual offscreen
pointer/key/wheel routing, clipped geometry, modal/disabled/stale admission,
bounded pools, source transitions, teardown and inert Cocoa table selectors.
The final integrated check uses repository-relative modules and sequential Cocoa
Grid/editor smoke. The Grid smoke measures actual completed frames and performs
explicit owner model commands; offscreen primitive routing is separate evidence.
The [independent example report](../macos-grid-example/README.md) records its
procedural workload and desktop scope. Source identities, commands, process exits
and historical failures are retained with this checkpoint.

[qualification.json](qualification.json) records the final integrated pass over
6,533 unchanged inputs. It explicitly retains four passed adapter checks from
the source-stable run that later failed at the consumer smoke; the corrected
consumer and both desktop smoke checks are new passing executions. The
[native qualification](native-qualification.json) records seven producer gates
over 8,424 unchanged effective inputs and the exact 26-file
[source map](source-map.json). Raw historical source/log/proof documents remain
round-trippable in [evidence.json](evidence.json).

Prism found overwritten model errors in five test state reads. The corrected
test checks model and owner errors separately; four affected offscreen tests
pass with race count3 and driver vet passes. The
[test correction qualification](test-correction-qualification.json),
[exact corrective evidence](test-correction-evidence.json) and
[integrated binding](test-correction-binding.json) distinguish this single test
change from the earlier source-stable runtime qualification. Runtime code,
original reviewer regression bodies and native smoke binaries remain unchanged.

The [Prism review record](../../reviews/macos-grid-native/README.md) covers the
adapter, example, specifications and integration wiring together using the
selected Anthropic Claude Sonnet 5.5. Each implementation phase is committed
separately after that combined review and final evidence verification.

This is a fixed-height grid checkpoint, not complete P6 or the Mac framework
milestone. Variable heights, range/all-row selection, cell editing/conflicts,
cancellable indexed projection, large tree workloads and sustained latency,
frame and memory measurements remain open. In-process AX and native selector
fixtures do not qualify external AX clients or human VoiceOver/IME workflows.
No macOS preferences changed, and no public release, signing or notarization ran.
