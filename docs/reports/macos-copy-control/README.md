# Offscreen Metal copy control result

All three predeclared pairs are valid. **Copying into the same source buffer after
its GPU work has completed is much faster than the first copy.** Default-cache
fresh copies are somewhat faster than write-combined ones in this control. Neither
result alone explains the larger CPU copy interval measured inside the framework.
No cache policy or memory pool has been adopted.

| Pair | WC fresh p95 | Default fresh p95 | WC subsequent p95 | Default subsequent p95 |
|---|---:|---:|---:|---:|
| 1 | 0.684 ms | 0.507 ms | 0.103 ms | 0.087 ms |
| 2 (reverse mode order) | 0.670 ms | 0.471 ms | 0.091 ms | 0.091 ms |
| 3 | 0.642 ms | 0.498 ms | 0.097 ms | 0.102 ms |

The direction of subsequent-copy differences between cache modes is inconsistent
across pairs; no significance or noise bound is established.
Each value describes its own 120 measured iterations; raw paired
subsequent-minus-first distributions are retained separately. Quantiles are not
additive. The first copy can include physical page backing; the second follows
GPU use and readback, which changes cache/memory state. These serial waits and
persistent source patterns differ from the framework's asynchronous upload and
source-image workload. The preceding framework CPU-copy p95 was 2.612 ms: do not
attribute that gap to cache mode, infer native binary equivalence or subtract
these independent distributions to claim a measured UI improvement.

Every iteration makes a fresh shared source buffer, copies pattern A, blits it
through a persistent private RGBA8 sRGB texture into readback, waits for successful
GPU completion and compares all 7,692,216 bytes. Only then does it copy different
pattern B into that source and repeat full verification. All 2,976 transfers pass
(22,892,034,816 bytes in total). A corrupted-last-byte
comparator control detects mismatch and restoration. There is no in-flight CPU
overwrite or asynchronous buffer pool. Default tracked hazards remain in both
modes; only CPU cache mode differs. Source data/destination getters and descriptor
validation have the scopes stated in the protocol.

The six cases run 21 October 2026 at 22:28:44.983–22:28:55.783 UTC
(17:28:44–17:28:55 Chicago), 10.800355 seconds within 120. All 54 owned native/probe
process groups retire. Source/compiler/binary/log pins, before/after configuration,
persistent geometry and retained SDK inventory (8,321 files) are unchanged. No
window opens; this supplies no foreground UI or physical presentation evidence.
Process retirement and ARC scope do not prove native allocation balance.

The [protocol](../../../tools/macos-calibration/copycontrol/PLAN.md),
[build identity](build-identity.json), [exact summary](summary.json) and indexed
archives retain five focused test methods, actual compiler/build receipts, all
warmups and measured rows, configuration probes and both complete Prism Anthropic
Claude Sonnet 5.5 preparation reviews. The first review's orchestration robustness
finding was fixed; follow-up directory sealing and stop-precedence findings were
disposed with source-backed explanations. One initial link failed because the
output directory did not yet exist; no native test ran, and that failure is
retained. The reviewed source checkpoint is 0f57532.

Next assess an actual private-framework cache-mode control before selecting any
policy, while keeping all ownership, resource completion and validation paths.
The normal framework, frozen SDK and Placement app remain untouched. Features
stay frozen, experimental rendering/pacing stays off, the 33.3 ms
receipt-to-actual-presentation gate is unchanged and M3–M7 remain open. No ordinary
UI latency, endurance, allocation-balance or distribution gate closes.
