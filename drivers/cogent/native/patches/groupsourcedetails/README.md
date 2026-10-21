# Grouped source-detail initialization

Stage0106 gives an opt-in source-detail FrameTrace.BeginGroup the same empty
bounded ticket storage already allocated by Begin. The zero-value ordinary trace
still allocates no source-detail storage. No rendering, event routing, readiness,
owner/reconcile, armed generation, default renderer/pacing or public API changes.

The omission caused existing grouped detailed diagnostics to fail their first
composition with "source diagnostic composition details unavailable". Two real
Core regression methods exercise prefix/final accounting, exact32-record shared
budget,33-record refusal, detached prefix scalar data, cancellation with in-flight
composition and ordinary trace isolation. The final tests are also run against
originalstage105 as an assertion-failure control, never native presentation proof.
This corrects diagnostic initialization; it does not repair display latency or
qualify the Mac framework. The immutable retained SDK remains unchanged.
