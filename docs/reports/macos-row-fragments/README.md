# Managed-row paint fragment feasibility

The private retained-row prototype on clean `b0534e9` is **rejected at the reuse
feasibility gate**. It passes correctness checks, but the actual GPU diagnostic
shows body-row reuse in only **12 of 120 measured frames**: **588 hits / 11,760
row opportunities (5%)**. The remaining 108 frames have zero hits and 98 misses
each; twelve have 49 hits and 49 misses. Owned fragment/proof storage reaches
**8,387,889 of 8,388,608 bytes**, with at most 612 entries and one retained font.
This does not demonstrate useful reuse for a material p95 improvement.

There is **no ordinary latency comparison or measured latency rejection**.
The CPU app also exits zero with 120 valid samples, but its observer is absent:
the existing benchmark runner filters `VELLUM_ROW_FRAGMENT_DIAGNOSTIC` from its
child environment. The combined controller fails with `one matching observer
per source`. CPU reuse and CPU performance remain unverified. Preserve that
failure; do not repair the observer and continue an already rejected prototype.

Stop the whole experiment before the 90-minute reassessment limit. Do not run
the planned ordinary/novel comparisons, enlarge the cache, relax exact-coordinate
proof, or extend this experiment into a translated-row design. No production
or prepared dependency file changes. No Mac acceptance gate closes.

## What was implemented privately

An immutable scene fragment detaches an interval of the existing paint builder
at its exact original coordinates and clip. CPU replay and GPU encoding retain
original painter order, alpha rules, partial-target behavior, resources, shader,
transport and admission charges. Mutable resources refuse retention. Pending,
custom or dirty renderer state follows the ordinary per-row path before mutation.

Reuse compares copied current text/shape/glyph/font/paint state and geometry.
A bounded content hash only narrows the lookup; full equality remains mandatory.
Borrowed layout pointers or model revisions cannot authorize a hit. Effective
blend operation is part of the key; nil base color refuses reuse. Tests cover
same-pointer mutation, hash collisions, A→B→A, pending/custom paths and mixed
math/bitmap/path rows. No translation or fractional-phase equivalence is assumed.

The cache bounds owned fragment/proof storage to 8 MiB and 1,024 entries, with
at most eight strong immutable font references. This is not a total process/font
memory bound. Resize/release retire its private state. GPU hits still charge
original work and outline budgets before appending original operations.
Internal review found and resolved blend, nil-color and per-row renderer-state
issues before the final pinned checks and first native admission.

## Verification and retained failures

| Final pinned check | Result |
| --- | --- |
| Sceneplan/rasterx/gpuscene offscreen race | PASS, 19.641 s; native GPU tests explicitly skipped here |
| Managed-grid paint/mutation and warm fixed/variable rows at 96/160/192 DPI | PASS, 36.882 s |
| Focused guard/budget/mutation/eviction regressions | PASS, 2.911 s |
| Renderer vet | PASS |
| Ordinary and separate diagnostic release builds | PASS |
| Actual native GPU suite, explicitly admitted | PASS, 2.420 s; no skips |
| GPU diagnostic, fixed 10k rows | 120 correlated accepted sources, no fallback or previous-frame reuse; inadequate body-row fragment reuse |
| CPU diagnostic | App exit 0 / 120 samples; observer correlation FAIL, reuse unavailable |
| Ordinary CPU/GPU comparisons and novel cases | NOT RUN; feasibility rejected |

Native GPU tests compare cold/warm positive fragment hits against uncached CPU
pixels and actual readback. A later partial batch over retained contents must
differ from a fresh target; five writes, exact byte ledger and one submit remain
unchanged. These checks establish neither physical interaction nor presentation,
ordinary latency, memory or endurance qualification.

The initial grid run fails an old optimization-specific glyph-placement-hit
assertion: row reuse bypasses that cache. Its six expensive serialized integer/
pixel tile fixtures and paint tests pass. The revised guard accepts an actual
row-fragment hit or the original placement hit, preserving pixel and warm-byte
assertions. The affected warm/paint suite passes; unchanged expensive fixtures
are reused. The initial controller NameError before test launch and intervening
checks are retained separately. Test-generated PNGs are preserved outside the
prepared tree; only those generated leaves were removed after proving all
original source bytes identical. Both trees still contain exactly 6,260 prepared
files with aggregate `1d9861ac10a0d83f607a36fd70dc79f00ce51d427c792b30b6a1331c0c3451ba`.

## Evidence and next action

[records.json](records.json) hashes each member of [records.tar.gz](records.tar.gz).
It includes the rejected source/tests, pinned plans and binaries' hashes (not
copies), commands/logs, raw diagnostics, failures, internal reviews and the
single scoped Anthropic Claude Sonnet 5.5 Prism checkpoint (three documentation
files, one fresh call, zero retained/discarded/suppressed findings). `assessment.json`
records the rejection; `diagnostic/gpu/row-reuse.json` preserves all correlated
observations. CPU raw samples remain in `diagnostic/cpu/`; their timings are
diagnostic only. Final post-native proof verifies source pins, clean checkouts
and identical prepared trees. No SDK or executable is copied into this report.

The fresh live lease was held through case cleanup and released; case stops are
sealed. Both diagnostic app invocations
exit zero and the runner attempts owned-group cleanup; this is not strict PID
absence proof. No physical pointer, input-source, VoiceOver, clipboard or
preference change is part of this experiment. Placement operator retries remain
prohibited; its required native acceptance is direct/manual.

M3 needs a separate architecture decision against the original frame budgets,
not more local AX/upload/culling/row-cache tuning. While that decision is open,
continue M6 with an actual minimum-Go-toolchain compatibility check: the module
declares Go 1.25.6, whereas recent consumers used Go 1.27.2. The legacy JSON
experiment on the newer compiler does not qualify the declared minimum. Native
M2, M3–M5, final SDK/rights closure in M6 and M7 remain open.
