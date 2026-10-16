# Current native grid measurements

The ordinary optimized consumer on `77ca764b2a7896058cc21872a742f3a57bdf058c`
completed 40 fresh native processes: CPU raster first, then the optional GPU
scene renderer, each with fixed/variable heights and 10,000/1,000,000 stored rows.
Each case used five processes, 128 warmup steps and 120 measured samples per
process, for 4,800 samples. The exact source vector, 6,227 prepared files,
`go build -trimpath` output association and binary hashes are retained with the
[composed campaign records](../macos-composed-campaign/records/qualification.json).
These are current-configuration observations, without an optimization control
arm or causal before/after claim.

All eight cases miss the unchanged 4 ms p95 / 6 ms p99 routed UI-work targets:

| Renderer | Heights | Stored rows | Pooled p95 (ms) | Pooled p99 (ms) |
| --- | --- | ---: | ---: | ---: |
| CPU | Fixed | 10,000 | 7.887 | 8.320 |
| CPU | Fixed | 1,000,000 | 7.491 | 7.986 |
| CPU | Variable | 10,000 | 7.874 | 8.441 |
| CPU | Variable | 1,000,000 | 7.630 | 7.979 |
| GPU scene | Fixed | 10,000 | 6.601 | 7.119 |
| GPU scene | Fixed | 1,000,000 | 6.082 | 6.810 |
| GPU scene | Variable | 10,000 | 6.186 | 7.004 |
| GPU scene | Variable | 1,000,000 | 5.696 | 6.867 |

The 341.146-second campaign completed with stable observed inputs, no stop or
cleanup errors, joined stop watcher and one recorded retirement per owned native
process. The validator reparsed raw records, recomputed timing/allocation/source
distributions and row-count geometry, checked original case results, and freshly
observed all 40 owned process groups absent. Observers run after measured native
work; group-absence observations have their recorded time scope.

The controller first required the completed 1,000-window trace and corrected
40-second soak preflight, with their exact source contexts and retired groups.
A shared lease serializes native campaigns. The proof-comparison correction
permits only a valid regenerated UTC observation timestamp to differ; every
other endurance-proof field must match. Six CPU-only controls check real owned
group presence/retirement, observer failures, serial admission, changed input
digests and stable-proof identity. Earlier controller revisions and findings are
retained separately. No incomplete predecessor is treated as a pass.

The measurements use ordinary unprofiled paths with diagnostic details disabled.
UI work through submission does not measure physical display presentation or
human input latency. Process/binary/source observations are not hermetic
compiler/cache/SDK attestation, native/GPU allocation or professional memory
budgets. Overnight soak, complete P6/P11 and full Mac release acceptance remain
open. A complete measurement trace does not qualify a failed timing target.
