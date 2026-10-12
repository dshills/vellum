# Mac modal geometry and native checks

Closing a label-only modal could restore its background accessibility elements
with zero-size frames. Escape did dismiss the modal; the cached main scene did
not render again after native modal publication pruned the old background frames.
The driver now requests one main-host render when the visible accessibility scope
changes. Unchanged scopes keep demand painting. Removed elements remain retired.

The Cocoa file-drop regression requires positive background frames after Escape
and controlled closure, then routes another drop through the restored frame. It
moves no physical pointer. The test also retains the last inspection on failure
and bounds the worker shutdown join with a five-second grace period.

## Verification and review

The final six source/contract files have internal approval. Fresh full native
offscreen race and vet suites pass; outcome tests pass twenty race repetitions.
Final native modal/drop, broader accessibility and demand-rendering checks pass.
The initial root/offscreen attempts were denied temporary Unix-socket binding in
the sandbox; authorized retries pass. Baseline main-queue normal/panic, rendering,
platform, multiwindow, lifetime, history and docking receipts retain their original
5a1f7f9 identity; intermediate fix receipts remain historical versions.

Prism ran once with Anthropic Claude Sonnet 5.5 on five original files, complete
coverage, 5,854 input and 1,329 output tokens. Its raw exit 1 and one medium finding
are preserved: the fixture join could wait indefinitely. The bounded join and its
new test file were internally reviewed and tested afterward. The final six-file
approval and explicit disposition bind that correction; no second provider run
was made. Full Prism coverage refers to its original five-file input.

A separate trusted, PID/executable/window-bound AX inspector answered the two
idle native windows in 0.092 and 0.048 seconds. Its dump route reads a partial tree,
performs no activation or input, and has an eight-second watchdog. The no-drag
fixture then closed normally with its expected deadline failure. This narrows but
does not diagnose two earlier CUA app-selection calls lasting 3,885 and 9,262
seconds. Those physical-drag attempts timed out without issuing a drag. The test
Finder window was closed, returning to Desktop. Physical Finder interaction
remains unqualified. The external receipt producer's pending-wrapper-stdout hash
was removed; original and corrected receipts are retained, with query data unchanged.

## Performance observations

The authorized fifteen-hour window began 2026-10-12 19:07:09 UTC and ends
2026-10-13 10:07:09 UTC. This host is an Apple M4 Pro Mac with 64 GiB RAM and one
5120×1440 60 Hz display. Background processes remain observable; user absence is
not proof of an idle OS. No user processes or power/display settings were changed.

Both renderers completed five runs for each fixed/variable-height 10k/million-row
case, 120 measured events after 128 warm events. All forty runs passed record and
matched-geometry validation, including the 400-cell workload. Timing targets
remain total UI work p95 ≤4 ms and p99 ≤6 ms. Both campaigns fail those targets.
These are elapsed same-ticket UI regions, not physical input-to-photon timing.

| Stored rows/heights | CPU p95 / p99 ms | GPU p95 / p99 ms |
|---|---:|---:|
| fixed-10000 | 10.167 / 10.971 | 10.076 / 10.992 |
| fixed-1000000 | 10.085 / 10.480 | 9.856 / 10.233 |
| variable-10000 | 10.286 / 11.219 | 10.606 / 11.241 |
| variable-1000000 | 10.157 / 10.435 | 10.112 / 10.393 |

Three small hello-fixture idle runs sampled about 63 seconds each and observed
one-core CPU percentages 0.270%, 0.286%, 0.286%.
They meet that fixture's 0.5% target; they do not qualify the integrated
1,000-control workbench, kernel wakeups or endurance/native/GPU memory.

## Records and remaining gates

[Checkpoint](records/checkpoint.json) binds the [archive](records/evidence.tar.gz)
and its per-member [index](records/evidence-index.json). Qualification binds all
6,174 prepared dependency files and the final source hashes. Raw benchmark
machine records retain serial fields locally; archived machine.redacted.json
derivatives remove serial/UUID values and retain the original digest.
No general clipboard write, IME/VoiceOver toggle or OS preference change occurred.
Earlier deferred evidence remains unchanged.

Rich clipboard interoperation, tray interaction/teardown, notification permission
and presentation, physical dragging, human IME/VoiceOver, multiple displays,
physical GPU loss, full P4/P6/P9/P11 performance/endurance and distribution rights
remain separate gates. This checkpoint fixes the reproduced modal defect and
records bounded checks; it does not complete the Mac UI framework.
