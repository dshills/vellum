# Native render/present submission control

The completed five-pair calibration remains sealed. The locked wgpu-hal 29.0.3
Metal `Queue::present` creates and commits a separate present command buffer;
the calibration reference used combined render/present submission. Registry
archive and member hashes bind that source audit. This difference is a hypothesis,
not a proven extra-frame cause or production repair.

Run three native-only pairs, orders combined/split, split/combined,
combined/split. Reuse the valid locked match from the completed calibration.
Both modes keep its pixels, cached textures, FIFO/vsync, two drawables, timeout
and transaction settings, foreground checks, 128 warmups, 120 measured requests,
120 ms schedule and strict per-state presentation accounting. Only command-buffer
submission topology changes; both modes add identical bounded callback timing
observation. Split mirrors the audited Metal queue's separate present command
buffer. No framework/API/SDK/default or release-gate change is made.

Record render commit start/end, present request/commit start/end, native render
GPUStartTime/GPUEndTime, command-buffer status and scheduled/completed callback
observation times. Callback clocks are CPU observation times, not GPU execution
or presentation timestamps. Empty present buffers may have no GPU timestamps;
require native render GPU times for every cohort frame and retain setup values.
Both callback roles are entered into the owned group before commit and fully
drained before output; storage is fixed/atomic with no eviction.

Use one exclusive lease, a 240-second campaign bound and 40-second process
watchdogs. One attempt per case; stop first invalid/focus/config/source/deadline/
process/timestamp failure, keep all raw data, no unmodified retry. Compare per-run
p50/p95/p99/max and three differences of p95 values without pooling or pure-cost/
OS-floor claims. No release acceptance follows. If split does not repeat the
extra refresh-period band, reject it as a sufficient cause and inspect another
boundary before any native library rewrite. If it does, verify a safe architecture
change independently; do not adopt from this diagnostic alone.

Focused classifier controls, warning-free native build, internal review and Prism
with Anthropic Claude Sonnet 5.5 precede execution; result validation/review and a
separate commit follow. No broad framework test campaign is needed.
