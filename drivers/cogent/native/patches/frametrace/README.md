# Opt-in native work tracing

Stage 34 follows the scene-owned cursor fix. Preparation verifies exact input
and output bytes; the default native event/render path retains its draw order.

A tagged initialized grid scroll occupies a protected deferred FIFO node and
unwraps before real native dispatch. The recorder tracks raw input, posted owner
work, native reconciliation, all post-input intermediate native update/compose
work, and the first frame armed after the accepted owner generation. A prefix
composition cannot complete a later generation that arms while it is running.

Elapsed work intervals include GC and driver work. Surface lock/acquisition
calls and the present call are separately reported; configuration and view
creation remain measured work. These are not kernel CPU counters, GPU execution
time or compositor presentation feedback. One active native window is required.

Cancellation keeps diagnostic ownership through queued input/owner/frame
terminals. Native window retirement releases pending observation. Invalid,
refused, unchanged, unsupported or abnormally terminated actions cannot qualify
a completed sample. Synthetic race tests check ticket/phase accounting, prefix
late-arm ordering, cancellation terminals, dispatch identity and draw order.
Physical GPU fault recovery and broader performance/endurance remain separately
qualified; this patch does not certify them.
