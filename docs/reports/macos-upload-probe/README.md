# Private CPU image-upload stage result

The one predeclared off/on pair is valid. **CPU copying into staging is the main
measured native upload interval; the explicit create-buffer call is much smaller.**
This locates a boundary to investigate, not a proven cache policy, memory pool or
shipping performance fix.

| Observer-on CPU interval per measured ticket | p50 | p95 | Maximum |
|---|---:|---:|---:|
| Create buffer | 0.102 ms | 0.117 ms | 0.175 ms |
| Map | 0.001709 ms | 0.002167 ms | 0.002375 ms |
| Copy into staging | 2.061 ms | 2.612 ms | 3.115 ms |
| Flush/unmap | 0.000834 ms | 0.001167 ms | 0.001500 ms |
| Encode/queue ownership | 0.063 ms | 0.070 ms | 0.652 ms |
| Other native CPU work | 0.052 ms | 0.062 ms | 0.089 ms |
| Native write function total | 2.287 ms | 2.842 ms | 3.894 ms |
| Go image API outside native function | 0.020 ms | 0.026 ms | 0.045 ms |

The intervals are nested in the enclosing Go image API. Disjoint native fields
sum within their own function span, not across independent quantiles. Signed
residuals remain in raw data; do not add table percentiles. The copy interval can
include first-touch backing, page faults, cache effects or scheduling. A small
create call does not prove physical allocation is free, nor that buffer reuse
could not help. No allocation pool is justified or adopted by this result.

Image API p95 is 2.746 ms with recording off and 2.874 ms on; the difference of run
p95 values is +0.128 ms. This single private pair does not prove observer neutrality,
calibrated overhead or equivalence to the retained original library/binary.
Both modes retain the same detailed Go source clocks. CPU function success means
returned Ok, not successful/physical GPU transfer, GPU completion or deallocation.

Observer-on retains 254 native call records: six startup writes and 248 uniquely
joined warmup/measured writes. Each measured frame has one 7,692,216-byte RGBA write,
1519×1266 with 6076-byte staging pitch. Native calls between receipt and the final
armed acquisition match each ticket's complete prefix/final image API attempt
count and logical bytes, including all 128 warmups and 120 measured states. Every
remaining write completes before first receipt. All nonempty writes observe all
five stages, with no duplicate stage, overflow, incomplete slot or unknown cohort
write. Startup rows and all tails remain retained.

The 21 October 2026 campaign runs 22:13:55.736–22:15:02.371 UTC
(17:13:55–17:15:02 Chicago), 66.635 seconds within 90 seconds. Both native and 16
probe groups terminate/retire. All source/native/binary/overlay/match/log hashes,
configuration, persistent geometry and immutable SDK 8,321-file inventory pass.
The private source derivative preserves original validation, buffer descriptors,
errors, create/map/copy/flush/unmap/encode order, pending-write ownership and GPU
completion lifetime. No handle/pointer is stored in the fixed 4096-row observer;
no pool, cache mode, renderer/pacing/default policy or SDK/app bytes change.

The [protocol](../../../tools/macos-calibration/upload/PLAN.md), [build identity](build-identity.json)
and preparation evidence bind exact locked wgpu-core 29.0.3 and native source,
private archive, three Go overlays and forced archive-derived Go link identity.
Seven new wrapper methods (eight bad-record/four terminal subcases) pass. Parent,
source-cost scalar, exact view/routing/pixel and stage106 Core/race/preparation
proof are reused. Internal and Prism Anthropic Claude Sonnet 5.5 review complete
with no preparation findings before source checkpoint 0640a2e and execution.
Actual journals and all 18 owned process groups are independently revalidated.

Next use an offscreen native first-versus-subsequent-copy/cache-mode control with
exact GPU byte readback before considering any policy or ownership change. The
existing Metal HAL chooses shared, write-combined storage for map-write buffers;
its pitch alignment is 4, not 256. The current data do not establish whether cache
mode, deferred page backing or another copy effect dominates.

Features stay frozen, experimental rendering/pacing remains disabled and the
33.3 ms receipt-to-actual-presentation gate is unchanged. No ordinary latency,
60 Hz professional workflow, physical input/pixel presentation, endurance,
allocation balance or distribution gate closes; M3–M7 stay open. See [exact summary](summary.json)
and indexed evidence archives. Libraries, executables, SDK/cache/target trees,
credentials and prior archives are excluded.
