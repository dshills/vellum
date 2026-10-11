# Native compute, HDR and asynchronous capture acceptance

The public `GPUView.SnapshotAsync` copies the last completed frame into independent
bounded staging and a detached straight-alpha image. Its worker progresses while
minimized, delivers through the bounded owner mailbox, preserves capture-time
metadata across resize, and quietly drops canceled, rebuilt or closed work.
GPU staging participates in every resource-budget admission path; one reservation
survives remount and remains held through callback return or worker cleanup.

[Recorded checks](acceptance.json) distinguish initial evidence from the final
full gate on the corrected backend. The final gate passed 89 Python tooling checks and root,
native driver, patched backend, four example modules and SDK contract race tests,
vet and builds. Final independent consumer
race/vet/build checks and two corrected native runs cover the mailbox and terminal
closure changes. Initial source hashes and native runs remain explicitly historical.
Patch24 reproduces the checked encoder-only texture copy seam and its input/output
hashes. Snapshot regressions cover row padding and bounds, direct owned copying,
linear alpha decoding, old metadata delivery, callback reservation, stale jobs,
full-mailbox cancellation, undrained closure, decoder cancellation and native-close
control flow.

The [independent imaging evidence](../../../examples/imaging/acceptance/README.md)
contains initial and two corrected native race runs, exact source and binary
hashes, and historical failed attempts. A 256² compute-written HDR storage buffer
feeds a single-sample RGBA16Float pass, followed by tone mapping into the main
four-sample sRGB target. Quantitative partial-alpha probes at exposures 1 and 2
match `(213,156,0,128)` and `(231,188,0,128)` on the left; the right probe detects
premature HDR clamping. Transparent pixels must equal zero exactly. Each changed
frame uploads only a 32-byte uniform; chrome changes and native resize retain
the program resources. Both windows have independent storage and float textures.

The native checks include admission refusal without allocation, explicit sync and
async coexistence, recursive busy refusal, capture metadata across resize, quiet
cancel/rebuild, completion while actually minimized without another frame, native
closure delivery suppression and surviving-window capture. Old-frame delivery
after target replacement was observed in both runs; scheduling is not guaranteed.
The initial WGSL reserved-field rejection and restored-size fixture failure are
preserved and explained alongside their corrections.

Internal review corrected completion admission under a full mailbox, native
teardown ownership, cancellation reservation, exact transparent/HDR probe rejection
and resize fixture precision before Prism. PNG replacement tests preserve existing
files on failure/cancellation and replace complete bytes with the existing mode.
The chooser was simulated, and native buttons used accessibility actions. No OS
preferences or input sources changed. Private staging destruction after forced
close is supported by ownership/control-flow review rather than a native private
counter measurement. This checkpoint does not certify presentation latency,
sustained performance, physical device loss, cross-display geometry, external
accessibility clients or public distribution. The remaining framework work is
listed in the [Mac completion audit](../MACOS_COMPLETION_AUDIT.md).

The stronger closure fixture exposed a real upstream registry race between focus
and another window closing, plus a fixture assumption that raw native Close also
removed the framework registration. The failing run remains visible. Checked
patch25 uses atomic existing-member promotion, detached registry snapshots, no
native callouts under the registry mutex and post-initialization publication.
Five collection/lifecycle fixtures pass twenty race repetitions; the final native
runs and existing 3D regression pass. The fixture performs explicit domain cleanup
after the raw-native-close observation and checks normal last-window cancellation,
registration removal, program release and the completed native teardown lock barrier.
It does not infer private decoder/reservation completion. Stock upstream core vet
still reports preexisting composite-literal/unreachable diagnostics; remaining
analyzers pass with those two classes excluded, as recorded in backend-results.json.
