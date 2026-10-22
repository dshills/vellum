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

The [texture-only cached staging GPU pair](../docs/reports/macos-staging-cache/README.md)
passes44readbacks/97,912,700bytes, exactly22/16actualdefault-cache
freshtextureallocations andclosedeligible/deviceledgers,18groups/4.654s.
No wrongcache/unconfigured/malformed counters; originalqueued/oversize/R8/unsent
proof retained. Next prepare onecachecomparison withreusefixedon; nogate, app,
SDK/default/shipping change or UI-latency acceptance. M3–M7 open.

The [texture-only cache timing pair](../docs/reports/macos-staging-cache-timing/README.md)
is prepared after full byte proofs, reuse fixed enabled in both modes. Sameprivate
library/binary,90seconds, all original validity/presentation/ownership checks. No
result or shipping/default/SDK/app/gate change claimed; M3–M7 open.

The [texture-cache/reuse timing pair](../docs/reports/macos-staging-cache-timing/README.md)
is valid: copy p95 0.804/0.770ms, presentation 36.089/34.931ms
(WC/default), bothmiss33.3. Both120measuredreusehits/3fresh/closedledgers, exact
cachecounts/18groups/SDKpins. No further unchanged cache campaign or adoption;
inspect remaining raster cost. Features/defaults/gate unchanged; M3–M7 open.

The [private uniform-fill control](../docs/reports/macos-uniform-fill/README.md)
is rejected before native testing: pixel oracle passes but bounded CPU benchmark
overlaps; pinnedGo already copies rows. No adoption/unchangedretry; inspectactual
rastercommands next. Features/defaults/SDK/app/gate unchanged; M3–M7 open.

The [offscreen raster command observer](../docs/reports/macos-raster-commands/README.md)
retains4renderer calls/80records while original full pixel/state checks pass.
Recordedcommands are notexecutedfillcounts ornativecostattribution; geometrydiffers.
Caps/reset/callback/snapshot tests pass. Next reviewedboundednative3transition
structuralcapture, no new repair/gate/default/SDK/app change; M3–M7 open.

A [three-transition native raster structural capture](../docs/reports/macos-raster-native/README.md)
is prepared with originalMetal/defaults, bounded actual root branch metadata and
strict existing routing/presentation/ownership checks. Focused tests/build pass;
no native result, repair, latency qualification or gate change yet. M3–M7 open.

The [native raster structural result](../docs/reports/macos-raster-native/README.md)
is valid: one attempt/4.782 seconds, 11 calls/162 records, three admitted roots.
Each root takes four covered branches, three uniform crop branches, four no-paint
identity RenderPath calls and one Params.Render call. Existing optimization is
active; eight recorded images are not eight executed fills. Nine groups retired,
SDK/source/native/configuration/geometry pins stable. Three transitions are not
a latency distribution, per-tag source attribution or displayed-byte proof.
Inspect no-paint path execution/raster boundaries offscreen next; no repair,
unchanged retry, default/SDK/app/gate change or release acceptance. M3–M7 open.

The [focused offscreen raster cost checkpoint](../docs/reports/macos-raster-offscreen-cost/README.md)
uses unchanged Core: full warm replay167.693–172.643µs, four no-paint paths
129.5–130.6ns, last fill163.943–165.436µs, zero allocations. ExplicitCPU2,
three observations/case, two focused Go methods and nine parser refusals pass.
Native scalar geometry matches; overwritten colors/full path bytes and temporal
conditions differ, so no native equivalence, phase sum or latency claim follows.
No path-elision candidate or new native run is justified. Investigate separated
replay/cache conditions offscreen next; features/defaults/SDK/app and33.3ms gate
unchanged. M3–M7 open; no production repair adopted.

The [separated offscreen replay control](../docs/reports/macos-raster-separated-cost/README.md)
retains32samples/eightpercondition in1.998s with unchanged renderer. Immediate,
120ms pause and32MiB scratch-footprint conditions overlap; observed replay ranges
0.178–0.902ms. Final legacy pixels/focused method/nine parser refusals and source/
SDK/log/group checks pass. Hot microbench timing is not native timing; no proven
cache flush, isolated CPU/scheduling/GC cause, native-p95 or repair/adoption claim.
Keep future timing joined to surrounding upload/schedule boundaries. No new
native test; features/defaults/SDK/app/33.3ms gate unchanged, M3–M7 open.

The [joined raster/upload/scheduling diagnostic](../docs/reports/macos-joined-timing/README.md)
is prepared: private bounded absolute raster brackets plus native wake/check
stamps, exact same-input source/upload/actual-presentation joins and contiguous
interval closure. Go timer offsets/nested work remain separate; no guessed clock
conversion or phase-quantile sum. Five recorder race methods, original routed
pixels, five classifier methods/23refusals, build and exact recipe/pins/SDK checks
pass. One reviewed/committed40s native case60overall is approved when ready; no
result or production repair yet. Features/defaults/SDK/app/33.3ms gate unchanged,
M3–M7 open.

The [joined native result](../docs/reports/macos-joined-timing/README.md) passes248
same-input joins/9groups in37.695s. Instrumented receipt→shown p95=35.353ms;
17/120measured observations exceed33.3. On that actual rank ticket, raster0.332ms,
native upload1.460ms, other before acquisition0.320ms and acquisition→shown33.241ms
partition the same response. Post-acquisition still includes encode/submit/GPU/
display phases; no compositor-only attribution, phase-quantile sum or fix follows.
Full startup/pins/SDK/configuration/geometry/closure retained. Next join backend
boundaries to the same identifiers. No native retry, production/default/SDK/app/
feature/gate change; M3–M7 open, no ordinary release qualification.

The [joined CPU/backend preparation](../docs/reports/macos-joined-backend/README.md)
continues the original Cogent/WebGPU design. Exact reviewed upload and backend
observers compose privately with the unchanged raster clock. Six focused methods/
19 negative mutations, original routed full pixels, fresh native/Go builds and
full source/log/SDK/reproduction checks pass; prior recorder race checks are
reused by exact source comparison. One newly approved 40s native/60s overall
capture may follow review and commit. GPU span includes gaps; final-render-GPU-end
to shown is not exclusively compositor cost. No native result, repair, feature,
default, SDK/app or 33.3ms gate change. M3–M7 remain open.

The [same-input backend result](../docs/reports/macos-joined-backend/README.md)
passes 248 joins in 37.509 s, with nine owned command groups retired. Instrumented
p95 is 36.329 ms; 34 of 120 measured observations exceed 33.3 ms. On that actual
ranked input, the render GPU span is 0.289 ms and final render GPU end to shown
is 30.933 ms. Presentation was requested 1.138 ms before final render GPU end.
Separate present-command GPU end to shown is 30.929 ms. This localizes the
remainder beyond the recorded commands without isolating compositor cost,
proving a fixed frame penalty or attributing the cause to WebGPU. All startup,
command, source, SDK, configuration and log accounting is retained. The original
design, SDK, app, defaults, features and 33.3 ms gate remain unchanged. M3–M7 are
open; no repair or ordinary-build qualification follows. The approved live scope
is consumed; this result does not authorize another native attempt.

The [display scheduling preparation](../docs/reports/macos-display-scheduling/README.md)
adds a private view-linked display-clock observer without GPU work, frame requests
or layer/presentation policy changes. Eight focused methods and 20 negative cases,
headless native ownership checks, full routed pixels, private consumer build and
source/SDK/log/reproduction checks pass. Exact reviewed native/Core bytes are
reused. Reported timestamps/targets are not optical scanout, guaranteed latch
opportunities or missed-vsync proof. One approved observer-off/on pair may follow
review and commit, capped at 100 seconds overall and 40 seconds per native case.
No native result yet. Features, SDK/app, defaults and 33.3 ms gate remain unchanged;
M3–M7 stay open.

The [display-clock off/on result](../docs/reports/macos-display-scheduling/README.md)
passes both cases in 78.604 seconds, with 18 owned groups retired and 248
joined inputs per case. Same-binary p95 is 51.101 ms off versus
33.947 ms on; both miss 33.3 ms. The difference of run p95 values is
-17.154 ms, about one refresh period, without a repeated or causal claim.
All 1,789 callbacks and full startup/held tails remain retained; actual shown
feedback is within 10.542 microseconds of the nearest reported timestamp.
These are reported clocks, not optical scanout or guaranteed latch opportunities.
Observer neutrality is unsupported; reverse order before proposing a scheduling
change. No permanent link, backend/default/SDK/app/feature or release-gate change.
No release qualification; M3–M7 open. Approved live scope consumed.

The [reversed-order preparation](../docs/reports/macos-display-reversed/README.md)
reuses the exact final display-clock binary and native archive for a fresh on/off
pair. Five orchestration methods, seven failure cases, syntax/fresh-scope guards
and full source/SDK identity checks pass without rebuilding unchanged native code.
The old off/on scope stays sealed. One newly approved quiet scope may follow
review and commit, capped at 100 seconds overall and 40 seconds per native case.
No result yet. Features, SDK/app, defaults and 33.3 ms gate unchanged; M3–M7 open.

The [reversed-order observer result](../docs/reports/macos-display-reversed/README.md)
passes the fresh on/off pair in 74.272 seconds with
18 owned groups retired and 248 joined inputs per case.
Receipt-to-shown p95 is 37.882 ms on and
37.923 ms off; on-minus-off run p95 is
-0.040 ms. The previous off/on pair remains
separate (33.947 ms on / 51.101 ms off).
No pooling, causal claim, observer overhead calibration or ordinary qualification.
Complete callbacks/startup/held tails and source/native/SDK identity are retained.
No permanent link or production/default/feature/SDK/app/gate change. M3–M7 open;
approved scope consumed, desktop free, no native job active.

The [offscreen journal phase analysis](../docs/reports/macos-journal-phase/README.md)
reclassifies all four immutable journals through the original pinned validators:
992 joins, 512 warmup and 480 measured, without native execution or rebuilding.
Seven focused methods and byte-identical independent reproduction pass. Actual
p95 inputs retain 81–92% of latency after final-present GPU end. Reported timestamp
and callback observation are independent event series: the preceding reported
callback is still pending at 118/120 reversed-on receipts. Earlier on versus
reversed on nearest-shown-timestamp callback observation lag has p95 10.912 versus
16.340 ms. These are descriptive clock relations, not optical/latch/causal proof.
No observer scheduling fix, ordinary qualification, feature/default/SDK/app/gate
change follows; M3–M7 open. Future controls should match request phase and record
pending presentation/drawable state before changing scheduling policy. No new
native scope is prepared or authorized by this offscreen checkpoint.

The [phase-matched pending-state preparation](../docs/reports/macos-presentation-phase/README.md)
adds private native/framework consumers with four predeclared request phases,
30measured inputs per phase and complete scalar acquisition/receipt/callback
censuses. Actual native receipt matching is required before comparing latency;
recorded pending feedback does not establish Metal pool availability or ownership.
Eight focused methods, headless native concurrency/lifecycle/bounds/JSON, both
builds, full routed pixels, recipe/source/log/SDK checks pass. No native capture
yet. One approved100second overall/40second per-process scope may follow review
and commit. Existing design/features/defaults/SDK/app and33.3ms gate unchanged;
M3–M7 open. Actual results or rejection receive separate review and commit.

The [first phase-matched control](../docs/reports/macos-presentation-phase/README.md)
is rejected in52.939seconds. Native completes248joins with30measured
per phase, p9534.960ms and zero other recorded pending drawables at measured
receipt/present/GPU points. Framework fails startup before any input or metrics
sample; no matched comparison exists. Linked-code audit finds pending hooks absent
from its acquisition method despite the Objective-C overlay. The original acquire
observer call is linked, so the whole startup failure remains unisolated. All18
groups/logs/config/geometry/source/SDK8321checks pass; rejection retained/sealed.
Next diagnostic preparation must materialize native changes in a private Core
copy and verify linked hooks. No unchanged retry or features/defaults/SDK/app/
gate change; M3–M7 open. Original quiet scope consumed, desktop free.

The [physical native-source repair preparation](../docs/reports/macos-presentation-physical/README.md)
materializes the known three Core replacements in a fresh6031file private copy.
Actual linked acquisition calls now include pending hooks plus the original
observer; a headless factory test verifies subclass and stamp publication/revocation.
Original Core/SDK8321files remain unchanged. Native reference reused byte-identically,
framework consumer rebuilt; focused integration/log/source/recipe proofs pass,
unchanged classifier/pixel/bounds evidence reused. First failure stays sealed and
is not reclassified. No new live capture; fresh quiet approval pending. Phase
protocol, features/defaults/SDK/app and33.3ms gate unchanged; M3–M7 open. Review and
commit precede execution; actual results/rejection receive separate checkpoint.


The [offscreen Objective-C overlay control](../docs/reports/macos-presentation-overlay/README.md)
reproduces the missing instrumentation with the diagnostic's exact Go 1.25.6:
Go/C replacements apply while `.m` compiles original bytes; a physical copy
applies all replacements. Three scalar fixture builds/runs plus the version
check pass, with seven bounded owned groups retired and prior build/capture pins
unchanged. This confirms the materialization rationale, not the separate startup
feedback failure. The user explicitly deferred the next live attempt; corrected
capture remains unstarted. No UI launch, feature/default/SDK/app/gate change or
release qualification; M3–M7 remain open.


The [corrected live phase-control attempt](../docs/reports/macos-presentation-physical-result/README.md)
stops after40.713seconds: native returns normally, but measured request215
arrives2.143599ms after its target, violating the predeclared2ms diagnostic phase
bound. All248 raw clock/tag/drawable/callback joins, parent/clock/replay checks,
9ownedgroups/log/configuration/geometry/source/SDK8321 checks pass; no filtering,
retry or cutoff change. Fail-first prevents framework launch, so corrected startup
and a matched latency comparison remain unverified. The attempt is sealed and
desktop free. A separate startup-only check is the next direct unresolved test;
none is authorized or executed here. Features/defaults/SDK/app and33.3ms release
gate unchanged; M3–M7 open. Actual rejection receives review and commit.


The [startup-only preparation](../docs/reports/macos-framework-startup/README.md)
reuses the corrected framework's existing pilot mode, original setup validator,
pending replay and owned Host, independently of native phase matching. One20s
process/45s total, no rebuild or routed cohort. Eight focused methods/23negative
subcases plus actual unauthorized-CLI refusal pass; source/compiled/SDK pins verify.
No UI has run and no release gate closes. Fresh up-to1minute quiet approval is
needed; features/defaults/SDK/app and33.3ms/60Hz gate unchanged, M3–M7 open.


The [isolated corrected-framework startup](../docs/reports/macos-framework-startup-result/README.md)
passes in8.111seconds (native1.462seconds):4positive scoped samples exactly join
attempts7–10, scope closes, pending/acquiring0. All10attempts/30events are retained,
including2zero callbacks among6pre-observer scope0attempts.9ownedgroups/18logs,
source/binary/geometry/configuration andSDK8321verify. This closes only the
corrected-binary startup prerequisite. Observer lifecycle passes with0ticks;
no phase pair, cadence, latency or release qualification follows. Earlier scopes
stay rejected/sealed; no unchanged retry or feature/gate/default/SDK/app change.
Next measure corrected-framework input/presentation/pending state under a separate
reviewed scope and fresh quiet approval. M3–M7 remain open.


The [corrected-framework cohort preparation](../docs/reports/macos-framework-pending/README.md)
reuses startup-qualified binary6825d90e for its first timed cohort independently
of the failed native-reference phase pair. Originalstrict128warm/120measured,
phase/backend/clock/receipt/drawable/callback constraints remain unchanged, with
40s process/60s overall caps andfail-firstsealing. Twofocusedrouting methods/
7negativecases andactualunauthorizedCLIrefusalpass, allcompiled/SDK8321pinsverify.
No UI or rebuild yet; ownerauthorizedscreenforremainingreleasetests. Review/commit
preparationthenactualresult. No nativecomparison, representativequalification or
feature/default/SDK/app/33.3ms/60Hz gatechange; M3–M7 open.


The [first corrected framework timed cohort](../docs/reports/macos-framework-pending-result/README.md)
is rejected during warmup: 32 completed inputs, request 33 foreground/configuration
loss, invalid display observer, zero measured inputs. All 42 attempts/158 events
and 259 ticks are retained; independent replay and 36 scoped sample joins pass.
All nine groups retire, pending/acquiring zero, recorder closed and all pins,
SDK 8321, geometry/configuration unchanged. Total 12.111 s, native 5.218 s.
No latency/cadence/native comparison or release qualification follows. Another PID
was foreground; its identity/cause is unknown. Nil key-window layer does not prove
window destruction. Record the first failed per-request state before a fresh
attempt; do not retry this sealed scope. Features and release gate remain unchanged.
