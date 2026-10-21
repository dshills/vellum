# Default-path macOS responsiveness assessment — 21 October 2026

The first bounded diagnostic round obtained valid native Metal presentation
feedback. **Default-path paced grid cadence passes this single run; input latency
fails. M3 and the Mac release remain open.** The planned stop-on-failure prevented
unchanged repetitions and the fourth case. No framework feature or rendering
policy changed.

## Candidate and procedure

Features were frozen at main `7047740e7d5ae5acf910d367cd3b16545aff2e70`.
The measured runtime is the retained, manifest-verified private SDK from
`bd3cbc9ce2566955dae5dbb61385b0e1a9307eaa`, not a newly packaged main build.
All 8,321 SDK members including its manifest were hashed before and after.
Optimized Go 1.25.6 Darwin arm64/cgo executables link the retained source-built
native archive SHA256
`786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.
Binary identities and build metadata are retained in the evidence. External
manifest/build/launch binding supplies source identity because SDK executables
have no VCS stamp.

The user approved up to 20 quiet desktop minutes. Actual native assessment ran
17:27:31–17:28:02 UTC (31 seconds). macOS 26.6.2/25G83, Apple M4 Pro, one online
LG HDR DQHD at 5120×1440/60 Hz, AC power. The paced fixture reports DPR 1, logical
DPI 162, and a 1519×1266-pixel client. Configuration probes bracket the two serial cases; paced grid has before-only
probes because failure stops subsequent commands. Geometry files are unchanged. Desktop exclusivity is user-coordinated, not an
instrumented assertion that all other applications were idle.

Each consumer uses 10,000 stored records and 128 warmups. Serial cases request
120 inputs; paced cases request 600 states at 60 Hz. The adapter reuses the
existing owned executor, exclusive lease, single-attempt orchestration and the
SDK's strict journal validator. It omits `-experimental-display-backpressure`;
the zero-valued configuration field is omitted by the existing JSON schema.
No source profiling or experimental rendering flag is enabled. CPU/GPU labels
below identify requested routes; the journal alone does not establish actual
GPU scene eligibility.

## Results

| Case | Native receipt-to-presentation p50 / p95 / p99 / max | Outcome |
| --- | --- | --- |
| Serial grid, CPU requested, 120 inputs | 47.605 / 48.130 / 48.343 / 48.413 ms | Valid capture; exceeds 33.3 ms latency gate |
| Serial tree, GPU requested, 120 inputs | 47.672 / 48.229 / 48.263 / 48.302 ms | Valid capture; exceeds 33.3 ms latency gate |
| Paced grid, GPU requested, 600 states | 47.820 / 62.224 / 62.435 / 62.492 ms | Valid capture; cadence passes, latency fails |
| Paced tree, CPU requested | Unexecuted | Stopped at preceding latency failure |

Paced grid has 600 distinct ordered final-state presentations and **599/599**
intervals within 18.333 ms; interval p95 is 16.666750 ms, maximum 16.666792 ms,
with no interval longer than two periods. Final CPU work overruns the planned
ten seconds by 14.668 ms; that late work remains included. All 734 acquired
prefix records present, without missing feedback, tail records or invalid states.
The two serial records retain their complete accounting, including two unarmed
startup skips in tree. Five-run qualification is incomplete; rounds 2–5 were
not launched.

The primary unresolved problem is latency despite sustained cadence. In the
paced grid, p95 UI work is 3.633 ms, acquisition call elapsed time 13.305 ms,
receipt-to-acquired 16.444 ms, and acquired-to-presented 48.505 ms. These p95s
come from separate distributions and must not be added. Acquisition and
presentation delays dominate typical samples; this is a timing observation,
not proof of a specific compositor, GPU execution or callback-arrival cause.

The existing source already requests Metal FIFO maximum frame latency 1.
Therefore simply adding that same hint is not a new repair. The next focused
investigation must establish the effective native drawable/presentation
configuration and outstanding-frame behavior before changing admission or
presentation policy. Do not enable the rejected one-credit experiment, loosen
the threshold, drop states, or rerun the same failed campaign.

## Verification and boundaries

Preflight passes 27 strict-validator controls, three adapter controls (also under
Python `-O`), focused offscreen grid/tree race tests, and both optimized native
builds. Internal review corrected the omitted-zero configuration expectation.
Prism with Anthropic Claude Sonnet 5.5 found dynamic-import registration and
optimization-sensitive assertions in the adapter; both were fixed before launch
and the focused controls rerun. The result/report checkpoint receives a separate
staged documentation review. That review found tree derived submetrics included
low-tagged warmups; the summary now uses the validator's measured-role tags only.
Independent checks confirm 120/120/600 measured samples, summary/journal agreement,
archive hashes and actual process receipts. All reported findings are resolved.

All three native and 20 configuration-probe groups returned zero and were
reaped/observed absent; source/SDK/binary/helper pins and geometry matched at
completion. This establishes process retirement and fixture cleanup, not native
object deallocation or resource/endurance acceptance. Automated routed scrolls
are not physical input-to-photon or human accessibility evidence. Resize, other
representative workflows, five-run matrix, M4 resources, M5 physical/human checks,
M6 final SDK/notices and M7 audit remain required. The internal test app is
unchanged and remains separately usable.

Reproduction: build the two existing SDK consumers with the pinned local native
archive, then use the archived `assessment.py prepare` and `assessment.py run`
in a fresh output directory with its recorded helper paths. The captured command
receipts retain every actual argument and environment whitelist. A sealed run
cannot restart; do not delete its seals. The archived adapter has no native
effects on import or preparation.

[summary.json](summary.json) contains full metrics and identities.
[evidence-index.json](evidence-index.json) binds the compressed
[evidence archive](evidence.tar.xz), including raw journals, per-case validation,
commands, logs, configuration, helper sources, plan and preflight records.
Executables, SDK payload, superseded preparation plans and historical archives
are excluded.
