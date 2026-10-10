# Windows/Linux adapter implementation evidence

The recorded implementation adds native-main-queue text clipboard services,
Windows STA/WinForms and Linux GTK file chooser transports, and command menu bars
inside each native window. The [specification](../../../specs/platform-support.md)
describes the contracts and remaining qualification gates.

Validation performed on macOS:

- Portable transport/dispatch tests pass with the Go race detector. They exercise
  Unicode encodings, direct argv/stdin data, native exit/startup diagnostics,
  cancellation, queued/shutdown callbacks and private result ownership.
- Both Windows and Linux service implementations compile against the pinned
  current-host Cogent APIs in explicit temporary fixtures. Actual service menu
  snapshots/independent native widgets and disabled-command availability rechecks
  pass with the race detector. GOOS filename/build filtering is deliberately
  removed only inside those fixtures, so this is not target compilation.
- The pure Go `internal/platformio` test package crosscompiles to Windows amd64
  PE and Linux amd64 ELF. Those binaries have not run on their targets, and do not
  compile or link the complete native driver.
- The new platform workflow passes local `actionlint`. No workflow has been
  dispatched and no target CI result is claimed.

The target runtime executable covers two-window GPU pixel readback, lifetime
isolation/shutdown, text clipboard read, menu snapshot installation and chooser
request cancellation. Optional disposable-desktop clipboard writing restores the
plain-text flavor only. Chooser display readiness/selected paths and menu
activation still need interactive acceptance. No Windows/Linux native runtime
smoke was performed on this macOS host. Shared macOS regression smokes may be
recorded separately after integration; they do not qualify either target.

Full platform qualification remains blocked by Windows native IME/TSF and UI
Automation, Linux IBus/Fcitx and AT-SPI, native Wayland, and actual hardware/DPI/
interactive desktop acceptance. `--require-complete` reports these gaps and fails;
a successful build or basic runtime smoke cannot claim release readiness.

Raw records are the small text files in this directory. `source-hashes.json`
identifies the adapter/check/spec source recorded here. The parent phase review
and commit records cover subsequent integration and any follow-up fixes.

The final current-host fixture checks include native clipboard error preservation. Follow-up race tests cover detached failure-report snapshots during concurrent worker progress and shared chooser filter validation. Fixture replacement directives are generated with `go mod edit`, which supports grouped directives and quotes native absolute paths.

Two final macOS race-instrumented platform runs pass with six GPU Snapshot/resize/readback cycles per run, independent window teardown and final shutdown. The first attempts exposed unsafe borrowed-view cache replay and a cross-window Snapshot layout read. Owned patch19 now forces fresh composition for borrowed views; Snapshot uses the viewport captured by Render after layout. Raw failures and final checks are retained in the rendering report. These macOS runs qualify shared integration only; Windows and Linux target execution is still required.

The final native records also verify image dimensions against the completed owner
viewport and apply a center/zoom edit immediately before Snapshot in the same
owner callback. The moved rectangle probe changes from background to colored
pixels without an intervening native Render. The fixture waits for actual canvas
Show publication before readback and waits on the captured window's atomic closed
state after teardown; it never polls names during native creation. Each record
includes race-build/source metadata and exit status 0.

The [diagnostic result follow-up](diagnostics-followup.txt) preserves an already
completed worker failure together with a native run failure, without waiting on
unfinished work after the native loop fails. Focused race tests, standard vet and
a race build pass. The final native GPU logs above now include this diagnostic
follow-up and the bounded two-axis resize acceptance fix. Their source SHA256
records agree with the current source and source-hashes.json.

The [bounded resize follow-up](resize-followup.txt) requires each requested native
pixel-size transition and completed canvas changes on every resized axis within
five seconds. Clamped/ignored requests and already-matching native sizes fail
qualification explicitly. Focused tests include height-only acceptance, ignored
and clamped requests, stale width/height and missing allocations. Both final native
runs use the current CLI source and pass six actual requested transitions; emitted
runtime hashes are checked against source-hashes.json and the current files.

The final native records also wait for actual canvas Show publication in the first
window before its initial readback, matching the second window's barrier. Both
current-source race runs pass with this startup topology ordering in place.

The [missing helper follow-up](missing-helper-followup.txt) verifies that missing
bare, absolute and relative executable paths all return `app.ErrUnavailable`.
An existing path that cannot execute retains a native startup error. Portable
transport race tests and vet pass after this final error classification fix.
