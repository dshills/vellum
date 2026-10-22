# Private completed staging prototype: first refusal

The first native off-mode process exits successfully, but its case is **rejected**
because the diagnostic uses an empty derived device display name. The controller
stops immediately; enabled-pool mode never runs. No native pooling correctness
or latency acceptance is claimed.

The frozen Core helper derives GPU.DeviceName from AdapterInfo.Description or
Vendor and ignores Device. The pinned native API puts the actual adapter name in
AdapterInfo.Device; the optional description/vendor fields can both be empty.
The fixture now records that primary Device field and complete AdapterInfo,
without changing the framework helper or loosening any requirement.

The rejected raw document retains 24 CPU writes, backend Metal and no native
error. A **posthoc-only** synthetic-name replay passes the remaining 22-readback/
source-bound lifecycle checks; it is not accepted as a valid native case or
hardware identity. The original raw journal and REJECT outcome remain untouched.
All nine owned native/probe groups retire, and the second case never starts.
Sources, binary/library, logs, configuration and geometry are retained/pinned.

A fresh v2 scope is prepared only for this concrete metadata fix, using the same
native library, 24-write/22-readback matrix and bounds. It needs its own reviewed
source checkpoint before execution. See the [protocol](../../../tools/macos-calibration/stagingreuse/PLAN.md),
[build identity](build-identity.json), [source audit](source-audit.json) and indexed
evidence. The original preparation has complete 15-file plus focused four-file
Prism Anthropic Claude Sonnet 5.5 reviews, policy5/protocol9/scalar-concurrency1
checks and source-backed dispositions. No shipping library, SDK, app, renderer,
pacing, feature or 33.3 ms gate changes; M3–M7 remain open.

Backlog: the Core adapterName display helper should eventually fall back to Device
when Description/Vendor are blank. It does not block application/GPU operation;
the frozen framework is not changed for this diagnostic display-label issue.
