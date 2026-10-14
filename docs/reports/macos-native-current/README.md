# Mac composed retention and native frame evidence

This checkpoint keeps three tested endpoints separate. The composed retention run
uses `e7f8e30` and 6,214 prepared files. The span-buffer comparison uses `87e7de1`
and 6,216 prepared files. The later ordinary geometry comparison uses `8c1709e`
and 6,216 prepared files. A later commit does not inherit earlier runtime results.

## Composed retention

The native lifecycle consumer completes 1,000 windows and 24,000 owner/native
activity steps, with 1,000 dense labels fully contained in each fixture's observed
native clipping geometry. Forty-one forced-GC
samples include the initial sample and every 25 completed windows. The run takes
9,404.36 seconds. From 100 completed windows onward, sampled Go heap is
44,163,024–50,707,512 bytes and RSS is
370,753,536–398,311,424 bytes. The last sample reports Go heap 50,474,304 bytes,
RSS 398,311,424 bytes, Go system memory 255,408,440 bytes and six goroutines.
The survivor profile is 450,135 bytes with SHA-256
`419805dbe0ceaf6bbce90d91a3a06d17412c31b4f4b154c09a0de8561c564da8`.
It is captured after the cycles, before global application closure.

Every checked step reaches its owner/native completion predicate. Sixty steps
require two attempts/admissions: 54 tab steps and six disclosure steps. Total
attempts and queue admissions are 24,060. Admissions are not accepted callback
counts, and retry causes were not observed. This run does not explain the earlier
343-window failure or the different retry counts in earlier source versions.

Final native cleanup runs on the main thread with callbacks stopped, zero tracked
windows, file panels, native outstanding work, accessibility windows and pending
callbacks. Menu/tray callback counts are zero. The menu was not owned, and its
`OriginalRestored` flag is false; this run does not qualify application-menu
restoration. The outer run and validator terminal receipts both exit zero after
owned-group retirement. Separate post-terminal group observations find no members.

The 512 MiB heap and 1 GiB RSS guards are safety bounds for sampled observations.
They do not qualify professional retained/peak process, native allocator or GPU
memory budgets. Internal review reconciles all 25,042 JSON events and exact
recorded build/run/producers, profile and cleanup bindings.

## Span-buffer native comparison

Forty fresh opt-in GPU scene processes collect 4,800 ordinary events across fixed
and variable 10k/million-row fixtures, five interleaved original/candidate pairs,
128 warm callbacks and 120 measured events per process. The original is an
old-builder overlay against the same application, not a separately built pristine
checkout. All paired geometries and bounded row-count workloads match.

| Case | Original p95 / p99 ms | Candidate p95 / p99 ms |
| --- | ---: | ---: |
| fixed 10k | 6.316 / 6.696 | 6.302 / 6.819 |
| fixed million | 5.907 / 6.369 | 5.996 / 6.288 |
| variable 10k | 6.279 / 6.800 | 6.223 / 6.660 |
| variable million | 5.871 / 6.533 | 5.880 / 6.807 |

Every run misses both unchanged 4 ms p95 / 6 ms p99 thresholds. Allocation p50
falls 1.00–1.74%, and malloc p50 falls 3.05–3.07%. Timing is mixed; variable-million
p99 rises 4.193%. Neither lower allocations nor these elapsed differences establish
a consistent frame improvement or its cause. Local software-capture allocation
savings retain their distinct earlier workload and are not native frame savings.

The raw native-summary binary objects retain `launched=false` from their pre-run
input snapshot. Those fields do not describe terminal launch state. Forty raw
runtime argv/result receipts establish the actual launches. The outer controller
exits zero, kills/reaps its owned group and preserves its exact producer/input
bindings; the successor separately observes that recorded predecessor group missing.
The mode flag is an opt-in renderer request, not proof of all physical GPU work,
presentation or native/GPU memory. Default CPU mode is outside this comparison.

## Ordinary geometry comparison

The later comparison completes 80 fresh processes and 9,600 ordinary events.
Both default CPU and opt-in GPU modes retain five interleaved pairs per each of
the same four cases, 128 warm callbacks and 120 measured events. No profiling or
source-detail instrumentation is active. Original behavior overlays only the old
adapter and removes the new helper against the same application and dependencies.
It is not a separately built pristine checkout.

| Mode / case | Original p95 / p99 ms | Candidate p95 / p99 ms |
| --- | ---: | ---: |
| cpu fixed-10000 | 7.742 / 8.200 | 7.730 / 8.155 |
| cpu fixed-1000000 | 7.427 / 7.795 | 7.362 / 7.680 |
| cpu variable-10000 | 7.792 / 8.160 | 7.732 / 8.020 |
| cpu variable-1000000 | 7.543 / 7.879 | 7.473 / 7.747 |
| gpu fixed-10000 | 6.136 / 6.521 | 6.160 / 6.592 |
| gpu fixed-1000000 | 5.810 / 6.202 | 5.784 / 6.122 |
| gpu variable-10000 | 6.319 / 6.849 | 6.246 / 6.656 |
| gpu variable-1000000 | 5.831 / 6.567 | 5.881 / 6.894 |

No run passes p95 or both unchanged thresholds. Two GPU runs pass p99 alone;
no CPU run passes p99. Timing remains mixed; GPU variable-million
p99 rises from 6.567 to 6.894 ms, approximately 4.97%. CPU fixed-10k process-wide
allocation p50 rises from 5,892,768 to 6,363,168 bytes, approximately 7.98%, while
other allocation observations vary. These observations do not establish a cause,
a consistent latency improvement or framework-only allocation. The prior sampled
geometry profile attribution is not a prediction of elapsed frame savings.

The native controller exits zero, and the reviewed launcher runs its unconditional
owned-group kill/reap path. This geometry campaign has no separate post-terminal
group-absence observation or native-resource inventory. Its completed independent
review reconciles every paired geometry, row-count workload and raw distribution.
The consumer binary digests were captured after their
successful compile receipts: candidate bytes match the earlier native acceptance,
while original bytes were captured before this comparison. Recorded source,
prepared, overlay and binary identities do not attest the complete compiler/cache
closure or loaded executable bytes. Physical GPU work, presentation timestamps
and native/GPU allocator budgets remain separate.

## Review and remaining requirements

All three completed evidence sets have independent internal reviews. Commit requires
final internal source/evidence review and one Anthropic Claude
Sonnet 5.5 Prism review. The [qualification](records/qualification.json),
[archive](records/evidence.tar.xz) and [member index](records/evidence-index.json)
will bind exact sources, raw observations, producer versions and reviews before
commit. No runtime implementation changes are included in this evidence checkpoint.

Full P4/P6/P11, memory/native GPU budgets, final-source endurance and overnight
soak, physical input/displays, human accessibility, service interoperability and Mac
distribution remain open. The existing scoped source, queue and selector checks
do not redefine those requirements.
