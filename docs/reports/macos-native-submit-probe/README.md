# Native submission topology result

The three predeclared native-only combined/split pairs are valid. **Separate
presentation submission did not reproduce Vellum's extra refresh-period latency
band.** Reject this difference as a sufficient cause in this fixture; do not
rewrite the backend or change the release gate on this evidence.

The pinned wgpu-hal source creates a separate presentation command buffer. The
preceding native reference combined render and presentation. Both modes here use
the same cached-texture blit, pixels, geometry, native event receipt, observer,
FIFO/vsync, two drawables, disabled transactions/timeouts, and 120 ms request
schedule. Only command-buffer submission topology differs. The new observers
are identical in both modes and were not present in the preceding calibration.

| Pair / order | Combined p95 | Split p95 | Split minus combined p95 |
|---|---:|---:|---:|
| 1: combined/split | 32.033 ms | 32.073 ms | +0.040 ms |
| 2: split/combined | 32.044 ms | 32.020 ms | -0.024 ms |
| 3: combined/split | 32.152 ms | 32.087 ms | -0.065 ms |

These are differences of per-run p95 values, not percentiles of per-frame
differences. Do not pool populations or add phase quantiles. Render GPU duration
p95 is 0.110–0.145 ms; render commit to GPU start p95 is 0.523–0.551 ms.
GPU completion to actual presentation p95 is 30.697–31.115 ms. Scheduled/completed
callback timestamps are CPU observation times, separately labeled. Tail maxima
35.438/45.419/45.511 ms (combined) and 35.356/32.343/49.189 ms (split) remain in
the populations; similar p95 values do not mean every response meets 33.3 ms.

Execution on 21 October 2026, 20:39:33.136–20:42:41.553 UTC
(15:39:33–15:42:41 Chicago), lasted 188.417 seconds within the 240-second bound.
All six runs account for 768 warmup and 720 measured accepted states. Twelve zero
startup presentation feedback records remain retained; cohort clocks/statuses
pass the unchanged parent classifier plus strict render GPU and commit checks.
All six native and 48 configuration-probe process groups terminate and retire.
SDK 8,321-file inventory, source/binary/match/log hashes, configuration and
persistent geometry remain unchanged. The locked match is reused byte-for-byte
from the completed calibration; no framework/SDK rebuild or app relaunch occurs.

Preparation passed a warning-free native build, 18 unchanged parent classifier
controls and seven submission classifier methods with nine invalid timing
subcases. Internal review and Prism using Anthropic Claude Sonnet 5.5 preceded
execution; both preparation findings were disposed with source-backed lifecycle
and exception-path analysis. Result review and exact raw-journal revalidation
are retained with indexed compressed evidence. Source checkpoint: `5dd9250`.

This sparse native texture-blit experiment does not qualify Vellum, a 60 Hz
changing workload, professional workflows, physical input, pixel readback,
endurance, native allocation balance, resize, or distribution. The earlier
framework misses remain failed. Keep features frozen and the 33.3 ms receipt to
actual-presentation release gate unchanged. The next useful measurement is the
framework backend's actual render submit/GPU-complete/present-request boundary,
with observer impact and source-built library identity explicitly controlled.
The separate-present-buffer hypothesis is closed as insufficient; no unchanged
retry is warranted.

See the [protocol](../../../tools/macos-calibration/submit/PLAN.md),
[exact per-run summary](summary.json), [source audit](source-audit.json), and
preparation/result evidence indexes. Executables, SDK bytes and credentials are
excluded from evidence archives. No release gate is closed.
