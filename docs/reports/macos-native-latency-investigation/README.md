# Native presentation latency investigation — 21 October 2026

**The latency blocker remains unresolved. No candidate is adopted.** This
checkpoint establishes the actual native configuration and rejects four private
presentation candidates under the unchanged capture and latency rules. Framework
features, shipping scheduling, the retained SDK and the internal application
remain unchanged. There is no Mac release-completion claim.

## What is established

The observed baseline CAMetalLayer really uses **two drawables, vsync enabled,
transaction presentation disabled, and acquisition timeout disabled**. Thus the
existing Metal maximum-frame-latency hint takes effect. Adding that hint again,
reducing drawables below the supported minimum, or enabling the rejected display
credit experiment is not a repair.

The [preceding assessment](../macos-default-responsiveness/README.md) retains
serial grid/tree p95 48.130/48.229 ms and paced-grid p95 62.224 ms. Its paced-grid
cadence passes one run. Existing timestamps have two acquired-but-not-yet-shown
frames for every serial measured acquisition; paced grid has two for 506 states
and three for 94. This is a timestamp-derived count, not native object allocation
or GPU queue-depth telemetry. CPU work and receipt/acquisition/presentation
boundaries remain separate; their percentiles must not be added.

The configuration probe is an observational Go/cgo overlay on immutable SDK
`bd3cbc9ce2566955dae5dbb61385b0e1a9307eaa`. The source-built native library and
optimized Go 1.25.6 Darwin arm64 consumers are retained. The standalone Objective-C
file overlay was not included by this toolchain, so that earlier run is refused
as configuration evidence; only the subsequently confirmed Go/cgo overlay proves
the settings. A still-earlier preflight inherited a sealed stop root and launched
no process. These preparation failures remain recorded, without restarting old
campaigns. A mistakenly created diagnostic directory was moved out of the SDK
before building/running; full SDK membership/hashes were restored and verified.

## Private candidate decisions

Each grid trial retains the existing 10,000 stored records, viewport, 128 warmups,
120 measured requests, actual native timestamps and strict prefix classifier.
No losing state, startup rule or threshold was removed to make a candidate pass.
The paced 600-state follow-up was conditional on passing serial validity/latency;
none qualified for that follow-up. These are separately bound source variants,
not additional baseline repetitions or a pooled performance population.

| Candidate | Result | Decision |
| --- | --- | --- |
| Existing transaction-presentation path | All 251 prefix records have skipped feedback | Reject; no latency number accepted |
| Same path with explicit render-thread transaction flush after Present | All 251 prefix records present; p95 47.856 ms, maximum 48.441 ms | Reject; valid capture still exceeds 33.3 ms |
| Temporary display link per acquisition | 83/251 prefix records skipped | Reject; invalid capture |
| Persistent paused display-link owner with one drawable slot | 1/252 prefix records skipped, including measured tag 62/sequence 194 | Reject; invalid capture |

The persistent prototype fixes the one-shot link lifetime boundary, uses one
private native thread/run loop and a coalescing wake source, and integrates
retirement with the existing native-layer teardown seam. Its focused stand-in
fixture covers persistent reuse, delivery, a 250 ms absent-drawable timeout,
retry, joined/repeated retirement and a late scalar presentation callback.
It does not establish actual GPU resource deallocation. A native-owner join
timeout is a two-second explicit refusal/quarantine, not successful cleanup or
a production lifecycle policy. This prototype remains private and unqualified.

Prism with Anthropic Claude Sonnet 5.5 reviewed both display-link implementations.
Internal and Prism findings corrected default-mode reentrancy, broker pointer
locking, and publication of shutdown completion before native release/pool drain.
Existing intentional first-acquisition admission and missing-feedback credit
refusal rules remain unchanged. Dispositions and focused reruns are retained.
The native prefix failure after those repairs prevents adoption regardless of
passing software fixtures. No broad framework campaign was rerun.

## Minimal native reference and limits

A separate 1519×1266 changing-color Metal clear reference bypasses Vellum, Go,
WebGPU, layout and text. It requests the same two-drawable synchronized windowed
presentation policy, 128 warmups and 120 measured states. It returns normally,
receives every callback, closes its window and retires its owned process.
All 120 measured presentation times are positive; their diagnostic p95 is
48.832 ms, maximum 48.890 ms.

**The reference is not qualification evidence:** 14 startup records have zero
presentation time, all retained. It does not have the framework fixture's
functional workload or strict capture association. It neither proves a macOS
latency floor nor exonerates the framework/backend. It does show why another
CPU-only optimization or a wholesale renderer replacement is not justified by
these observations alone.

Apple documents that disabling
[display synchronization](https://developer.apple.com/documentation/quartzcore/cametallayer/displaysyncenabled)
can cause tearing, and that
[display-link latency requests](https://developer.apple.com/documentation/quartzcore/cametaldisplaylink/preferredframelatency)
may be exceeded in windowed macOS. Neither is permission to lower the release
standard. Apple's
[frame-pacing reference](https://github.com/apple/game-porting-toolkit/blob/main/game-porting-skills/skills/presenting-metal-drawables/references/frame-pacing.md)
delivers the drawable to the render loop; wrapping `nextDrawable` is not the
same integration boundary.

## Next implementation boundary

Stop these rejected acquisition/presentation-policy approaches. Before another
performance trial, define the native frame request at **frame production**, ahead
of scene reconciliation and GPU acquisition, and explicitly hand the provided
drawable through the existing surface owner. Avoid nesting acquisition under an
already-running display link. Preserve the input receipt before queueing, all
required states, FIFO/vsync, cancellation and native resource ownership.

The smallest next diagnostic must bind a matched native reference to real
window readiness, preserve startup losses separately, and record display-link
callback/target-presentation timestamps. The current missing target timestamps
and failed prefix leave the reason for the extra presentation periods unresolved.
That diagnostic and the explicit provided-drawable seam are prerequisites for
selecting a safe backend repair; this checkpoint does not claim they exist.
Do not repeat these rejected candidates, alter the 33.3 ms/60 Hz gates, enable
tearing or ship an experimental path as a fix. M3 and M4–M7 remain open.

## Integrity, cleanup and record

Seven native commands and 16 configuration probes were reaped and observed
absent. This establishes process retirement, not complete native deallocation.
The two large private source copies differ from the immutable prepared Core in
one recorded Objective-C member. No complete SDK or historical archive is copied
into this report. The seven binaries and helper/source hashes remain bound in
[summary.json](summary.json) and the archived receipts/plans.

The second quiet window began at 17:50:07 UTC; the last reference started at
18:00:07 UTC and ended about five seconds beyond its ten-minute limit. No further
screen work followed. This small overrun is retained rather than claiming the
window was strictly met. Configuration/geometry limits in each actual record
remain authoritative; the reference does not use the framework geometry store.

[evidence-index.json](evidence-index.json) binds the
[evidence archive](evidence.tar.xz): probe/candidate sources, failed captures,
logs, configurations, exact commands, source/binary bindings, review dispositions
and focused stand-in checks. Executables, full prepared Core/SDK payloads and
historical evidence archives are excluded. The staged report review with Anthropic Claude Sonnet 5.5 completed with full
coverage and zero findings. No production behavior changed in this commit.
