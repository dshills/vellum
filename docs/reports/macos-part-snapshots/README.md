# Immutable component snapshots in the composed workbench

Starting calculation or sorting previously cloned every component on the UI
owner. The workbench now stores scalar components in immutable three-level pages.
A snapshot shares the root and copies its descriptor/revision. A property edit,
undo or redo copies one root, one branch and one 256-value leaf before publishing;
on this arm64 host those arrays total 3,072 bytes. No old snapshot node is mutated.
The source stays confined to its owner; detached snapshots support worker reads.

Storage addresses at most 1,048,576 records; existing application/document limits
remain 100k..1m. Base storage and snapshots retain arrays for their actual records.
Repeated replacement does not accumulate an internal history of roots; external
workers can retain their own old snapshots. This is not a total process-memory
bound. The existing history stores sparse scalar deltas, not complete roots.

Persistence captures the shared component snapshot on the owner. Encoding expands
records and validates/serializes on the document worker. Decoding validates and
builds the page store on the worker; admission transfers its root in constant
time. Document v1 JSON remains byte-compatible. Geometry/reference capture still
copies its slices, and startup still constructs initial records before the UI
opens. Legacy table projection admission still scans/copies its result on the
owner. This change does not remove every document-sized owner operation.

## Correctness and local cost

Tests compare seeded updates with an independent flat oracle, page/branch
boundaries, historical snapshots, concurrent frozen readers during edits,
undo/redo, caller-slice detachment, exact v1 JSON and worker-built load admission.
Existing inspector/domain/semantic/document tests remain in the selected suite.
The initial sandboxed run failed only at the existing Unix socket; its receipt is
retained. The authorized count-three run passed. Final adapter race/vet checks
qualify the current source; earlier source vectors remain explicitly distinct.

Five sequential local benchmark repeats compare the original flat-copy operation
with shared snapshots. Medians:

| Operation | Time | Bytes / allocations |
| --- | --- | --- |
| 100k original snapshot | 33,230 ns | 802,850 / 2 |
| 100k shared snapshot | 7.713 ns | 24 / 1 |
| 1m original snapshot | 142,389 ns | 8,003,618 / 2 |
| 1m shared snapshot | 6.488 ns | 24 / 1 |
| 1m edited path | 310.6 ns | 3,072 / 3 |
| 1m original flat traversal | 296,821 ns | 0 / 0 |
| 1m paged traversal | 919,944 ns | 0 / 0 |

Full worker traversal is roughly 3.1 times slower in this local summation fixture;
its absolute increase is about 0.62 ms. The tradeoff removes large owner-side
snapshot allocations while moving indexed lookup work to workers. This is a
local operation comparison, not frame-rate, startup, GC or process-memory
qualification. No matched whole-frame campaign is claimed for this checkpoint.

## Native composition checks

The first native document run caught a stale readiness target, inspector.rotate,
left by the earlier inspector replacement. It is retained as a 90-second failure.
The fixture now waits for the current inspector.value field. The full 1m-record,
100k-outline run then passes actual Cocoa CloseReq/Cancel, asynchronous v1 save/
load with complete geometry/reference equality, failed Save preserving the dirty
window and successful Save closing only after acknowledgment. Full diagnostic
document comparisons run on the fixture worker.

A separate native history/canvas/calculation smoke commits revision 7 with sum
597,995,654, verified by an independent arithmetic formula for the stored values
and accepted edit. A ten-second profiled native editing preflight completes 40
steps, 10 confirmed Cocoa edits and 30 joined workers (10 fresh, 20 stale refused).
These are bounded composition checks, not the complete ten-minute/1,000-cycle/
overnight P11 workload, primary 1,000-control geometry or physical human input.
The original P6 frame targets and full Mac/release gates remain open.

## Review and records

The checkpoint requires internal review and one bounded Prism review using
Anthropic Claude Sonnet 5.5. Raw producers/receipts, source vectors, all benchmark
observations, native output and the short diagnostic trace/heap will be archived
with a verified SHA256 member index. The earlier failed/superseded receipts are
preserved separately from selected final passes.
