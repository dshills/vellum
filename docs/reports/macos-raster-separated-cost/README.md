# Offscreen replay conditions: hot-loop timing is not native raster cost

A single two-second condition control shows substantial variation with unchanged
rendering code. The earlier0.17ms hot microbenchmark cannot stand in for native
raster timing at120ms request intervals. This establishes a measurement limit;
it does not identify cache behavior or provide a production repair.

| Condition, eight observations each | Min | Median | Max |
| --- | ---: | ---: | ---: |
| Immediate after opposite-color primer | 0.178 ms | 0.421 ms | 0.710 ms |
| After120ms pause | 0.337 ms | 0.572 ms | 0.902 ms |
| Immediate with32MiB scratch touch | 0.269 ms | 0.370 ms | 0.441 ms |
| Pause plus32MiB scratch touch | 0.322 ms | 0.627 ms | 0.728 ms |

Eight rounds rotate the starting condition, retaining all32raw values. Each
condition has one persistent renderer and detached commands; an untimed opposite-
color full replay primes it before the timed color change. The primer color may equal the previous round’s final color; “opposite” refers
to the following timed target. Direct frozen raster replay still performs the
image writes, without the scene-source/GPU unchanged-image cache. This source
assumption was checked during review and retained in the evidence. Optional pause
and scratch touch precede timing; recorded actual gaps include those preconditions.
Scratch touching each64-byte region imposes a footprint, not a hardware cache
flush. It also performs CPU work and may alter scheduling/frequency. The values
overlap; this short run does not isolate a cause or establish repeatability.
No native p95, difference significance, speedup or cost share is inferred.

The single new focused Go method finishes final full-image comparison against
legacy replay for allfourrenderers after measuring. Existing routing/full-pixel
fixture and warm helpers are unchanged and reused. One new parser method checks
nine refusals covering scope/partial data/order/duration/separation/pressure/color.
CPU2, target1519×1266, all32complete samples and bound1.998s<10s pass. The entire
owned Go test command takes7.351s, including preparation/build; its group retires.
No new hot benchmark, desktop application, physical input or native presentation
check occurs.

Independent replay verifies complete source/module/overlay/log pins, every private
Core file equal to frozen originals and all8,321SDKfiles unchanged. The same
manual/native color and path-coordinate limits from the
[warm checkpoint](../macos-raster-offscreen-cost/README.md) remain. Raw values,
configuration and claim boundaries are in [summary.json](summary.json); sources,
receipts, validation and Prism review are retained in [indexed evidence](evidence-index.json).

The [protocol](../../../tools/macos-calibration/rastercost/PLAN.md) keeps features,
renderer/SDK/app/native archive/defaults and33.3ms release gate unchanged. No
no-paint elision or cache policy is adopted. A future timing diagnostic must retain
same-ticket renderer plus surrounding upload/schedule boundaries rather than
subtracting independent hot microbenchmarks. Cache eviction, CPU frequency,
scheduling, GC and native upload sharing remain unseparated. M3–M7 remain open;
this is not macOS release qualification or permission for another native run.
