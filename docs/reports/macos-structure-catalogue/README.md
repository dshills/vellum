# Controlled tab keyboard navigation

Managed Tabs now handle unmodified Left/Right, Home/End and Space/Enter through
public controlled selection callbacks. Arrows wrap and skip disabled options;
Home/End choose enabled endpoints. Native header focus changes at the originating
event boundary; owner delivery proposes stable IDs and preserves rejected state.
Read-only input moves focus without mutations, selected activation is a no-op,
and existing pointer repeats and Delete/Backspace close behavior remain supported.

The queue captures absolute movement intent before owner delivery: rapid A→B→A
must produce B,A even before reconciliation. One bounded pending intent also
preserves explicit Space/Enter after focus-only Shift-Tab returns to A. It stays
effective through owner delivery until originating authoritative reconciliation,
with exact-pointer refusal rollback and lifetime retirement. A header that becomes
disabled yields focus to the selected/first enabled header without a selection
callback. With no eligible header, focus traversal skips the complete strip subtree
(including private Parts and still-unsynchronized child widgets), choosing the
next enabled control or clearing focus when none exists. Input-time write permission cannot
revive later. Origin/target binding epochs retire on hiding, disabling, removal,
closed dialog scope and nil/non-nil Change transitions. Owner delivery uses owner
values and atomic lifetimes, with no native traversal/focus from another window.

Final complete driver-root (`.`) and gallery race suites, focused tab/lifecycle
regressions, root/gallery vet and native build pass. The public Manual consumer
checks shared ordinary Disclosure/SplitPane/document constructors, retained child
incarnations, actual keys and pointer routing, exact divider geometry, scrolling
endpoints and controlled document close/stale refusal. The old burst implementation
reproduces B-only; both read-only transition cases reproduce an unwanted callback.
The fixed implementation passes. Initial absent-arrow, test-event cloning and
missing-overlay build failures remain explicit in the raw records and versions.
One positive and 34 negative synthetic oracle controls pass, separately from
native evidence. No new public API signature or catalogue kind is added.

The new runnable `gallery -structure-native-check` authors 50 inputs and 42
observations. Its **first native run FAILED** after 41 acknowledged inputs and
35 source-asserted observations ending at `page-documents`. Both subsequent
`document-notes` pointer packets are acknowledged, but the next Frame fails with
`composition did not submit a measured frame` before Snapshot/model observation.
No notes-selection result or failed snapshot is inferred. The planned original-
window close was not reached; failure cleanup closes the session after 194 issued
requests, with no pending
request/live window and confirms owned process group32943 absent. A separate
post-failure check proves complete source/prepared/SDK/binary boundaries unchanged
before the later review fixes;
it is not the controller's unexecuted post-validation boundary. The scope is sealed.

The measured GPU path starts Supported=true; a nil render pass with nil error
can leave Submitted=false and no specific failure, producing this generic trace
error. Actual acquisition status/attempt count is not retained. Occlusion,
timeout, exhausted lost/outdated recovery and null-handle cases cannot be
identified from this record. The next implementation action is a bounded measured
acquisition outcome/attempt diagnostic with injected unavailable/lost/null-handle
regressions. Preserve ordinary no-frame behavior, retry count and strict Frame
success criteria; do not convert this failure into a pass or retry the sealed case.

The pre-admitted independent oracle is also too strong about expanded disabled
geometry: section-after Y is250 normally,236 disabled and138 collapsed. A later
prefix validation stops at this original equality guard; no independent35-state
pass is claimed. Review confirms that disabled button shadow margins and text-
field borders change these metrics: the heading, draft and action heights shrink
by 6+2+6px, accounting for the 14px shift. This establishes a false cross-style
oracle assumption, with no new product layout defect. Keep that oracle unchanged
with this failed record. A future scope must compare restoration within the same
effective style or independently cover state-specific metrics. The compiled consumer's
accepted state, visibility/focus and divider guards are unchanged.

The first Prism Anthropic Claude Sonnet 5.5 review covers all 20 staged files
(89,210 bytes) in one complete uncached call. Its raw CLI exits 1 with two medium
findings. Both identify valid gaps despite partly overstated examples: arrow-only
A→B→A already worked, but actual Core Right/Shift-Tab/Space and Enter reproduce a
lost activation; disabled-header arrows stall although Tab can escape and read-
only focus movement is intentional. Both failures are preserved. The fixes above
pass those actual Core queue/manager routes, plus pending delivery/publication,
queue refusal and hidden/disposed cleanup regressions. An internal follow-up
also reproduces and fixes focus escape into still-enabled child descriptors when
the entire strip becomes disabled. The complete final driver and gallery race
suites, focused tab tests, root/gallery vet and native build pass after these fixes.

The original native executable/source, failed run, oracle and first review stay
immutable. The native prefix describes that original source, **not qualification
of the subsequent review fixes**. No new native run occurs. Full 50/42 qualification,
notes selection/close and complete independent geometry validation remain pending.
Internal fix and pre-provider reviews approve the corrected source and software
evidence. The second complete uncached Sonnet 5.5 call covers all 20 corrected
files (107,458 bytes), with no excluded/skipped/truncated input or discarded/
suppressed findings. Its raw CLI exits 1 with two retained medium concerns.
Independent review refutes them: repair first checks actual focused membership
and runs within the originating scene, preserving valid/unrelated focus without
OS window activation; production control retirement clears pending intent and
epochs, while closed window lifetimes refuse subsequent input. Disposal and
pending/publication regressions pass. Both raw comments remain preserved; no
source change or third provider call is needed. This factual README-only delta
receives final internal archive/commit review. The bounded
[archive](records/checkpoint.tar.gz) and [manifest](records/manifest.json) retain
both provider payloads, original source, failed tests, final checks and raw native
evidence without binaries or SDK copies.
This is a tested keyboard product checkpoint. Full M2/M5 and M3–M7 remain open.

Evidence uses Go1.25.6 ARM64/GOMAXPROCS2 CPU-scene over retained source-built native
archive786a only. Both checkouts keep identical stage84 prepared files; all7880
retained SDK files are unchanged. The overlay matches all current non-prototype
Go source and prepared bytes. The SDK's native library module override is scoped
and is not qualification of the checkout's default cached library. No OS settings,
input sources, clipboard, SDK refresh or public release changes occur. Go FIFO
does not establish physical/OS input, native pixels, VoiceOver/IME, actual split
dragging, arbitrary resize/display transitions, presentation, performance or
native deallocation/endurance. Earlier full workspace records retain their pinned
sources; this tab-selection delta does not requalify that full record. Follow the
unchanged [Mac delivery plan](../../../specs/macos-delivery.md).


## Subsequent complete native record

The [d1232b5 structural/publication checkpoint](../macos-editor-publication/README.md)
now independently passes the entire unchanged50-input42-state structural record,
including rejected-close focus, following Delete and accepted removal. Original
failures above stay failed/sealed and describe their historical source only.
This is native Go FIFO routing; physical input and native pixels remain separate.
