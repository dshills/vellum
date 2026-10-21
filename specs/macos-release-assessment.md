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
