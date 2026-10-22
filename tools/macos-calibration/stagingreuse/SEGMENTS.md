# Segmented request schedule diagnostic

The first private framework timing attempt is retained as rejected: fifth request
42.563959 ms late after four warmup receipts, zero measured samples; enabled mode
never started. The existing lateness metric runs after a synchronous main-thread
foreground/configuration observation. It cannot distinguish a late timer wake from
a slow observation; neither is established as the cause. No pool timing conclusion.

Add only two time.Now observations around that same foreground check, storing one
preallocated scalar record per request. Keep the original post-check time.Since,
20 ms schedule refusal, receipt deadline, foreground/config checks, 120 ms period,
128 warmup/120 measured, actual presentation and all original accounting. Signed
wake offsets and residuals remain raw. This adds observer cost; it is diagnostic,
not an observer-neutral calibration or a timer/presentation repair.

One newly instrumented off/on pair, original WC and identical byte-verified native
library/binary both modes, 40 seconds/case and 90 seconds overall. Reuse all source,
SDK, geometry, lease, owned execution and stop-first-invalid rules from TIMING.md.
A valid pair need not reproduce the old delay and does not explain it. If rejected,
retain partial segments as diagnostics only and stop; no unchanged retry/waiver.

Three focused methods cover full parent validation in both modes, signed wake,
six malformed segment refusals and partial failed schedules without acceptance.
Go build and Prism Anthropic Claude Sonnet 5.5 precede source checkpoint and native
execution. No shipping SDK/app/policy or gate change: features frozen, experimental
rendering/pacing disabled, 33.3 ms presentation gate unchanged, M3–M7 open.
