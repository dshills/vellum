# Source102: failed native comparison; policy remains experimental

Source `04ce5192833b71b04cc2e964e011917e85620b81` completed two valid serial
captures, then failed the paced-grid capture. Paced tree never ran. This scope
qualifies neither cadence nor M3/M4 or a framework release. The failed journal,
raw logs, source/SDK binding and original admission remain sealed and unchanged.

| Workload | Source101 p95 | Source102 p95 | Source102 status |
|---|---:|---:|---|
| Serial grid, CPU scene |34.801 ms|38.657 ms|Valid serial diagnostic; latency fails 33.3 ms|
| Serial tree, GPU requested |52.648 ms|39.502 ms|Valid serial diagnostic; latency fails 33.3 ms|
| Paced grid, GPU requested |52.576 ms|39.623 ms|Failed capture; raw timings unqualified|
| Paced tree, CPU scene |50.431 ms|No run|Unexecuted|

The serial effects are mixed: grid p95 increased 3.856 ms and tree decreased 13.146
ms. Both remain above 33.3 ms. Serial intervals are not sustained cadence evidence.
GPU requested does not establish a scene GPU route.

## Exact capture failure

Paced grid exited 1 after 12.815 seconds with `presentation capture: lost prefix
feedback` reported twice. Its frozen prefix retains 734 acquisition records:
728 armed final presentations, three unarmed/tag0 presentations and three
unarmed/tag0 skipped callbacks. There is no recorder failure, overflow, eviction
or missing armed presentation. All 728 setup/warmup/measured armed records
genuinely presented, including all 600 measured tags.

Two skipped records acquired 18.939 ms and 10.503 ms before the first receipt are
reported as startup observations. Sequence 3, composition 6, drawable 2 has
`Armed=false`, `Tag=Generation=0`, `Presented=0`, and acquisition time
981907.4000455836. The first receipt is 981907.3978588335: acquisition is 2.186750 ms
later. It therefore fails the unchanged strict pre-receipt startup predicate.
The native callback returned zero presentation time, which remains skipped.
The original failure is not relabeled or filtered, and the public validator
continues to refuse `FAILED_DIAGNOSTIC`.

The native layer reads `Acquired` and the current scalar stamp after
`super.nextDrawable`; input receipts occur before FIFO send. An ordinary
composition may have begun earlier, but there is no composition-entry timestamp
to prove this was pre-input work or to identify why Metal skipped the drawable.
This is not evidence of a GPU crash or loss of a tagged measured callback.

## Descriptive timing and integrity

The complete 600 measured armed records in the **failed** paced-grid journal give
p95 39.623 ms latency, 16.666750 ms intervals, 597/599 intervals within 18.333 ms and
a maximum interval 33.333375 ms. These raw distributions grant no cadence or
latency acceptance. CPU work ended at 10.0369485 seconds against a 10-second plan,
a 36.9485 ms overrun; acknowledgement-observation p95 was 17.142209 ms.
CPU elapsed UI/Input/Acquire p95 values were 3.781669/0.0145/0.03825 ms.
Native receipt-to-final-acquire p95 was 16.710375 ms and acquired-to-present
p95 23.294667 ms, compared with 16.972250/36.262708 ms in source101.
Elapsed CPU stages include pauses; they are not thread CPU utilization or GPU
execution. Native-clock phase comparisons do not establish policy causality.

The independent audit made 672 checks. All 23 actually created process groups
(two serial cases with before/after probes, then paced-grid before probes and
its failed native process) retired and were observed absent; the shared lease
was free at audit. Complete clean C/T source and prepared vectors, all 8,285
retained SDK files and actual archive members, plan/review/association pins and
457 historical evidence pins match. Recorded geometry and all observed raw
configuration probes match. Failed-case after probes and the fourth case were
not executed; no observations are invented for them. Worker, runtime and native
scope retirement passed for all attempted cases. This does not prove Cocoa or
Metal resource deallocation or resize/close during blocked acquisition.

The independent reviewer recommends reverting the optional lower-latency request
unless it is retained explicitly for a separately reviewed further experiment:
all observed input-latency p95 values remain over 33.3 ms and this capture failed.
Root retains it only as **EXPERIMENTAL**, while evaluating a separate Core103
busy-retry candidate. That decision adopts no release policy, changes no
validator or thresholds, retries no failed scope and claims no measured effect
for the next candidate. Full release and lifecycle gates remain separate.

[summary.json](summary.json) contains compact metrics and
[evidence.json](evidence.json) binds the full independent audit and original
records. These drafts add documentation for the next shared code checkpoint;
they perform no build, test, native run or probe.
