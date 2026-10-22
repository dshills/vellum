# Mac release responsiveness policy

**Developer-preview policy, approved 22 October 2026:** the
[first-release plan](macos-first-release-plan.md) governs the macOS preview.
Numeric latency/cadence qualification and full endurance/platform certification
below remain future full-qualification work, with historical failures preserved.
Correctness, bounded resources, required workflows, independent SDK consumption
and applicable distribution rights remain preview blockers. This does not mark
M3–M7 complete or authorize public publication/signing.


The owner selected this policy on 19 October 2026: **“Use measured 60 Hz
responsiveness as the release gate; keep 4 ms as an optimization target.”**
This decision supersedes treating the prototype's provisional CPU budgets as
mandatory Mac release thresholds. It does not establish a passing measurement.

## Release scope

The initial Mac release targets the tested local Mac configuration at 60 Hz.
120/144 Hz remain additional performance tiers to measure and document; they
are not prerequisites for this 60 Hz release. Compatibility and high-DPI support
must still be stated accurately for the configurations actually tested.

The release decision uses sustained, measured frame pacing and input response
for representative professional workflows: the composed workspace, fixed and
variable 10k/1M stored tables with matched visible work, representative trees,
text/property editing, canvas interaction and resize. A button-only animation or
empty frame loop does not qualify the framework. Preserve functional assertions,
data sizes, sample counts, raw samples and ordinary/diagnostic separation.

## Measured acceptance

Use the existing 60 Hz pacing/input criteria from the measurement contract:

- During continuously changing visible work, at least 99% of presentation
  intervals meet the 1.1-times-period criterion (18.333 ms at 60 Hz). Report all
  intervals longer than two periods, including outliers.
- Routed-input receipt to the corresponding actual presentation has p95 no
  greater than 33.3 ms. Record the input method; automated routing does not
  establish physical input-to-photon latency.
- Resize remains responsive, with no application-caused stall over 50 ms.
  Report observed resize timing and separate any known OS modal behavior.

These retain the existing end-to-end criteria; the owner changed which budgets
govern release, not the recorded results. Record the actual refresh, scale,
power/configuration, optimized build, source identity and measurement boundary.
Use warmup and at least five measured runs, retaining p50/p95/p99/max and failures.
Refresh/compositor quantization must be identified and any alternative cadence
criterion declared before its campaign; it cannot be an after-the-fact excuse.

Presentation means native display feedback. A successful Surface.Present call,
GPU submission, CPU-work duration, timer tick or software frame completion does
not prove presentation. Unavailable feedback leaves the measured gate open.
Dropped, missing or nonpositive presentation timestamps must not disappear from
the accounting. Occluded/minimized/idle phases retain their own lifecycle and
idle checks; an idle application is not required to repaint at 60 Hz.

## Optimization targets and prior evidence

Keep 4 ms p95 / 6 ms p99 UI work, the 60/120/144 Hz CPU sub-budgets, startup,
layout, shaping, allocation and GPU-time budgets as reported optimization
targets. An isolated miss does not block this Mac release when the required
60 Hz responsiveness and the remaining functional/safety gates pass. The 0.4 ms
gain/four-of-five-pairs rule was an experiment-adoption policy, not a user release
requirement; it no longer obliges another renderer redesign before measuring
the release gate. Preserve existing rejected experiments and their sealed scopes.
New performance changes still need demonstrated benefit, correctness and no
material regression on the supported release path.

The historical table matrix still misses its original 4 ms target in 33/40 runs.
It did not measure actual presentation, so it neither passes nor fails the new
release gate. No result, golden, timing sample or prior failure is relabeled.

Correctness, exact commands/history and input ordering, accessibility/platform
workflows, lifecycle cleanup, bounded resources, explicit allocation/admission
limits, idle/endurance checks, final SDK/notice checks and the final audit remain
required. The provisional footprint numbers remain reported targets; ownership,
bounded retention, declared cache limits and leak/regression checks remain hard
safety requirements. This is not approval to ship unbounded growth or stale work.

## Immediate completion work

1. Finish the three dedicated Cocoa methods in M2 when the desktop is available.
2. Qualify the implemented [bounded native presentation recorder](native-presentation.md)
   with real Metal feedback, timestamp/loss identity and native lifecycle checks
   before the 60 Hz campaign. Headless checks do not establish display feedback.
3. Measure the existing candidate against this policy before selecting another
   rendering optimization. A failing 60 Hz workload gets a focused repair.
4. Complete resources/endurance, real Mac workflows, SDK/rights and final audit
   under M4–M7. Full Mac completion remains open.

See [delivery plan](macos-delivery.md), [measurement contract](prototype-plan.md#measurement-contract)
and [Mac completion scope](macos-completion.md).


## Coalescing busy paint requests

A paint request received while composition owns the render context must schedule a later dirty-state rescan. Busy work alone must not force another unchanged composition. Queued input/scene/sprite/rebuild invalidations must remain visible to that scan, and resize/cache fallback must retain its separate required composition. Test unchanged retry, actual queued dirty pixels and resize fallback independently. Stage0100 implements this scheduler correction; native timing comparison is required before attributing a performance improvement. No latency/cadence threshold changes.


## Reusing a successfully applied cursor image

A repeated cursor request may omit the owned pixel copy and synchronous native-main handoff only after resolving and validating the requested enum, units, raster and current generation. Require the same live window, unchanged packed display scale and pixel size, exact preceding successful cursor-request sequence, a valid bounded standard RGBA raster, normal known enabled/nonraw mode, no explicit mode request, and an open main queue. Publish only a genuinely successful native assignment; retire its mirror with the window. Failed, superseded, size/theme/scale-changing, stopped or malformed requests must retain normal validation/ordering. Any explicit visibility/capture/raw-mode request permanently retains the existing path for that window. Native cursor-owner state remains confined to native main.

Stage0101 implements this narrow optimization. Software tests verify duplicate queue/copy avoidance and invalidation, with an original-source negative control. A fresh source-bound native comparison is required before attributing a cadence or latency improvement. The source100 native record in [the checkpoint report](../docs/reports/macos-presentation-post100/README.md) remains a measured 60 Hz failure; its thresholds and populations are unchanged.


## Mac Metal FIFO latency request

Stage0102 requests DesiredMaximumFrameLatency=1 only on Darwin with an actual Metal adapter. Preserve FIFO and all configuration/ownership/retirement paths. The native capability clamp remains authoritative; the request does not guarantee a drawable count on every OS/device. Fewer available drawables may block acquisition, so compare unchanged paced cohorts and verify closure/reconfiguration before retaining this policy. Remove the request if it regresses accepted cadence, cleanup or responsiveness. Existing GPU acquisition/lifetime race tests validate unchanged ownership support; they do not execute native InitConfig or establish timing. No denominator, latency threshold, callback rule or release gate changes.


## Busy render completion rescan

A paint admitted during an active composition must coalesce one payload-free pending dirty-state rescan. The completing ordinary or measured batch first decrements active work, then requests an immediate rescan only when it is the last active batch and pending work exists on a live window. Keep the existing one-refresh timer for late-publication/lost-wake races; normal idle admission and closure clear pending state. Preserve all composition/native ownership, resize forcing and frame-trace completion. A rescan must not force an unchanged frame, add an unbounded queue, spin, or schedule work after every completion. Stage0103 verifies this behavior with a controlled software compositor and original-source behavioral negative; native latency causality/improvement requires a new source-bound comparison. Stage0102 remains experimental and unqualified; source102's failed capture stays refused, with no changed threshold or startup predicate.


## Empty inactive native-text publication reuse

Only the no-focused/no-current-client branch may reuse a previously successful empty, inactive client-zero native publication. Keep one owner-confined backend/lifetime/exact-ack key; retain no document, geometry, cells or callback copies. Reuse requires live owner/window lifetime, native readiness, a nonterminal native main, the same safely comparable backend identity and zero reserved, committed or in-flight ordered input credit. Unsupported native publishers remain uncached. Revalidate the key, liveness and readiness after the real synchronous native call before retaining acceptance; refusal/error/not-ready attempts must never establish success.

A cache hit must still update Go activation snapshots and grid admission rules. Focus-loss finalization precedes branch selection. Every active/client publication attempt invalidates before validation, and explicit deactivation must continue to perform the real native handoff/unmark. Preserve FIFO, preflight/commit, acknowledgments, settlement and shadows. Native attachment retirement atomically revokes readiness before closing the main-owned backend, including the interval before owner WinClose cleanup.

Regress duplicate dispatch avoidance, Go-rule refresh, ack/backend/lifetime/readiness changes, refusal/error retry, pending ordered input, activation/deactivation and terminal native refusal. Original-source behavioral counterfactuals must distinguish redundant work from required terminal refusal. Headless fake backends and aggregate execution-trace profiles do not establish native performance or physical input. A fresh unprofiled source-bound native comparison is required; no cadence/latency gate, cohort, capture classifier or prior failure is changed.

The [checkpoint report](../docs/reports/macos-empty-inactive-publication/README.md) records eleven focused race regressions, meaningful original-source/V1 counterfactuals and the independently audited complete632-method offscreen race pass, preserving the first short timeout. Fresh native comparison remains separate.
