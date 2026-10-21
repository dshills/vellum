# Actual backend timing diagnostic preparation

A private source-built observer records actual Metal render submit, GPU execution,
present request and drawable feedback. It is never installed in the retained SDK
or packaged app. The [protocol](../../../tools/macos-calibration/backend/PLAN.md)
limits execution to three on/off observer pairs under four minutes; all earlier
failed campaigns stay sealed and the release gate remains unchanged.

The original locked backend source and package versions are preserved, except an
explicit local wgpu-hal path patch adding fixed atomic trace storage and callbacks.
The same diagnostic binary/archive runs both observer modes. This controls observer
impact but does not prove binary equivalence or zero overhead. Callback observation
clocks are labeled separately from GPU execution and actual presentation clocks.

Native and Go diagnostic builds pass. Nine backend classifier methods, including
eight invalid-record and four drain/overflow subcases, pass alongside the existing
18 parent controls. The original view/routing/pixel tests are reused by exact source
identity and unchanged SDK, rather than rerun. Inherited compiler warnings and
preparation failures are disclosed in [build identity](build-identity.json).
No live or release acceptance follows from preparation. Reviewed results will be
added after source checkpoint commit and bounded native execution.
