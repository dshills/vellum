# Bounded macOS release responsiveness assessment

## Candidate and unchanged criteria

Freeze features at main `7047740`; use the existing immutable private SDK from
`bd3cbc9ce2566955dae5dbb61385b0e1a9307eaa` as the release runtime baseline.
Record its manifest, compiler, linked native archive and optimized executable
hashes. App-only additions and CI portability repairs after that checkpoint do
not silently change the SDK under measurement. Final release packaging remains
a separate checkpoint.

Use the default rendering/scheduling path: omit experimental display-backpressure,
source profiling, novel-scroll and other experimental flags. Presentation metrics
are observation only. The source105 61–63 ms comparison explicitly enabled
one-second experimental backpressure and remains a failed experimental result,
not a default-path measurement.

Keep the [existing release thresholds](macos-responsiveness.md): 99% of ordered
changing-state intervals within18.333ms, input-to-actual-presentation p95≤33.3ms,
and resize stalls≤50ms. The4ms CPU target is optimization only.

## Bounded first checkpoint

Reuse the existing stored grid/tree consumers, strict check-presentation validator,
owned process executor and exclusive native lease. Reuse existing single-attempt
case orchestration rather than rebuilding historical campaign machinery.

1. Headless preflight: validate journals and focused consumer configuration/
   capture/pacing regressions, then optimized native builds and metadata.
2. During one explicitly authorized quiet desktop window (at most20minutes),
   run serial grid CPU-requested120inputs, serial tree GPU-requested120inputs,
   paced grid GPU-requested600states and paced tree CPU-requested600states.
   Each uses10,000stored records and128warmups, with unchanged viewport/workload
   assertions. Record actual display/refresh/scale/power and geometry isolation.
3. Every native process has a330second watchdog. Stop the campaign on the first
   invalid journal, source drift, user interruption, resource/lifecycle failure,
   or failed paced responsiveness predicate. Retain all failures; no same-case
   blind retries, smaller workloads, discarded timestamps or changed cutoffs.
4. Only if this diagnostic round passes, use the remaining admitted window for
   four more rounds (five per case total). Each round has a fresh output path.
   Never infer full release qualification from this grid/tree subset.

Missing native clock feedback or an occluded/lost-prefix journal is a concrete
assessment blocker, not a numeric latency result. Inspect the exact failure
boundary and choose one focused repair before running it again. An unmodified
failed campaign is sealed. Stop an unsuccessful investigation after90minutes
and explicitly reassess the implementation boundary.

## Following checkpoints

After this first result, choose only evidence-supported repairs or extend passing
coverage to composed workspace, text/property editing, canvas and resize. Finish
resource/hidden-Cocoa-window cleanup, physical/human acceptance, final SDK/notice
closure and release audit against the existing M3–M7 requirements. No new UI
features or renderer/pacing experiment is authorized by this assessment.

Each checkpoint retains a concise requirement/source/command/result/next-action
entry, focused tests, internal review and Prism review using Anthropic Claude
Sonnet5.5, then a commit. Preserve historical evidence without recursively copying
it. No completion claim follows while a required release gate is open.

## First result

The [21 October default-path report](../docs/reports/macos-default-responsiveness/README.md)
records valid feedback, passing single-run paced-grid cadence and failed latency.
The campaign stopped at that failure; no repeated or fourth case was run. The
next focused boundary is effective native configuration/outstanding frames,
before any presentation policy change. M3 and the Mac release remain open.

## Focused native investigation result

The [native investigation](../docs/reports/macos-native-latency-investigation/README.md)
confirms the two-drawable synchronized native configuration. Transaction mode,
explicit transaction flush, temporary display-link acquisition and a persistent
acquisition broker are rejected: invalid prefix feedback or latency above the
unchanged gate. No shipping framework change is adopted. The minimal Metal clear
reference also has startup feedback loss and is not qualification or an OS floor.

Do not tune those rejected paths again. The next implementation boundary is a
window-owned native frame request before reconciliation, with explicit provided-
drawable ownership through the native surface rather than interception of
`nextDrawable`. First bind a matched ready-window reference and record actual
callback/target-presentation timestamps. Preserve receipt-before-queue semantics,
all states, FIFO/vsync, cancellation and shutdown. Missing target timestamps and
failed prefixes currently prevent choosing a safe repair; this does not waive M3.

## Ready native reference result

The [ready-window reference](../docs/reports/macos-native-ready-reference/README.md)
closes the readiness/target-timestamp diagnostic prerequisite with two valid
128-warmup/120-measured cohorts. Raw render-request p95 is 48.843 ms; direct native
display-link callback p95 is 49.908 ms and its presentation target is about 50 ms
ahead, followed within 0.018 ms. Boundaries differ; no causal A/B pooling or OS-floor
claim follows. Actual foreground/key-window state was not instrumented.

No framework repair or threshold change is adopted. Before another backend
change or a baseline-derived release-policy proposal, verify foreground/key-window
state and scheduling configuration. Any policy revision needs owner approval,
predeclared criteria and fresh representative measurements; the minimal reference
does not qualify the framework. Existing failed 33.3 ms scopes remain failed.

## Foreground-controlled follow-up

The [foreground control](../docs/reports/macos-native-foreground-control/README.md)
records valid actual active/key/visible/foreground-PID boundaries and no focus-loss
notifications in both native cohorts. Raw p95 is 65.265 ms; direct provided-drawable
callback p95 is 49.868 ms with approximately 50 ms target lead. These different
receipt boundaries must not be pooled or treated as a causal A/B pair. Both paths
remain above 33.3 ms; this is not an OS-floor or framework-qualification claim.

The bounded investigation found no safe accepted repair. Stop rejected-variant
retries and obtain owner direction on retaining the absolute budget versus a new
predeclared native/framework calibration. No threshold is changed here; any revised
policy requires approval and fresh representative measurements. M3–M7 remain open.

## Matched calibration result

The [predeclared matched calibration](macos-calibration.md) and its
[source-bound result](../docs/reports/macos-matched-calibration/README.md) complete
five pairs: native p95 31.951–32.416 ms, Vellum 37.420–52.250 ms, with all states
accounted for. Keep the release gate unchanged; the earlier native-reference
numbers are not a universal floor. This sparse rectangle fixture is not 60 Hz or
professional-workflow qualification, nor does its difference isolate observer or
production-framework overhead. The next bounded boundary is actual submit/
present scheduling after acquisition; causal extra-frame/GPU timing is unproven.
No earlier failure is relabeled; M3–M7 stay open and features remain frozen.
