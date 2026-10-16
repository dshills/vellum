# Native accessibility parent graph checkpoint

Mac snapshot validation now proves parent completeness and acyclicity once per
snapshot, using two local sets, instead of allocating an ancestor set and
rewalking each node's chain. It keeps all existing field/scope admission rules
and the preflight-before-mutation transaction boundary. See the
[contract](../../../specs/accessibility.md#native-parent-graph-validation).

## Correctness and lifecycle

The new Darwin test uses an independent Go oracle with the original separate
ancestor sets and literal host membership. Actual baseline and revised Foundation
validation agree on 1,441 exhaustive parent graphs (146 accepted/1,295 refused)
and 15,625 exhaustive parent/scope graphs (83 accepted/15,542 refused). Other
fields are fixed valid values; existing envelope/integer suites continue to
cover those validators. Corpus SHA-256 values are respectively
`bc97bdddad84841f51f2a955c1a299877d3cc42aada8d23e3752b398b0e0bfcb`
and `f36790b1b98954df3ee4ab753e7c2a5b339e81f0b1891e4595591800de71407b`.

Seeded random and 1,024-node deep trees pass four input-order shuffles each;
missing parents, cycles and a scoped non-host with a host-like role are checked.
Five actual higher-generation refused native transactions preserve the previous
published generation and exposed element records, including identity, geometry,
value, focus and liveness. These observations do not directly inspect every
private text-cache/hierarchy field; the unchanged transaction ordering provides
the broader source argument. All traversal references remain invocation-local.

Deliberate guard omissions prove the oracle detects admission defects: the first
candidate's missing whole-graph and same-scope-parent controls compile but fail
the intended assertion; the final candidate's missing unresolved-parent refusal
also compiles and fails that assertion. Exact candidate bytes are restored after
each control. Full current-source offscreen `go test -race -p=1 -count=1
-tags=offscreen ./...` and `go vet -p=1 -tags=offscreen ./...` pass in the adapter
module. Native/dependency deprecation and duplicate-linker warnings are retained.

## Matched local performance

Two unchanged test bodies are compiled as optimized baseline/final executables
from base `cebae61`, with the identical 6,229 prepared dependency leaves. Each of
five alternating AB/BA pairs launches fresh sequential processes, explicitly uses
two Go benchmark CPUs, and takes 50 iterations per case. Medians and full observed
process-value ranges below are milliseconds per operation. No heavy test job or
owned native UI overlaps these selected runs; interleaving does not establish
continuous global host isolation.

| Fixture | Original median [min, max] | Revised median [min, max] |
| --- | ---: | ---: |
| 451-node grid topology | 1.939758 [1.927129, 2.052976] | 1.861528 [1.844302, 1.891862] |
| deep64 | 0.343221 [0.337753, 0.353624] | 0.111831 [0.104588, 0.113390] |
| deep256 | 4.278035 [4.268117, 4.357920] | 0.440938 [0.428776, 0.474202] |
| deep1024 | 313.541682 [312.396281, 314.071542] | 2.552770 [2.472002, 2.624668] |
| wide1024 | 2.200920 [2.154934, 2.255222] | 2.169899 [2.102478, 2.195002] |

Each operation includes CGo entry, Foundation JSON parsing and complete native
snapshot validation. It excludes native commit, rendering and a real window.
The 451-node fixture is a representative topology, not a captured render
snapshot. The wide-case ranges overlap; its small median difference is not a
material performance claim. Deep-chain savings address the repeated traversal
visible in source. Go `B/op` and allocation counts exclude native/Foundation
allocations and cannot establish a native-memory budget.

The first candidate passed correctness but regressed wide1024 from 2.203794 to
3.063945 ms median (~39%); it was rejected. Its full runs, code, reviews and
controls remain retained. The final design skips traversal state for leaves.
A prior baseline pilot exited zero but its producer refused a CPU suffix of 14
instead of the expected 2. That failed producer run is preserved and excluded;
selected campaigns explicitly set `-test.cpu=2`. The suffix's original cause is
unassigned.

## Live scope and remaining gates

One existing `cmd/accessibility-smoke` executable built from the final native
source runs successfully in a real Cocoa application. It queries the reachable
native tree, roles/labels/values/hit testing, controlled actions and disabled
refusal, keyboard routing/focus, virtual table/outline paging/disclosure,
modal isolation and the live workbench bridge. Keyboard events are injected
through the toolkit; this is automated native routing, not physical input or
human VoiceOver acceptance. The controller observes a terminal zero exit and
owned process-group absence, exact source/binary/log stability, the shared native
lease and human-return markers. It does not inventory all OS native resources.
The fixture requests no changes to OS input sources, VoiceOver, the clipboard
or preferences and reports `preferences_changed=false`. Global OS state is
outside this observation.

The baseline build/test and final build/test/race/vet/native build/run pin their
actual source delta, baseline manifest, command, terminal status and raw logs.
The 4,485 tracked baseline files and 6,229 prepared dependency leaves are checked
before/after code checks and the native run. The final code endpoint matches the
reviewed bytes; subsequent changes are documentation/evidence only. Earlier
candidate results are historical and explicitly separate.

Snapshot validation is outside the exact `FrameMetrics.TotalUIWork` interval.
These measurements do not establish a whole-frame or ordinary P6 improvement;
the 4 ms p95/6 ms p99 targets remain failed. Current-source endurance, a successful
overnight soak, native/GPU allocation and app-wide wake/present budgets, physical
pointer/display checks, human IME/VoiceOver, interoperability and distribution
remain open. The earlier failed overnight result stays failed. Full Mac and
public release acceptance are not granted.

Checkpoint acceptance requires the raw evidence, producers/dependencies,
original and intermediate source, checks, failed pilot, both benchmark campaigns
and internal/provider reviews in `records/evidence.tar.xz`, indexed by
`records/archive-index.json`. Private executable identities remain in receipts;
executable files are excluded from that archive. `records/qualification.json`
must bind the scoped checkpoint to its exact reviewed sources and archive.
