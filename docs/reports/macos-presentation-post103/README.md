# Source 103: native cadence passes; input latency remains open

Clean source `33183c7706d0791188474aff45db8ff2804573c1` and its private SDK
complete four valid native captures. Both paced cohorts meet the unchanged
numeric cadence predicate; every input-latency p95 exceeds 33.3 ms. Mac completion
remains open. This four-case, single-run comparison is not the full repeated M3
matrix, physical input acceptance or a release qualification.

| Workload | Input-to-presentation p95 | Within 18.333 ms | Evidence scope |
|---|---:|---:|---|
| Serial grid, CPU scene |42.781 ms|119/119|Serial diagnostic; no sustained cadence acceptance|
| Serial tree, GPU requested |43.620 ms|119/119|Serial diagnostic; no sustained cadence acceptance|
| Paced grid, GPU requested |43.832 ms|599/599|Numeric cadence PASS; input latency FAIL|
| Paced tree, CPU scene |41.812 ms|599/599|Numeric cadence PASS; input latency FAIL|

Both paced interval p95s are 16.666750 ms and maxima 16.666792 ms. The full 600 measured
states remain in each cohort. Grid CPU work ended 2.533 ms after its 10-second plan;
tree completed 1.325 ms before plan end and retained the hold. All 36 actually
created native/probe groups retired and were observed absent, with the shared
lease free at independent audit. Original scopes, inputs, worker/runtime closure,
raw logs and strict prefix/clock checks pass. Paced grid retains two proven
pre-receipt startup skips as skips; all armed final acquisitions genuinely present.
No record is filtered or assigned a fabricated display timestamp.

The fresh private SDK has 8,292 files including its manifest. An independent
association audit verifies all archive members/retained bytes and clean C/T source
(5,891 tracked files) plus 6,322 prepared files, genuine package/relocation receipts,
compiler and owned native archive. The actual primary binaries are bound to those
SDK sources. Source/SDK, geometry and all observed display/power/OS fingerprints
remain unchanged across the four native cases. This establishes scoped consumer
and capture evidence, not final packaged `.app`, public distribution or rights
approval.

Paced receipt-to-final-acquire p95 is 16.687/15.662 ms for grid/tree; CPU elapsed UI
work p95 is 3.630/6.000 ms, and native final-acquire-to-present p95 is 27.436/26.900 ms.
Input-stage p95 is 0.010583/0.014791 ms and acquisition-call p95 is 0.024126/0.037334 ms.
Elapsed CPU-side intervals include pauses and are not CPU utilization or GPU
execution. Native phases use the shared host clock without subtracting Go times.

Against validated source 101, paced latency is lower by 8.744/8.619 ms, mainly in the
descriptive acquired-to-present phase. Against **unqualified source 102** raw paced
grid, latency is 4.208 ms higher, receipt-to-acquire differs by -0.024 ms, UI work
is 0.152 ms lower and acquired-to-present is 4.142 ms higher. Source 102 paced tree
never ran. Its failed prefix journal remains failed and is not retrospectively
qualified. These single-cohort differences do not isolate a GPU policy or
completion-triggered-rescan cause. Stage 0102 remains experimental.

The remaining milestone work is concrete:

- M2 previously passed its controlled eleven-method functional scope on `93b4730`;
  that does not substitute for later candidate evidence, physical/human work or
  complete Mac acceptance.
- M3 needs input latency at most 33.3 ms, the existing full repeated matrix and
  representative text/layout/canvas/input/resize workflows. The 4 ms CPU budget
  stays an optimization target.
- M4 needs native/font/GPU resource attribution, required bounded-resource/idle
  acceptance, ten-minute composed activity, 1,000 composed lifecycle cycles and
  the complete eight-hour visible/occluded/minimized/background soak. Earlier
  cleanup/idle and AppKit-reference observations retain their limited scope.
- M5 needs physical pointer/preview and keyboard workflows, IME/human VoiceOver,
  platform-service and supported display/DPI transition acceptance, recorded
  separately from automated Go routing/native selectors.
- M6 needs the final accepted-source SDK/application, component/native/font/asset
  and notice/rights closure and relocated native packaged-app acceptance. Later
  changes require affected evidence renewal; signing/publication is a separate
  explicit external step where required.
- M7 needs the final frozen-candidate requirement audit, support/limitations,
  migration/tutorial and runnable consumer records after preceding gates close.

Two separate existing native lifecycle helpers also passed on this exact source
and SDK. The public client-shrink helper records the requested 640×480 dp window
at 648×486 pixels and then 324×243 pixels on the observed DPI configuration. The
visibility helper retains 30 records, including seven explicit frames, two typed
terminal hidden-frame records, five successful snapshots, five visibility
observations and four native notifications. All 18 created native/probe groups
retired and were observed absent; source/SDK/geometry and configuration checks
pass. These are scoped helper observations, without deterministic proof of a
blocked acquisition overlapping resize/close, physical input, pixels or native
object deallocation. The final Cocoa census retains one hidden, untitled
500×500 `NSKVONotifying_TUINSWindow` (#158118), unclassified. Tracked session,
callback and registry counters reaching zero do not remove that window or prove
its ownership or deallocation; no class is whitelisted.

A separate observer-instrumented grid trace captures 600 measured events with
input-to-presentation p95 47.640 ms, receipt-to-final-acquire p95 16.624 ms,
final-acquire-to-present p95 31.316 ms and elapsed UI-work p95 3.619 ms. One native
and six headless Go trace/profile commands completed and were observed absent.
These trace timings are descriptive and unqualified: instrumentation changes
timing, and the native and Go trace clocks cannot be subtracted directly. The
aggregate profiles place substantial delay under native-main text-input
publication and CGO crossings, making repeated publication/handoff a concrete
source lead for the next narrow cache experiment. Concurrent scheduler, polling
and synchronization totals do not establish per-input causality or CPU/GPU work.
A later parsed chronology export failed its bounded headless command and stays
failed, outside the completed trace/profile audit. No successful per-tag
chronology or cache-fix performance result is claimed. The trace controller
checked geometry equality but did not retain raw before/after geometry values;
the independent audit separately found current geometry equal to sealed primary
source103 observations.

The source103 short M4 preflight has only passed headless prerequisites, including
an explicitly disclosed one-test-file overlay correcting a stale WindowContext
expectation. Its prepared plan was held unexecuted. After the cache and test
checkpoint, it requires a genuine new SDK binding and fresh checks; these
headless results do not qualify the later runtime. No new gate, threshold,
completion percentage or timing acceptance is introduced. [summary.json](summary.json) stores compact current metrics;
[evidence.json](evidence.json) binds both independent audits and original records.
