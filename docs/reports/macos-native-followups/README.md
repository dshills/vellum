# Unattended Mac native follow-ups

The authorized 14 October 2026 campaign completed scoped native lifecycle and
renderer measurements. These results retain their exact historical source
endpoints; they do not certify the latest framework or complete the Mac milestone.

The interacting composed lifecycle run on `4cf9e36` with 6,209 prepared leaves
passed 1,000 fresh windows and 24,000 checked steps, including two color edits
per window, 1,000 dense labels and 18 completed tree accessibility rows. It took
9,649.522 seconds. All final main-thread cleanup counters were zero and callbacks
stopped. Forty-one forced-GC samples were recorded. The 37 warm samples at/after
100 windows ranged from 44,134,200 to 50,910,592 bytes of Go heap and 327,024,640
to 383,664,128 bytes of process RSS. The final heap profile is preserved. The
512 MiB Go heap and 1 GiB RSS limits are safety guards, not qualification of the
professional memory budgets or a measurement of all native/GPU allocations.

The 24,000 completed steps required 24,011 attempts/admissions. Eleven steps
retried once; their cause was unobserved. Queue admission is distinct from the
exact accepted callback and native-publication checks. Action wall time includes
inspection, waits and owner/native publication; it is not input/frame latency.
This complete campaign follows the prior 343-window failure, whose triggering
close path remains unestablished; the new pass does not assign that cause.

Stage 59 compares `4cf9e36`/6,209 leaves with `9686e38`/6,210 leaves in CPU and
opt-in GPU modes: 80 processes and 9,600 ordinary samples across fixed/variable
10k/million-row cases. Stage 60 compares `9686e38`/6,210 with `68c3750`/6,211 in
opt-in GPU mode: 40 processes and 4,800 samples. Each case has five interleaved
original/candidate pairs, 128 warm-up frames and 120 recorded frames per process.
Geometry remains exact, and four separately gated native GPU roots pass without
skips on each candidate. All arms miss the unchanged 4 ms p95 / 6 ms p99 targets.
Native effects are mixed; local capture gains are not whole-frame gains.
Stage 60 GPU/variable-10k pooled p99 increased from 6.799 to 7.900 ms
(1.101 ms, about 16.2%). Stage 59 GPU/variable-million pooled p99 increased from
6.732 to 7.266 ms (0.534 ms, about 7.9%). These observed tail regressions remain
unexplained; the data do not establish that they are noise or assign a cause.
The stage-59 original fixed-10k CPU arm includes a 49.511 ms maximum with no
assigned cause. No physical presentation or compositor pacing claim follows.

Full P4/P6/P11, current-source composed endurance, overnight background soak,
physical input/display transitions, human VoiceOver, native/GPU memory budgets,
service interoperability and Mac distribution remain open. The GPU scene path
remains opt-in. Historical paused-test notes do not revoke the current human
unattended-test authorization.

Pooled UI-work timings in ms (600 samples per arm/case):

| Stage | Mode/case | Original p95/p99 | Candidate p95/p99 |
| --- | --- | --- | --- |
| 59 | cpu/fixed-10000 | 7.891/8.458 | 7.759/8.288 |
| 59 | cpu/fixed-1000000 | 7.439/7.691 | 7.385/7.835 |
| 59 | cpu/variable-10000 | 7.802/8.251 | 7.933/8.336 |
| 59 | cpu/variable-1000000 | 7.587/8.042 | 7.504/7.959 |
| 59 | gpu/fixed-10000 | 6.279/6.733 | 6.189/6.580 |
| 59 | gpu/fixed-1000000 | 5.863/6.606 | 5.832/6.150 |
| 59 | gpu/variable-10000 | 6.330/7.002 | 6.279/6.746 |
| 59 | gpu/variable-1000000 | 5.915/6.732 | 6.045/7.266 |
| 60 | gpu/fixed-10000 | 6.121/6.440 | 5.962/6.347 |
| 60 | gpu/fixed-1000000 | 5.722/6.066 | 5.712/6.386 |
| 60 | gpu/variable-10000 | 6.262/6.799 | 6.386/7.900 |
| 60 | gpu/variable-1000000 | 5.953/6.898 | 5.887/6.749 |
