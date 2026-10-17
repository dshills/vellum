# Key-event publication ownership and native imaging

The key-publication ownership regression and fresh native imaging case pass.
One Prism checkpoint using Anthropic Claude Sonnet 5.5 covered all thirteen
source/test/spec files (22,910 diff bytes), with one fresh provider call and
complete coverage. Prism retained one medium finding, and its CLI exit 1 remains
preserved. An independent review refuted that finding for the actual CLI path:
`main` starts its 60-second watchdog before `run`, and cannot stop it until the
worker join returns. The owned native controller separately enforces 75 seconds.
A timeout that returned before joining would recreate the fixture lifetime bug.
No source change or second provider call was required. Fatal watchdog exits count
as failures, and no injected hung-worker/watchdog execution is claimed.

The first race-enabled imaging smoke on modified base `e96f642` exited 66 with
a Go data race after 4.636 seconds. It produced no success summary. The retained
trace shows `Source.Key` copying the already-published key's `Base` through a
value-receiver modifier method while render dispatch calls `SetLocalOff` on
that same event. This is a queue publication ownership bug, not a demonstrated
race in `Source.Last`. The owned process group 50388 was recorded absent after
retirement. That original native failure remains failed.

Owned stage 0077 derives the remaining chord decisions from the original key
code/modifier arguments. It removes post-publication event reads and preserves
key-before-chord ordering, independent event identity, keypad rune mapping,
NumLock, modifiers, producer state and the injected clock. No added lock,
polling timer, event clone or general producer-serialization claim is involved.
The manifest binds exact input, patch and two output hashes. The ordinary patch
order regression now includes stage 0077.

The new producer/consumer test acknowledges 128 ordinary keys, lets the consumer
perform normal local-offset/handled mutations, and joins that consumer. Ordinary
keys produce no second chord-send lock that could conceal the old ownership
violation. A separate twelve-case matrix checks chord behavior and timestamps.
The old-source-only overlay is a deliberate negative control and must retain its
expected race failure separately from corrected test successes.

The same checkpoint retains the imaging fixture's explicit worker completion
join and process-local geometry isolation. Work-context cleanup precedes result
publication, application-close callback precedes completion, and every post-launch
return joins before restoring fixture state. A blocked-cleanup regression covers
both successful and failing work. Normal interactive behavior stays on its
existing path. The first source review's worker-lifetime finding remains retained.

Tests reconstruct both pinned dependency modules into a fresh private directory
and require exactly the original 6,249-file vector with one replaced source and one
added test: 6,250 files. Subsequent guards check the complete file set
and hashes. The earlier unadmitted runner draft lacked a file-set check; that
review finding and its correction are retained. No build ran under that draft.
The checkout prepared trees remain unchanged during testing. After commit, their
preparation and equality with the tested vector will receive a separate receipt.

The retained clean `3da753e` SDK supplies only the pinned private static GPU
capsule for this check. Its 7,027 baseline source/prepared counterparts are
compared with the clean baseline; that number does not claim the new stage is
unchanged. Actual consumers use current application/driver source and the fresh
stage-0077 modules. The new executable is built from the recorded modified base;
it is not relabeled a clean build from the later evidence commit. No SDK refresh
or installed/public GPU library adoption occurs here.


The corrected continuation completes twelve zero-exit commands. The entire
workbench package passes under IP-network denial with Unix IPC allowed; full
module vet and imaging race/vet/build pass. The earlier full-driver command
still failed. Its ledger contains 13 `ok` outcomes (one reports no tests to run),
seven packages with no test files, and the one failed workbench package. Reuse
of those completed component outcomes and the successful whole-workbench rerun
provides package-scoped evidence; it does not relabel the old command a pass.
Its five Unix-bind errors from three workbench roots remain in the archive.

| Check | Actual evidence |
| --- | --- |
| Dependency reconstruction | Exact 6,250-file output from all pinned stages |
| Old-source control | Expected Go exit 1; race at `Source.Key` line 100 against consumer `SetLocalOff` |
| Corrected events/key | Both race suites and vet pass |
| Preparation suite | All eleven tests pass |
| Workbench rerun | Whole package race passes in 10.138 s with local Unix IPC |
| Current executable | Go 1.27.2, race-enabled ARM64 `MH_EXECUTE`, minimum macOS 13.0 |
| Current static GPU linkage | 141 exact source-built archive members; no WGPU dylib dependency |
| Fresh native imaging | Exit 0 in 6.991 s; group 75299 retired and recorded absent |
| Native streams | 4,044-byte one-summary stdout; empty stderr |

The current executable is 68,806,914 bytes with SHA-256
`4fcdf59bcfbe85f39f89c93f25c5364144fd6b35b426fd54f8c32467218b941e`.
Its 13,139,015-byte linker map has SHA-256
`30a858a53f5abd756a188a415fa5310353d3490728eaa4f9414ea64b8da38da7`.
The selected package graph contains 487 imaging packages; the separate workbench
test graph contains 497. Those counts describe selected packages, not test counts.

The fresh native summary has SHA-256
`956bd461d15446de43c9219c7f306250ca83f98b5838742e013505343c00d32c`.
Exposure 1 and 2 reference pixels are respectively RGBA `(213,156,0,128)` and
`(231,188,0,128)`, with the source's pixel probes passing. Initial/chrome state
stays at 792×537, frame/preparation/draw counts 2, target generation 2 and two
buffer allocations. Storage plus uniform accounting is 1,048,608 bytes and the
float texture is 524,288 bytes. These logical resource values do not establish
physical native/GPU memory bounds. Delivery after resize was actually observed
as true; the fixture accepts either correctly preserved ordering. Canceled/closed
capture callback count is zero. All six recorded normal-close predicates pass.
The fixture observes geometry isolation enabled; restoration ordering is covered
by source review and worker regressions, not a separate post-return scalar probe.

The native imaging source asserts quantitative compute-to-RGBA16Float-to-sRGB
readbacks, partial alpha/transparent pixels, admission refusals, synchronous and
asynchronous capture, recursive-capture refusal, old capture metadata across
resize, cancel/rebuild cleanup, independent windows, minimize/restore, forced
native closure, survivor capture and final application closure. The resize
delivery overlap is the actual observed boolean; either scheduling order is
valid. These are automated source assertions and scalar summaries. A successful
new run does not establish which physical keyboard events occurred in that run.

All actual command streams, ownership records, selected source/module/linkage
metadata, scripts and internal/provider reviews are indexed with the source
vectors in [the evidence manifest](evidence.manifest.json) and
[evidence archive](evidence.tar.xz). The [result record](result.json) preserves
scalar native summaries and the separate run statuses. Principal native binaries, private SDK and static archive are separately
fingerprinted and copied to a durable local checkpoint after commit. The same
eight summary-control cases were reexecuted for revised runners; they are not
sixteen new cases. The earlier 27 pure guard/finalization controls are reused
support, not new process executions.

Physical pointer/keyboard workflows, compositor/full-image presentation,
ordinary frame latency, private snapshot decoder joining, native/GPU memory,
endurance, full IME/human accessibility, cross-display behavior, actual native
save-chooser interaction, component/rights closure, public distribution and full
Mac acceptance remain open. No OS input source, VoiceOver, clipboard or preference
changes are made by the fixture. Process-only geometry state is restored after
normal joined teardown; process-fatal race/watchdog exits are retained failures.
