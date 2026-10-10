# Exact text and bounded admission contracts

Recorded 9 October 2026. **Exact multiline export is now demonstrated through an explicit opt-in buffer. A bounded ordered mailbox is implemented and tested independently. G1 remains open and workbench expansion stays on hold.** These changes are isolated in [foundationcontracts](prototypes/foundationcontracts/README.md); earlier prototype sources and evidence remain unchanged. See [ADR 0006](decisions/0006-foundation-contracts.md) and the [manifest](prototypes/foundationcontracts/evidence/manifest.json).

## What changed

Patch 0007 adds `lines.NewLinesExact()`. The previous representation removed one final empty line during import, then legacy export added synthetic newlines. The new constructor opts its buffer into retaining the full line representation and exporting it without adding newlines. The policy lasts for the buffer lifetime, including later replacement, editing and Undo/Redo. Existing constructors keep their previous semantics. This solves exact valid-UTF-8 application-text import/export in the tested path; arbitrary byte preservation and explicit formatter transformations are outside the contract.

A further probe found that legacy line-region diff edits cannot remove a final empty line. Exact buffers now apply the supplied diff operations to a line snapshot, then commit a grouped full-text replacement; unselected changes remain untouched. The temporary file comparison buffer uses the same text policy. This trades incremental formatting/caret preservation and allocation efficiency for correctness in the prototype; complete file-open/revert workflows remain unverified under the existing lexer/render failures. The initial failing probe is retained.

The `mailbox` package reserves fixed-ring slots and payload credit before mutation, owns a payload copy at commit, delivers in reservation order, and holds credit through application acknowledgement. It permits one in-flight delivery. Nonblocking preflight reports capacity failure before acceptance; waiting submission/consumption is restricted to dedicated goroutines. Shutdown retains accepted messages for drain and cancels only uncommitted reservations. Tests cover concurrent producers, stale/reused tickets, size/slot limits, cancellation, close and acknowledgement ownership.

## Evidence

| Check | Result | Boundary |
|---|---|---|
| Exact import/export | Eight standalone cases preserve every byte; focused tests additionally cover 12 cases, repeated replacement, line-array input and empty input. Includes trailing/blank lines, Unicode and CRLF. | Valid UTF-8 buffer contract; no OS clipboard test |
| Edit history | Insert/Delete/Undo/Redo and actual editor newline/emoji backspace pass under race detection. An independent reference model verifies 800 deterministic edits and their Undo/Redo. | Logical editing; visual bidi/pointer/word navigation remains open |
| Exact diff/history | Five boundary pairs, selected changes and 100 full diff pairs with Undo/Redo pass under race detection. | Full snapshot replacement; incremental formatting/caret behavior and file-open/revert still need validation |
| Native fixture | The seven-patch derivative exports the full seven-script corpus with no added newline. All nine deterministic Cocoa protocol probes complete; ten updates are acknowledged, no actions rejected, final field `A中文漢👩Z`, adapter active; process closes normally. | Single-field protocol calls under U.S. input source; no live CJK conversion/panel or VoiceOver workflow |
| Bounded admission | Stalled-consumer control holds 128 messages/1,024 payload bytes. 10,000 further reservation attempts report full before hypothetical mutation. | Explicit preflight refusal; rejected attempts are not accepted actions |
| Lossless delivery | 100,000 worker messages delivered in FIFO order with zero accepted-message loss. Observed peak 128 outstanding slots, 1,024 payload bytes and one in-flight delivery; drain leaves zero slots/bytes. | Configured limits 128 slots, 4,096 queue-owned payload bytes, 32 bytes/message; caller/native/downstream memory not covered |
| Mailbox concurrency | Six race-enabled tests pass, including 2,000 messages from four concurrent producers. | Standalone mailbox, not native or toolkit event-pump integration |
| Resource control | The synthetic 100,000-message run takes about 0.04 seconds and allocates about 28.3 MiB total; forced-GC heap after drain has no positive delta in that run. | Single-run diagnostic; allocations are not zero, no application throughput/latency claim |
| Platform compilation | Mailbox test binaries cross-compile for Windows/Linux amd64 with CGO disabled. | Compile only; native platform matrix remains untested |
| Build/regression | Clean owned patch preparation, vet and focused exact/grapheme/cluster/lazy-spelling race checks pass. | Broad upstream markup panic/render-suite stall from the prior report remain unresolved |

## Remaining integration requirements

The mailbox does **not** replace the existing nativeinput event transport in this prototype. That native client changes Cocoa state before invoking its Go callback. Rejecting an enqueue afterward would discard input the platform has already accepted. Bounded native transport therefore remains an open gate: reserve before mutation, commit only after successful native changes, cancel on failure, serialize view shutdown, and provide a visible refusal or pause policy when capacity is unavailable. The toolkit must use a bounded/rearmed wakeup and bounded batches; moving messages from this ring into the old infinite deque would defeat the bound. Caret/progress snapshots require a separate bounded coalescing policy.

Queue-owned byte limits include reserved and in-flight payload lengths; fixed slot metadata and allocator rounding are additional. Consumer payload references are borrowed until acknowledgement. Application-held copies, native objects and tickets retained by callers are outside this budget. Abandoned reservations can block FIFO progress and must be canceled by the owner.

The native fixture still uses the existing system-font policy and has a large live heap; no font-memory improvement is claimed. Verified font discovery/coverage and the roughly 257 MiB isolated shaper result from [foundationchecks](FOUNDATION_CHECKS_RESULTS.md) remain relevant. Before-show/focus/detach seams, native/GPU retention, full visual bidi and native ranges, real candidate-panel placement, VoiceOver navigation/speech/invocation, Windows/Linux runtimes and trusted clean-machine distribution remain open. No OS setting changed, upstream message was posted, or release was published.

Next implement the native preflight/lifetime seam as a bounded fixture and rerun saturation and close-during-composition checks. In parallel with that engineering priority, investigate loaded-font/table ownership against the recorded profile. Continue D only if these contracts can fit maintainable public/upstream extensions; otherwise compare B on the same acceptance criteria.
