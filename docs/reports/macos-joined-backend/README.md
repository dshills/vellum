# Joined CPU/native backend timing: native result

The single approved capture passes complete joined accounting in **37.509 s**
overall / 31.206 s native. The original Cogent/WebGPU design continues.
Instrumented receipt-to-actual-presentation p95 is **36.329 ms**;
34 of 120 measured observations exceed 33.3 ms. No production repair or ordinary
release qualification follows, and the gate remains unchanged.

## Actual p95-ranked input

Nearest rank 114 of 120 selects accepted revision 143, scope 1,
opaque tag 288, generation 146, drawable 151. The parts below
are differences between adjacent, monotonically ordered event timestamps for
that same input. These table intervals are disjoint; they are not independent
quantiles or estimates of time spent exclusively doing one activity.

| Part | Elapsed |
| --- | ---: |
| Receipt to raster start | 0.360 ms |
| Raster | 0.656 ms |
| Raster end to upload start | 0.032 ms |
| Native upload | 2.385 ms |
| Upload end to acquisition | 0.341 ms |
| Acquisition to first render commit | 0.339 ms |
| First render commit to first GPU start | 0.994 ms |
| First-to-final render GPU span | 0.289 ms |
| Final render GPU end to shown | 30.933 ms |
| **Receipt to shown** | **36.329 ms** |

The GPU span includes gaps and overlap between five render buffers. Individual
CPU/GPU activities can overlap inside a table interval without overlapping the
adjacent table intervals. The span is not summed GPU busy time. The signed
present-request offset below is a separate observation, not an additional table
part. An independent audit of the raw endpoint timestamps confirms the partition. The present request occurs **1.138 ms before**
final render GPU end on this input. Its separate present command reaches GPU end
4.125 µs after final render GPU end; **30.929 ms** then remains
until actual presentation. The completed-handler CPU observation is a distinct
clock event, not GPU completion itself or optical input-to-photon timing.

This localizes the dominant remainder beyond the recorded GPU commands. It does
not isolate compositor, driver, display synchronization or scanout, prove a fixed
frame penalty, or establish a WebGPU-versus-direct-Metal cause. More immediate
present requests or faster shaders are not justified as fixes by this capture.
A next hypothesis should concern present/display scheduling within the current
backend, using a bounded control; this result grants no new live attempt.

## Accounting, checks and review

All 248 inputs (128 warmup/120 measured) join the existing absolute raster/upload
records to actual native drawable feedback and complete command histories.
There are 1542 backend records: five render buffers plus one present buffer per
cohort frame, and 45 commands/9 presents explicitly retained before the first input.
The 256 raster and 254 upload records retain 8 and 6 startup records respectively.
No unmatched late submissions or overwritten records are accepted. Full clocks,
command IDs, actual p95 ticket and parent validations remain in indexed evidence.

Nine owned command groups retire. Native stderr is empty; configuration, geometry,
compiled source/binary/native/module/overlay pins and all 8,321 SDK files verify
unchanged. The native archive combines the exact reviewed upload and backend
observers in fresh copies without changing allocation/cache/render/pacing policy.
The archive is not installed into the SDK or packaged app. An initial build chose
an incomplete offline cache and stopped before compilation; the corrected build
and all logs are retained.

Six focused classifier methods/19 negative mutations pass, including a render with
an unrelated intercohort present. The original routed complete-pixel/state fixture
passes with the raster clock; prior five recorder race methods are reused by
byte-identical source comparison. Fresh native/Go builds, exact recipe reproduction
and full log/pin/ownership verification pass. Prism Anthropic Claude Sonnet 5.5
review of 13 preparation files identified the command-ownership gap; it is fixed.
The six-file follow-up's conditional registration concern is dispositioned against
the actual explicit-path import graph. Complete reviews, source audit and
regression evidence are retained in the preparation archive.

The [protocol](../../../tools/macos-calibration/joinedbackend/PLAN.md) and
[preparation checks](preparation-checks.json) describe this private diagnostic.
Preparation source checkpoint is `c2345f9`. Extra observers perturb
execution; there is no overhead calibration, shipping-binary equivalence,
physical input, displayed-pixel readback, whole-GPU allocation balance or release
acceptance. Do not interpret differences from earlier captures as a speedup or
regression. Features/defaults/SDK/app and 33.3 ms gate remain unchanged; M3–M7 open.
The result-data review completed with no findings. Targeted documentation reviews
led to clearer numeric spacing and an explicit distinction between overlapping
activity and disjoint endpoint intervals; the raw partition audit verifies the
latter. Complete reviews and dispositions accompany the result evidence.
The desktop attempt is finished and no native job remains active.
