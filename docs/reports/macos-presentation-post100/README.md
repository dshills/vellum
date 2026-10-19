# Source100 native responsiveness checkpoint

The four-case native feedback preflight passes journal validation on committed source `3d426825da9656501332a4e45c99185d3493ab97`. Both paced workloads fail the measured 60 Hz release gate. Mac responsiveness remains open.

| Workload | Measured inputs | Input-to-presentation p95 | Final-state interval p95 | Intervals within 18.333 ms |
|---|---:|---:|---:|---:|
| Serial grid, CPU scene | 120 | 64.318 ms | Serial diagnostic | Not a paced cohort |
| Serial tree, GPU requested | 120 | 65.159 ms | Serial diagnostic | Not a paced cohort |
| Paced grid, GPU requested | 600 | 65.478 ms | 33.33346 ms | 2/599 (0.334%) |
| Paced tree, CPU scene | 600 | 65.141 ms | 33.33342 ms | 3/599 (0.501%) |

The unchanged release policy requires at least 99% of changing final-state intervals within 1.1 display periods and input-to-presentation p95 at most 33.3 ms. Both paced cohorts took about 20 seconds against a 10-second changing-state schedule. Structural capture success is not responsiveness acceptance.

The source100 paint retry correction removed most unarmed companion acquisitions: the earlier paced-grid prefix had 1,451 acquisitions/721 unarmed entries, versus 735/7 here. The earlier source99 journal remains a refused FAILED_DIAGNOSTIC; its raw counts are descriptive comparison evidence only. Its original failure and records have not been reclassified. This run prospectively retains two valid pre-input skipped startup acquisitions and requires all 600 measured changing states and 128 warmup states to present.

The current private SDK was packaged from clean source 3d42682 with the owned source-built GPU library. Archive `9a1dca6bc39dca096be3cc5edc9e670455d22bbecc0e05670448ff47a0c0646f` and manifest `4c8e11aa2d62bf4fa0320cc663b78780f0ba2be8a6305b524e90b19ba21fa163` bind 8,276 retained files. The genuine relocation consumer/public-contract tests, affected race/vet checks, optimized native builds and strict presentation-validator checks pass. The public owned-build collector also finds all three selected driver weather SVGs and their original root/Apache/NOTICE candidates automatically; no manual supplement was used for that inventory check.

Actual native commands ran serially with the exclusive live lease on the 60 Hz desktop after all background compilation stopped. The controller preserves complete source/prepared/SDK vectors, binary metadata, original journals, display/power probes, geometry guards and owned process receipts. It leaves OS input sources, VoiceOver, preferences and personal application data unchanged. Automated framework routing and native presentation feedback are the evidence here; physical pointer delivery, visual fidelity, human accessibility, native deallocation, endurance and the full performance matrix are separate remaining gates.

The next focused experiment removes a redundant synchronous native-main-thread cursor handoff only when the same live window has already successfully applied the exact raster/generation with no intervening cursor request. Timing causality is unproven until a fresh unchanged-workload native comparison. The release thresholds and measured populations remain unchanged.

Raw journals and owned receipts remain locally retained in the source100 campaign. The compact evidence record binds the full independent review and original files by SHA256. GPU request flags are externally bound to genuine optimized binaries; the journal alone does not certify a GPU rendering route.
