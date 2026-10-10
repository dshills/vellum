# macOS native dispatch and termination

Owned patch22 makes native callback completion explicit. A failing admitted
callback stops admission, acknowledges failure and preserves its exact panic on
both loop and dispatch caller. Startup-inline callbacks share that policy; Goexit
releases the caller with an incomplete-callback failure. Accepted normal callbacks
still finish their output after concurrent stop. Terminal core recovery logs and
rethrows without trying to create an unreachable crash window.

Desktop initialization, empty event posts and termination use a process-wide
barrier. All five desktop input posts now pass through it. It holds no lock over
polling, callbacks, queue waits or application locks. Existing standalone GPU
example Init/Terminate helpers and unrelated GLFW operations remain outside the
barrier's contract.

Fresh dependency preparation and the uncached scripts/check.sh pass: 87 Python
regressions, portable/native race tests, vet, selected backend tests, native
workbench build and independent consumers. Twenty additional race repetitions
exercise failure, shutdown, event lifecycle and terminal recovery regressions.
Internal design review identified the terminal crash-dialog deadlock; the
implemented guard and regression resolve it. Subsequent code and integration
reviews found no actionable defect.

Both real macOS queue probes pass separately. Normal mode completes 100 callbacks;
panic mode starts one failing callback and preserves exact identity on both
threads. Each submits 11,000 explicit wake requests across and after actual GLFW
termination, then confirms callback/native-creation refusal. This number counts
requests, not native posts; dispatch also requests wakes. These probes open no
windows. A separate race-enabled last-window handoff and a four-second 100,000-row
workbench/GPU readback check pass and finish native shutdown. Commands, exact
source identity, hashes, exit statuses and raw logs are retained in
[source-and-check-proof.json](source-and-check-proof.json).

This checkpoint establishes synchronous failure and desktop wake serialization,
not recovery of all GPU/AX/native resources after arbitrary panics. No physical
Finder/IME/screen-reader, sustained performance or public distribution gate is
qualified here. OS input-source and VoiceOver settings were unchanged; bounded
windowed probes closed. Document persistence/unsaved-close workflows and the
other [macOS completion work](../../../specs/macos-completion.md) remain open.
