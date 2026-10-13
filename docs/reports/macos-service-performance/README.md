# Native services and accessibility allocation

The Mac menu/tray lifecycle fixture passes 100 extra create/clear cycles, and the
candidate passes the stored 400-cell Cocoa accessibility fixture. Accessibility
publication now avoids one semantic byte-to-string copy and preallocates completed
visible grid records. Process-wide allocations fall in the matched table campaign;
frame timing remains mixed and every case still fails the unchanged 4/6 ms targets.

## Implementation and ownership

The baseline is commit `2338d7f56d6acf7c5213b21f1a78cfc8f471909c`. The candidate changes
two adapter implementation leaves and one independent-oracle test leaf. A local
JSON encoder writes compact semantics into an unreused string builder; only its
record-terminating newline is removed. Its output must equal stdlib `json.Marshal`,
including escaping, unusual values and encoding refusal. Later calls cannot mutate
the returned string. Publication rejection ordering, pending/completed transactions,
detached grid caches, action guards and native envelope ownership remain unchanged.

The grid prepass counts nonempty completed cells/headers and unique visible row
keys, then reserves node/action capacity. It preserves first-visible row order and
excludes empty overscan. It introduces no domain callback or native geometry read.

## Native services and correctness

The existing real Cocoa menu/tray fixture passes native selector dispatch, disabled
and stale action refusal, menu replacement/reordering, named/template bitmap
metadata, one-window closure with a surviving service and terminal cleanup. Its
final report records 100 extra create/clear cycles, the original application menu
restored, tray/targets retired, callbacks stopped and no tracked window, bridge,
pending callback or native operation. It takes about 1.66 seconds. Native selector
dispatch is not a physical click, VoiceOver or third-party service workflow.

The candidate separately passes the stored 400-cell Cocoa publication/identity/
action/geometry fixture. Focused accessibility race tests run three times; the full
adapter offscreen race suite and vet pass. These checks do not change input sources,
VoiceOver, clipboard contents or OS preferences. Live execution uses the human's
fresh machine-availability authorization; continuous host isolation is not proven.

## Matched ordinary performance

Two optimized executables use the default scene renderer, the same 6,198 prepared
dependency identities and 710 selected source identities apart from the exact
three candidate leaves. Four fixed/variable 10k/million-row cases each retain five
fresh interleaved baseline/candidate process pairs, alternating launch order,
128 guarded warmups and 120 measured events. All 4,800 ordinary samples are retained.
Every paired geometry record matches exactly and each frame has at least 400 clipped
cells. No profiling or source-detail instrumentation is enabled in this campaign.

Pooled total recorded UI work p95/p99 in milliseconds and median process-wide
allocation bytes per event:

| Case | Baseline p95 / p99 | Candidate p95 / p99 | Allocation p50 bytes |
| --- | ---: | ---: | ---: |
| fixed-10000 | 9.125 / 9.590 | 8.952 / 9.752 | 8,200,640 → 7,475,848 (-8.8%) |
| fixed-1000000 | 8.492 / 9.717 | 8.584 / 10.590 | 7,609,184 → 6,871,280 (-9.7%) |
| variable-10000 | 8.586 / 8.868 | 8.580 / 9.048 | 7,767,016 → 7,125,848 (-8.3%) |
| variable-1000000 | 8.531 / 8.853 | 8.566 / 8.845 | 7,464,864 → 6,761,632 (-9.4%) |

These observations support lower process-wide allocation for this campaign, not
a consistent frame-latency improvement. Both builds fail p95 ≤4 ms / p99 ≤6 ms.
Per-run distributions, maxima, source callbacks, geometry and end-only RSS remain
in the raw records. Process allocations include inspection/output and concurrent
runtime work; they are not framework-only or native/GPU allocation counts.

Separate baseline/candidate differential Go allocation profiles are diagnostic
only. They include observers and profile encoding and are excluded from ordinary
timings. Sampled profile attribution cannot establish retained-memory savings,
physical transfers, GPU completion, presentation or input-to-photon latency.

## Evidence and review

Internal source review checks the exact three candidate leaves; protocol review
checks child ownership and the matched workload; independent evidence review
checks receipts and raw observations. Prism reviews this source/spec/report batch
using Anthropic Claude Sonnet 5.5. The [checkpoint](records/checkpoint.json),
[archive](records/evidence.tar.xz) and [member index](records/evidence-index.json)
bind sources, producers, receipts, profiles, raw logs and review results.

See the [checkpoint contract](../../../specs/macos-service-performance.md).
Selected source/prepared/binary identities are not a complete compiler/SDK/cache
closure. Human input/VoiceOver, physical menu/tray use, notification delivery,
cross-display, GPU loss, sustained P6 performance, integrated P11/endurance and
distribution/release qualification remain open.
