# Stored-data grid checkpoint

This checkpoint adds reusable worker-built sort/filter/height indexes and an
independent Mac consumer with actual stored million-row ×20 numeric data.
Applications keep domain snapshots separate from accepted view revisions; the
owner admits only the latest matching request. Workers build immutable indexes
and edit copies. A completed edit may remove its row from the filtered view
without losing the version check or leaving the draft open.

Apply settles the ordinary filter field before capturing its controlled value.
This fixes an actual routed final-key→Apply ordering failure without adding an
intervening frame or delay to the test. Explicit settlement retires older queued
changes, and repeated unchanged settlement does not repeat `Change`. Cell saves
continue to use the separate native inline-editor barrier.

## Validation

The [portable gate](core-v2/qualification.json) passes affected race3, all core
race tests and vet. The [native gate](native-v3/qualification.json) passes
pinned preparation tests, upstream concurrent cursor race20, affected
text/editor/grid/control race3 and both offscreen/default vet modes. It includes
the independently authored field regressions installed unchanged after formatting.
Each gate records selected input identities, commands, raw logs and actual exits.
The final portable regression uses semantic validation phases instead of an
exact callback count. This test-only correction followed the native and consumer
runs; their compiled production inputs remain unchanged. The ledger records that
selected dependency test-file difference explicitly.

The [consumer gate](consumer/consumer-v2/qualification.json) passes controller
race20, its complete routed offscreen race suite, both vet modes and the desktop
race build. The [Cocoa gate](consumer/desktop-v2/qualification.json) passes
fixed- and variable-height million-row processes with exit 0, all required stages,
no race warnings and unchanged selected sources/binary. The runs took 34.612s
and 50.797s respectively; these are workflow durations, not frame-latency
benchmarks. The handoff retains raw logs byte-exact inside `raw-evidence.json`.
The [qualification ledger](qualification.json) binds the final records and
current selected sources.

Independent source reviews cover the portable builder/admission, consumer,
ordinary-field adapter and contract documentation. Concrete corrections include
late/empty input-stamp drift, full-domain revision advancement for filtered edits,
held-save external-update refusal, and duplicate unchanged-field settlement.
The first actual Cocoa process reached its workflow stages but exited 66 with
a cursor-owner race during concurrent window closure. Patch 0033 moves that
association into each main scene under its sprite lock; fresh preparation and
independent actual-source red/fixed-source green tests preserve that correction.
Historical actual regressions, counterfactual cases and reviewer fixture errors
are distinguished in the preserved review archives.

One [Prism review](../../reviews/macos-grid-loaded/prism.json) used Anthropic
Claude Sonnet 5.5 and returned one high and three medium findings (process exit 1).
The [adjudication](../../reviews/macos-grid-loaded/findings.json) refutes three
with source checks and focused race20 regressions: the token initializer is valid
in the actual source, prepared edits register no pending request, and user
read-only cells reject editing before worker admission while explicit external
domain updates remain supported. The fourth finding was fixed in the test fixture
and independently verified against the missing-check counterfactual. The raw
Prism result is retained; the final resolution is a separate internal review.

## Scope

One million records contain 160,000,000 logical value bytes. A numeric-sort view
uses 16,000,000 logical index bytes; cached strings and scratch arrays are charged
separately by the builder. Borrowed domain arrays, allocator overhead, multiple
live revisions and native/GPU memory are excluded from these logical budgets.
The consumer makes a full immutable domain copy for an edit; these measurements
are not total process memory or a peak-memory bound.

Manual tests use supplied fonts/theme, explicit worker/owner steps and actual
managed pointer/key routing. Desktop commands run on the owner and exercise real
Cocoa frames and native settlement. These checks do not establish physical input,
human IME/VoiceOver, external AX, sustained p95/p99 latency or complete P6/P9
acceptance. Large-tree workloads, sustained memory/latency and broader Mac
acceptance remain open. No OS preferences changed. Namespace and MIT licensing
are already selected; public distribution is still separately gated.

`native-v1` predates ordinary-field settlement; `native-v2` predates the
cursor correction. Both are historical compatibility evidence. The initial
consumer CPU proof likewise qualifies the earlier prepared dependency; its
Cocoa failure is retained as a failure, regardless of the preceding workflow
pass marker. The builder-focused gate also predates installation of the two
portable reviewer test files; the final core gate includes those installed bytes.
`core-v1` preserves the prior exact-count test fixture; `core-v2` qualifies the
installed semantic-phase fixture. No unchanged production was sent for a second
Prism review.
