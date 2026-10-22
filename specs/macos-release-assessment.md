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
Sonnet 5.5, then a commit. Preserve historical evidence without recursively copying
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


### Native submission topology control (21 October 2026)

The [three-pair native control](../docs/reports/macos-native-submit-probe/README.md)
finds combined/split p95 near 32 ms in all six runs. Separate present submission
alone does not reproduce the framework extra-refresh-period band; reject it as a
sufficient cause in this fixture. Actual render GPU clocks, all cohort states,
startup losses, owned-process retirement and unchanged SDK/configuration are
validated. No framework repair, release-gate revision or Mac qualification follows.
Keep features frozen; inspect actual framework backend submit/GPU/present-request
timing before another architecture change. Do not repeat this closed hypothesis
without a concrete new failure or source change. M3–M7 remain open.


### Actual backend timing observer (21 October 2026)

The [private-backend timing control](../docs/reports/macos-backend-timing/README.md)
completes three observer off/on pairs: all six p95 responses are35.481–36.719 ms,
above the unchanged33.3 ms gate. Actual GPU span across five render buffers has
p95 below0.3 ms; receipt-to-submit is about4.1–4.2 ms and GPU-end-to-show about
31.4–32.0 ms. Quantiles are not additive; these are private-derivative measurements,
not isolated shipping cost or proof of zero observer impact. All startup prefixes
and three prior rejected attempts remain retained; no SDK/default changes occur.
Compare the retained original binary against the private observer-off derivative
before choosing a production repair. Features remain frozen and M3–M7 stay open.


### Retained-versus-derivative binary control (21 October 2026)

The [three-pair control](../docs/reports/macos-backend-build-control/README.md)
finds retained-original p95 35.190–36.014 ms and private observer-off derivative
35.575–45.871 ms. All six miss 33.3 ms. No binary equivalence or consistent private
build improvement follows; earlier original 51–52 ms failures remain unwaived.
All states/process/source/SDK/configuration checks pass without rebuild or default
change. Inspect actual input/observer source work before selecting a repair;
features remain frozen, release gate unchanged and M3–M7 open.


### Matched-rectangle source cost and diagnostic repair (21 October 2026)

The [one-shot source-cost diagnostic](../docs/reports/macos-calibration-sourcecost/README.md)
locates about77% of source Draw CPU time in image copy/upload API work and23% in
raster for this fixture; image API p95 is2.650 ms versus raster0.742 ms. Extra
clocks prevent ordinary latency qualification. Stage0106 fixes missing grouped
diagnostic storage with original-failing/race-passing regressions; it changes no
ordinary rendering/pacing behavior or public API. Retained SDK and native archive
remain unchanged. Inspect native allocation/mapping/copy/queue work inside image
upload before choosing a performance repair; gate/freeze and M3–M7 stay unchanged.


### Native CPU upload-stage control (21 October 2026)

The [private off/on upload probe](../docs/reports/macos-upload-probe/README.md)
finds staging copy p95 2.612 ms, create-buffer0.117 ms and map0.002167 ms. Copy
may include deferred backing/cache effects; no pool/cache policy is adopted.
All248 on-mode cohort writes and six startup writes are retained/matched; source,
SDK, resource ownership and defaults remain unchanged. Next compare first versus
subsequent native copies and CPU cache modes with GPU byte readback. This remains
diagnostic-only; features/gate stay frozen and M3–M7 remain open.

The [offscreen Metal copy control result](../docs/reports/macos-copy-control/README.md)
validates all three pairs and 2,976 full GPU readbacks within 10.801 seconds.
Subsequent copies are faster; default-cache fresh-copy p95 is 0.471–0.507 ms versus
write-combined 0.642–0.684 ms. This standalone serial workload does not explain the
framework CPU-copy gap or justify asynchronous pooling/cache-policy adoption.
Next compare cache modes in the actual private framework without changing shipped
bytes, ownership or the 33.3 ms gate. Features frozen, M3–M7 open.

The [private framework upload-cache result](../docs/reports/macos-cache-control/README.md)
validates all three pairs. Default-cache CPU-copy p95 is 1.631–1.793 ms versus
write-combined 2.433–2.577 ms, but presentation remains 35.993–36.242 ms p95 in the
default-cache cases, above the unchanged 33.3 ms gate. No shipping policy adopted.
All 517 eligible buffers/case have requested cache/shared storage, including other
transient write-buffer staging; ownership unchanged. Next inspect bounded reuse
only at proven GPU-completion retirement. Features frozen, M3–M7 open.

The [completed-staging reuse policy/source audit](../docs/reports/macos-reuse-policy/README.md)
provides five passing Rust policy methods and two compiler rejection controls.
No native pool is integrated. Proposed texture-only reuse must return explicitly
from real completed-submission retirement, keep original failure/teardown paths,
bound idle storage and prove GPU bytes/ownership before timing. Gate unchanged.

The [private staging byte/lifetime pair](../docs/reports/macos-staging-reuse/README.md)
passes after the concrete primary-name fix: 44 full GPU readbacks/97,912,700 bytes,
48 CPU writes, reuse6/eviction5/idlepeak8 and closed eligible/device ledgers. All
18 groups retire within 4.588 seconds; source/SDK/config/geometry pins hold. The
original metadata refusal remains rejected. Next one reviewed bounded framework
timing pair; no shipping/SDK/app, feature/pacing or 33.3 ms gate change.

The [private staging framework timing pair](../tools/macos-calibration/stagingreuse/TIMING.md)
is prepared after the44-readback/closed-ledger proof. One off/on pair, same library/
binary and original cache policy,90seconds overall. No gate/shipping change.

The first staging timing attempt is retained rejected: request five42.563959ms late,
four warmup receipts/zero measured, enabled mode not run. No timing comparison or
unchanged retry. [Segmented request telemetry](../tools/macos-calibration/stagingreuse/SEGMENTS.md)
now separates timer wake and synchronous foreground observation without changing
20ms validity,33.3ms release gate, source/SDK/app/defaults or M3–M7. One90second
instrumented pair after review; old cause unproven.

The [segmented staging pair](../docs/reports/macos-staging-timing/README.md) is valid:
CPU copy p95 2.953→0.813ms, actual presentation 38.273→34.969ms,
all120 measured uploads reused on,18groups/closedledgers/SDKpins. Both miss33.3ms;
one ordered instrumented pair does not qualify noise/equivalence or shipping.
Original42.564ms schedule refusal remains unexplained; no unchanged retry, policy
adoption or gate change. Inspect remaining source/presentation cost; M3–M7 open.

A [texture-only cached completed-staging candidate](../docs/reports/macos-staging-cache/README.md)
is prepared privately after the original-WC pool still misses33.3ms. New explicit
HAL marker scopes cache mode to existing texture allowlist; pool/completion/limits
unchanged. Offline builds and focused tests pass; one44-readback GPU pair required
before timing. No shipping/SDK/app/default/gate change; M3–M7 open.
