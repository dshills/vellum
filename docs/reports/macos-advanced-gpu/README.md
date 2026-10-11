# Native advanced GPU canvas acceptance

The public `cogent.GPUView` uses the existing window device, an ordered preparation
encoder and framework draw pass, retained MSAA/sRGB targets and optional checked
depth. The independent viewport consumes the SDK through its own Go module.

[Recorded checks](acceptance.json) identify source inputs, commands and successful
exit status. [Native viewport counters](viewport-native.json) prove near-first
red/far-second blue depth pixels, changed orbit pixels, camera-only 64-byte
uploads, independent helper buffers, acknowledged chrome retention, controlled
Prepare/Draw failure cleanup, budget refusal, rebuild, three settled target
resizes, empty clearing and closed resource queries. Four-sample depth adds16
bytes/pixel, giving36 target bytes/pixel. No input/VoiceOver settings changed.

The first native attempt exposed cached window-list indexing across unlocked
CloseReq callbacks and exited with an index panic. Reproducible backend patch23
fixes this through mutex-protected snapshots and removed-window guards. Twenty
repeated race tests exercise concurrent removal, veto/new windows and reentrant
quit callbacks; the final native process exited0.

The strengthened resize check initially waited for a target change after sizes
that converged to the same actual allocation. It now uses distinct larger native
sizes and awaits settled layouts while retaining the changed-size/generation,
36-byte target and unchanged geometry-allocation assertions. This is fixture
allocation acceptance, not cross-display/DPI qualification.

The full suite passes89 Python tooling checks and root/native/backend/example/SDK
race tests, vet and builds. The final consumer-only smoke correction has affected
race/vet/build/native checks. [Existing drawing regression](drawing-native.json)
passes the million-sample,100k-object and real4096² retained-image workloads on
the changed backend, including one changed64KiB tile.

Internal review corrected distinct callback helper leases, guarded stale input
geometry and one-shot destruction after remount. Prism identified swallowed key
and scroll events; the explicit consumed-result callback now preserves
unhandled navigation/shortcuts/parent scrolling, with real-event race tests and
affected native-driver/consumer checks. The three other Prism findings were
independently adjudicated against ownership, cleanup and logical accounting. Compute, floating-point
presentation, asynchronous readback, presented clipping/MSAA edges, sustained
performance, physical device loss and broader Mac qualification remain open.
See [contract](../../../specs/advanced-gpu.md).
