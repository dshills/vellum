# Local SDK checks and host baseline

This evidence was captured on 2026-10-10 from the SDK implementation staging
checkout, based on `ff4fdb0`, before integration with the other concurrent phases.
It verifies independent preparation and consumption of the existing framework API;
integrated changes need the normal final checks and fresh runtime evidence.

- [Preparation output](prepare.txt): all fifteen Cogent patches and two go-text
  patches applied from driver-owned copies and verified input hashes.
- [Preparation safety tests](preparation-tests.txt): fail-closed input hashes,
  preserved prior installations on second-module failure, and patch ordering/digests.
- [Portable/native checks](check.txt): root race tests and vet, native tests/vet/build,
  and standalone hello module race tests/vet/build passed.
- An application generated outside the repository at a path containing spaces,
  with module path `example.com/independent-ui`, also passed race tests, vet and build.
- [Portable benchmarks](portable-benchmarks.txt): three repetitions of one-million
  record viewport and standalone UI construction/validation, with allocations.
- [Native measurements](native.jsonl): three independent processes each verified
  the GPU rectangle and completed quiet CPU/RSS sampling and timed shutdown.

Native host: Apple M4 Pro, arm64, macOS 26.6.2, Go 1.27.2. The actual readback image
was 3054 × 722; these are not measurements at an assumed 640 × 480 resolution.
Application-entry-to-verified-GPU startup ranged 0.298–0.859 seconds; thirty
synchronous render/readback samples ranged 2.039–2.613 ms. Observed RSS ranged
254.55–256.33 MiB and quiet CPU samples 2.21–3.84% of one core. These values include
the boundaries defined by the [SDK contract](../../../specs/sdk.md), and do not
establish production latency, memory or sub-1% idle CPU guarantees.

The initial sandboxed launch failed macOS graphics-service access. The successful
measurements used approved desktop/process access. No input sources, accessibility
settings or other macOS settings were changed.

Remaining gates: choosing/publishing a framework namespace and versioned native
fork or upstream patch disposition, distributable third-party license audit,
Windows/Linux runtime tests, native input/accessibility qualification, and repeatable
production performance targets on representative hardware.

Subsequent multiwindow ownership correction: driver-owned patch 0016 was added
after live race-instrumented window handoff exposed composer withdrawal and scene
title routing races. Fresh bootstrap reproduces all sixteen Cogent patches and
verifies final file hashes; four targeted upstream regression tests pass with the
race detector. Both native handoff and native services lifecycle smokes also pass
with `-race` against the fixed source. The benchmark records above predate this
correction and are an initial host baseline, not evidence of the final integrated
framework's exact performance.

See [final canonical checks and measurements](final/README.md) for the integrated
framework after all five expansion phases. The earlier records above are
historical staging baselines.
