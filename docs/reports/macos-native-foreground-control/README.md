# Foreground-controlled native reference — 21 October 2026

**Both native reference cohorts are valid with foreground boundary controls.
Neither meets the existing 33.3 ms latency threshold. No framework fix, release
qualification or threshold change is adopted.** This addresses the foreground
instrumentation gap in the [ready reference](../macos-native-ready-reference/README.md).

## Bounded control

The private Metal reference completes application launch, requests activation,
and waits at most one second for actual active/key/visible state and its own
foreground PID before rendering. It records those states before frame production
and after callback drain. Any application resignation, key-window resignation or
foreign-application activation notification invalidates capture. Both cases have
valid start/end states and zero such notifications. This is bounded state and
notification evidence, not continuous physical-input verification.

Rendering, two synchronized drawables, transaction presentation disabled,
next-drawable timeout enabled and the previous strict cohort classifier are
preserved. All setup records are retained. Positive setup presentation precedes
128 warmups and 120 measured frames per path; every cohort presentation must be
positive and strictly ordered. No failed earlier framework capture is repaired.

One attempt per path ran in the newly approved two-minute quiet window,
13:34:03–13:34:15 America/Chicago. The 12-second total includes probes and
activation; it used part of the approved two-minute allowance. The M4 Pro, macOS 26.6.2/25G83,
LG 5120×1440/60 Hz, scale 1, AC configuration is unchanged before/after each case,
as is the framework geometry store. The reference uses a 1519×1266 drawable.

## Results

| Receipt boundary | Measured p50 / p95 / p99 / max | Interval p95 / max |
| --- | --- | --- |
| Raw Metal request before acquisition | 64.727 / 65.265 / 65.481 / 65.557 ms | 16.666708 / 16.666709 ms |
| Direct display-link supplied-drawable callback after acquisition | 49.834 / 49.868 / 49.886 / 49.913 ms | 16.666708 / 16.666750 ms |

Each case has 119/119 measured intervals within 18.333 ms. Raw retains two setup
records with one zero presentation time; direct link retains four with two zeroes.
All 128 warmups and 120 measured frames per case have valid feedback. Startup
losses remain explicit; neither single run establishes five-run qualification.

Direct link still requests `preferredFrameLatency=1`. Callback-to-presentation
target p95 is 49.872 ms; actual presentation differs from its target by at most
0.008250 ms in absolute value. The slightly negative signed differences (p95
−0.004583 ms) are retained. No target or actual timestamp is substituted for the
other. CACurrentMediaTime and MTLDrawable presentation times use host time; the
local Apple SDK declarations and their hashes are recorded in the archive.

The two receipt boundaries differ. These runs are not a causal A/B comparison,
and the raw-path difference from the earlier uninstrumented reference must not
be attributed to foreground activation alone. This empty changing-clear program
also differs from a real framework workload and does not prove a universal OS
latency floor or exonerate Vellum's queueing cost. It does show that neither of
these ready, foreground-controlled native configurations meets the current gate.

## Review, integrity and decision

The warning-free native build and 15 focused controls pass (nine inherited
classifier controls and six foreground controls). Prism with Anthropic Claude
Sonnet 5.5 reviewed the diagnostic. Its callback-wait and foreground-validation
findings were checked against synchronous group entry before submission, producer
freeze before handoff, main-thread end snapshots and strict zero-loss validation.
The clock basis is documented; dispositions retain the evidence and no findings
remain unresolved. The final actual captures pass these controls.

Both native process groups and all 16 configuration-probe groups exit cleanly and
are observed absent. All 8,321 SDK members, runtime pins, raw log hashes, geometry
and configuration checks pass. This does not establish native allocation balance,
endurance, physical pointer response or a production fix.

Stop repeating the rejected presenter variants. The bounded investigation has
not found a safe repair that meets 33.3 ms. Before further backend changes, the
owner needs to decide whether that absolute budget remains the requirement or
whether to calibrate a new budget through a matched native/framework assessment.
A new budget requires an approved, predeclared protocol and fresh representative
framework runs; these results cannot retroactively pass existing failures. Until
then the gate and M3–M7 stay open, and framework features remain frozen.

[summary.json](summary.json) records exact metrics and artifact-relative pins.
[evidence-index.json](evidence-index.json) binds [evidence.tar.xz](evidence.tar.xz),
including source, controls, actual records, ownership/configuration receipts and
reviews. Immutable execution receipts retain original local paths. The executable,
full SDK and external reused helpers are not bundled; their availability and hashes
are explicit. This is not a standalone rebuild or public-release package.
The final report/specification Prism review has complete coverage and no
actionable findings; its timing clarification is recorded in the dispositions.
