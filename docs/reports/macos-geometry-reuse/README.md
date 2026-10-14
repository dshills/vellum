# Ordinary native accessibility geometry reuse

The adapter skips an identical ordinary geometry payload when the same native
bridge still has the same ordinary identifiers. Sender-oracle and race checks
pass, as do the stored 400-cell Cocoa grid consumer and general native
accessibility workflow. No ordinary native latency improvement is established by
this checkpoint; the existing 4 ms p95 / 6 ms p99 acceptance targets remain open.

## Change and ownership

The preceding endpoint is `59055e6be4bce2b4fd4f19462f4bd58e49c2d87e` with 6,216
prepared dependency files. Three adapter source/test files change. One bridge
retains at most 64 KiB of last encoded geometry, plus its current ordinary ID set.
An opaque version changes whenever that set changes. Reuse compares exact payload,
version and token. Larger projections and bounds containing any currently unknown
ID always publish and clear the reusable payload. The limit is a reuse fallback;
it never refuses valid larger geometry.

The cache is shared across main and modal render hosts. A host cannot reuse an
older value after another host overwrote the native geometry map. Removing and
restoring a same-ID control between renders, or moving it between ordinary and
scoped semantics, forces another publication. Ordinary native elements already
retain ID identity and geometry across role updates and ordinary reparenting.
Semantic enqueue/invalidation and geometry enqueue/cache commit now share the
bridge mutex. The native producers copy bytes and queue their main-thread work;
they do not synchronously call the owner while that mutex is held. Retirement
clears both membership and payload. Text-range geometry and scoped grid/tree
transactions retain their existing publication paths.

## Source-guided diagnosis

Before the change, two separate diagnostic processes use the stage-62 rendering
endpoint `87e7de1` and binary SHA-256
`1087196588699b419f1a10ecb5395ea52473d9faf84c7c428220572ad630246d`.
Each captures a process-wide CPU profile during 512 measured scroll events after
128 warm callbacks. Default CPU and opt-in GPU mode have identical frame geometry
with 400 clipped cells; retained pool cardinality ranges from 468 to 540.

The CPU profile samples 8.15 seconds, with 0.58 seconds cumulative attribution
through ordinary geometry publication. The GPU profile samples 7.94 seconds,
with 0.83 seconds through that chain. These are overlapping process-wide sampled
attributions, not per-frame elapsed costs or predicted savings. The main-loop
`WaitEventsTimeout` CGO samples and verification/runtime work are separate from
frame work. Profiles run concurrently with the composed endurance campaign and
do not qualify ordinary latency, physical GPU work, presentation or native memory.
The source/harness and completed diagnostic evidence receive independent internal
reviews. A separate post-terminal PID/group observation finds no remaining member
of the diagnostic's owned process group.

## Verification

Seven code gates pass: dependency preparation, focused geometry race tests, the
complete offscreen adapter race suite, adapter vet, candidate and old-behavior
native loaded-grid builds, and the native general-accessibility build. The full
adapter race suite takes 92.54 seconds. Two local overlays fail the intended
assertions: always sending fails unchanged-frame reuse, and a stale membership
version fails semantic remove/readd. The overlays never alter production source.

The existing stored-grid Cocoa consumer passes on one million stored rows with a
minimum of 400 clipped cells. Its report records 2,268 overlap reuse observations,
459 replacement retirements, admitted native press and refused attachment query
after closure. These counts retain the consumer's meanings; they are not distinct
lifetime object totals. Its stderr also records `Cannot query content scale without
screen`; the exit is zero and all consumer assertions pass. The warning's cause is
not established here and its raw bytes are preserved.

The general Cocoa accessibility consumer passes with 27 reachable elements in its
final main-window snapshot. It checks roles, values, native hit testing, controlled
actions, disabled refusal, focus/Tab order, selected-page visibility, bounded
virtual paging, independent tree selection/disclosure, modal focus/Escape and
workbench namespaces/geometry/retirement. This is native selector and queued-event
acceptance, not physical input or human VoiceOver.

Both native wrappers hold the shared live-test lock, check frozen source/prepared/
input/binary identities immediately before launch and after exit, enforce bounded
output/timeouts, and kill/reap the owned process group on every exit. Their source
binding is the preceding commit plus exactly three changed files. Binary digests
were captured after successful compile receipts and before native launch; the
compile wrapper itself did not capture hashes at build completion. No clipboard,
input sources, VoiceOver or OS preferences change. Owned process-group cleanup
is separate from full native/GPU resource or memory-budget qualification.

## Evidence and remaining acceptance

Internal source/harness review approves the exact three files and corrected
post-lock launch checks. Commit requires final internal evidence review and one
Prism review using Anthropic Claude Sonnet 5.5. The [qualification](records/qualification.json),
[archive](records/evidence.tar.xz) and [member index](records/evidence-index.json)
will bind source versions, producers, commands, raw output, profiles and reviews
before commit.

The ordinary stage-62 comparison is still queued behind the independently frozen
endurance campaign. This geometry change does not inherit those endpoints' results.
Matched native measurements for this change remain separate. Complete P4/P6/P11,
sustained memory/native GPU budgets, overnight soak, physical input/displays,
human accessibility, service interoperability and distribution remain open.
