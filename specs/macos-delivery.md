# Mac framework delivery plan

**Current Mac release policy (19 October 2026):** the owner selected measured
60 Hz responsiveness as the release gate and retained 4 ms as an optimization
target. The [release policy](macos-responsiveness.md) supersedes treating the
provisional work/footprint budgets below as mandatory Mac release thresholds.
Historical results retain their original targets and pass/fail labels; they do
not establish presentation or close the new gate. Functional correctness,
lifecycle, bounded-resource safety and final delivery checks remain required.


The [paced presentation checkpoint](../docs/reports/macos-paced-presentation/README.md)
implements an opt-in fixed-deadline 60 Hz grid/tree workload and strict version 2
journal analysis, preserving late/unstarted work and collapsed-display anomalies.
This is preparation and headless evidence. Actual callback/clock/lifecycle proof,
five-run native performance and the other representative M3 workflows remain
open; no Mac milestone is closed or private SDK refreshed.


The [serial capture checkpoint](../docs/reports/macos-presentation-capture/README.md)
adds opt-in stored grid/tree journals and a strict offline validator. Race tests,
vet, native builds and synthetic journal controls pass. No real native feedback
was captured here; serial scroll intervals do not qualify sustained 60 Hz pacing.
Callback/clock/lifecycle proof and a continuously changing campaign remain open.


The [native presentation checkpoint](../docs/reports/macos-native-presentation/README.md)
implements optional bounded display recording and passes headless accounting,
actual Darwin bridge compilation and controlled drawable callbacks. It has not
run with a real native drawable. Stage88 changes the owned Metal layer, so older
native records retain their original source/configuration scope; affected render
and lifecycle behavior needs renewal before current-candidate qualification.
The desktop remains deferred while the owner is online.


Updated 19 October 2026. The latest complete table matrix measures `a2b4143`
Go code with stage0080 overlays over the retained source-built native library.
The unchanged private SDK uses clean `abe67c6`, Go 1.25.6 and stage0079. Four serial starter/drawing/3D/imaging native source-assertion/readback
cases now pass from that exact SDK on this Mac. Two composed lifecycle cycles
and four eligible 60-second idle CPU intervals also pass from the same SDK;
full M4 acceptance remains open. Extra queue admissions in two traces and one
unclassified hidden Cocoa window remain open anomalies; tracked retirement and
process-group absence do not establish loss-free delivery or native deallocation.
Owner: the main Vellum development chat. This is the execution plan for the
existing [Mac completion requirements](macos-completion.md), not a replacement
for those requirements or a declaration that implemented APIs have qualified.


The [native workspace checkpoint](../docs/reports/macos-native-workspace/README.md)
now passes all 216 operations and 27 captures at both 10k/1M rows through the public
optional DesktopSession and dedicated native consumer. Actual focus, independently
predicted viewport scroll, semantics, painted placement/history and surviving-
window state pass; both owned processes and tracked sessions/workers/services
retire. Failed earlier cases remain failed. This closes the native full-record
subgate, while catalogue coverage and full M2 remain open. Controlled GOMAXPROCS=2,
the retained SDK's separately pinned native library, autonomous scheduling and
one unclassified hidden TUINS Cocoa window remain explicit limits. Physical/
human input, presentation/performance, M3–M7 and full Mac completion remain open.


The [current workflow checkpoint](../docs/reports/macos-current-workflows/README.md)
renews the complete native workspace432/54 on9013a73/stage87 and passes the full
text/editor/modal68/33. Current native float/redock/close-return and document
save/dirty-close source assertions are accepted, with independent persisted
session and retained document output checks. The original document controller
remains failed due a checker node-count error; a separate source-derived1002total
expectation is reviewed against the same output without a native rerun.
Six original frozen software records and portable root race/defaultvet pass.
M2 remains partial for current native catalogue shared-runtime qualification and
dedicated-contract audit. Resolve those two named gaps next, then M3's materially
different measured frame-cost fix and unchanged full matrix; M4–M7 remain open.
The SDK is unchanged; this is controlled Go2/retained-native-archive evidence.


The [popup lifecycle prerequisite](../docs/reports/macos-popup-lifecycle/README.md)
adds stage82: timed native popup expiry and terminal event-loop stage destruction
share the render owner, stale callbacks lose authority and successful expiry
requests a repaint. Owned software race checks and reproducible preparation are
separate from native/physical acceptance. The broad upstream diagnostic failure
and restored SDK test-image side effect remain explicit. No new catalogue or
full Mac gate is closed; ordinary Selector/NumberField checks follow in the checkpoint below.

The historical [popup and Selector integration checkpoint](../docs/reports/macos-native-controls/README.md)
adds stage83 bounded actual-menu inspection, repairs owner wakes starved by modal
menus, and restores actual Selector click focus before its menu opens. Final
Core/driver/gallery race, vet, build and preparation checks pass. Four native
attempts remain failed and sealed: the final source-asserted prefix reaches ten
Selector observations plus one accepted numeric edit (24 inputs/11 observations),
then refuses the next key because Enter cleared numeric focus. No complete native
catalogue, M2/M5 or other Mac gate closes. The corrected checker authors four routed refocus clicks after numeric Enter
(57 inputs/21 unchanged observations); that checkpoint left revised native qualification pending. The subsequent
result below completes this record; remaining catalogue checks continue;
physical/OS input, human accessibility and M3–M7 remain open. No SDK refresh or
OS settings change occurs.

The subsequent [retained-state and multi-click checkpoint](../docs/reports/macos-control-refocus/README.md) completes the unchanged native
Selector/NumberField record: 57 routed inputs and 21 independently expected
observations. Go adapter fixes restore inherited disabled styling and actual
focus before ordinary double/triple-click handling, preserving selection and
exactly-once edits. Two preceding 57-input cases remain failed at 54 receipts/19
source-asserted observations; the styling-only case did not fix actual focus.
Controlled Go2/SDK786a CPU-scene routing is qualified for this record only. Full
catalogue/M2, physical/human M5 and M3–M7 remain open; no SDK refresh occurs.

The [choice catalogue checkpoint](../docs/reports/macos-choice-catalogue/README.md)
adds a complete 30-input/23-observation native record for Checkbox, Switch,
ToggleButton, RadioGroup, SegmentedControl and Slider keyboard behavior. Shared
state, rejected changes, disabled skipping/refusal, focus and exactly-once
callbacks pass. ProgressBar/Ring/Gauge and status publish expected values and
fully visible geometry; native progress pixels and Slider drag are not claimed.
Original-reference/session/process retirement passes. This advances named
catalogue coverage in controlled Go2/SDK786a CPU-scene routing; remaining catalogue,
full M2/M5, physical/human input and M3–M7 remain open. No SDK refresh occurs.

The [list/menu catalogue checkpoint](../docs/reports/macos-navigation-catalogue/README.md)
adds stage84 support for direct Button embeddings in popup inspection and
47 native Go FIFO inputs/41 independent observations. The original 17-input
case remains failed/sealed; only the distinct corrected scope passes. ListBox scrolling,
explicit repeat proposals, rejection/read-only/disabled behavior and Tab pass;
MenuButton duplicate-label actions, actual popup pointer/keys, disabled/stale
refusal, Escape/outside consumption, modal owner disable, check/separator/reset
and original-window closure pass. Callback totals are 8 list, 2 menu, 3 check/reset
and 1 outside action. This is controlled Go2/SDK786a CPU-scene routing. Remaining
catalogue, full M2/M5, physical/human input and M3–M7 stay open; no SDK refresh.

The [current catalogue and contract audit](../docs/reports/macos-current-catalogue/README.md)
qualifies all five current native records: 255 routed inputs and 169 observations.
The assertion audit covers 36 catalogue rows, 29 requirement groups and 205 exact
executed test functions. M2 remains open for three explicit dedicated Cocoa
methods: the eight-stage tab-close, twelve-stage Breadcrumb and eight-step
ColorPicker selector records. Their fixtures pass current builds/software checks
and internal source review, but have not run natively on this candidate. The
owner is online, so foreground testing is deferred. Reconcile the current source
baseline before preparing their exact launch plan. No old scope is replayed.
After M2, resume the materially different measured M3 rendering/publication work;
M4–M7, physical/human qualification and final SDK acceptance remain open.


The [owned linker checkpoint](../docs/reports/macos-owned-linkers/README.md)
replaces five undeclared upstream linker-module distributions in the private Mac
SDK profile with explicitly licensed, freshly authored compatibility source.
137 affected tests, independent relocation/race/vet/build and all 26 default mains
plus one tagged helper pass. Original resolution metadata stays separate from
the effective owned source; the exact native archive and prepared dependencies
are unchanged. This closes that effective-source declaration gap for this SDK
profile, while direct source/historical SDK graphs and their original findings
remain unchanged. Native/font/asset applicability, distribution approval and full
M6/Mac acceptance remain open. No application is launched in this checkpoint.

The [extracted SDK owned-build check](../docs/reports/macos-sdk-build/README.md)
now passes from a Git-free private candidate: normal Python app generation,
actual optimized ARM64 compilation with a linker map, and notice-bearing unsigned
app packaging. Manifest-backed source integrity replaces the Git-only collector
assumption for extracted SDKs; declared provenance remains unauthenticated.
The candidate is a disclosed precommit snapshot, and the app was not launched.
This closes the SDK build-tool blocker, while final native app acceptance,
component/asset applicability, owner distribution approval and full M6 stay open.

The [test-font notice checkpoint](../docs/reports/macos-test-font-notices/README.md)
adds discoverable full declarations for the four typesetting test fonts through
hash-pinned preparation. All font/runtime bytes remain unchanged. Fresh recipe
and collector checks pass; Selawik modification/reserved-name provenance, broader
asset applicability, distribution approval and full M6 remain open.

The [cursor artwork contract](cursor-artwork.md) replaces all 39 built-in SVGs
with original MIT-licensed Vellum geometric artwork through bounded, pinned
preparation overlays. Public identities and existing rendering/custom fallback
semantics remain. This changes executable inputs and requires fresh SDK/build
checks; it does not establish physical cursor, native lifetime, other asset
applicability or M6 acceptance.

## Finish line

Finish the Go-native framework on the local Mac: an independently consumable SDK
and a professional multiwindow application must exercise the supported controls,
document/history, docking, large data, custom GPU views, native input/accessibility
and platform services with correct ownership, bounded resources and the specified
performance. The acceptance and distribution gates must close, with supported
Mac configurations and any unavailable capabilities explicitly documented.

Windows/Linux qualification and a CAD application are outside this milestone.
The Go API, Cogent-based native driver and shared-device rendering foundation are
settled. New controls are added only to close a specific documented application
gap; the existing control catalogue is the milestone baseline. This preserves
the requested breadth without making completion depend on an unlimited widget list.

Full Mac completion requires every milestone below under the current release policy. An alpha label, compilation,
native readback or a smaller successful fixture cannot replace an open gate.
Public upload, tagging, signing/notarization with an owner's identity, and changes
to system input/accessibility preferences are not implicitly authorized here.

The [Go standard-library input checkpoint](../docs/reports/macos-stdlib-inputs/README.md)
closes a specific M6 collector gap: selected runtime/standard-library source,
embed/object inputs and named ancestor declarations now participate in sampled
collection and the owned build's before/after checks. The corrected minimum-Go
private build records 236 packages, 1,348 selected files and nine declaration
candidates; 96 related test methods pass. No produced app is executed and no
rights or complete Mac gate is approved. Final SDK refresh and actual notice
applicability/native/font/asset closure remain open.


The [Mac app notice payload checkpoint](../docs/reports/macos-app-notices/README.md)
adds explicit binary-bound notice bundles to private `.app` packaging, including
original-byte copying, staged/final executable and manifest validation, bounded
member/reference checks and guarded cleanup. Thirty-six focused test methods
and a headless real-binary package pass. Its selected 109 declarations remain
partial; five native-library Go graph modules lack recorded declarations, and
the mismatched source-built native notice bundle is excluded. No app is launched,
no signing is requested and M6 rights/final runtime acceptance remains open.

The [managed grid body-layout checkpoint](../docs/reports/macos-grid-cell-surface/README.md)
replaces supported body Text widgets with current-projection pane layout values.
Five prototype pairs per CPU/GPU improve median run-p95 by 1.12/0.77 ms, with
5/5 improved pairs each; all run-p95s still exceed 4 ms. Corrected source passes
headless and separate native functional route checks. Whole-process allocation
p50 rises about 15.21%/10.17% despite lower retained Go heap. This advances the
M3 implementation; the prototype timing remains separate from later current-source
qualification, with every original target unchanged.

The subsequent [current-grid matrix](../docs/reports/macos-grid-current/README.md)
completes all eight table cases: 4,800 ordinary samples, five runs per case,
matched 10k/1M geometry, all forty p99s within 6 ms and seven p95s within 4 ms.
Only fixed-height million-row GPU meets both in all five runs; the full M3 gate
remains open. Separate sampled allocation diagnostics locate substantial new
allocation at the per-cell render-style copy. The subsequent
[render-local style experiment](../docs/reports/macos-grid-paint/README.md) passes
correctness checks and reduces sampled allocation, but is rejected: CPU/GPU
median run-p95 gains are only 0.057835/0.150298 ms, CPU improves 3/5 pairs and
median p99 worsens. The fixed 0.4 ms/4-of-5 both-mode bar fails; no runtime is
adopted and no conditional full candidate matrix runs. Retire this style
experiment. The full native workspace record now passes at both row counts; next close
M2's remaining native catalogue coverage. A materially different
measured M3 decision and the remaining table/tree/general-UI matrix still precede
full resource/endurance qualification and M7 closure.

## Current position

| Area | Implemented and demonstrated | Still prevents completion |
| --- | --- | --- |
| Application/API | Public Go APIs, controls/themes, commands/tasks, documents/history, docking and independent windows | Complete composed workflow and lifecycle acceptance on a release candidate |
| Data and canvas | Managed tables/trees, typed placement, retained drawing/images, compute/HDR and custom 3D | Frame targets, representative interaction, full resource accounting |
| Mac integration | Cocoa input/accessibility, menus, clipboard, choosers, drops, tray and optional notifications | Human/native workflow coverage, scale/display changes, declared service support |
| SDK | Current relocated minimum-Go SDK, all default main builds, font-byte mapping and four native source-assertion/readback cases | Final clean candidate including the M1 cancellation fix, component/rights closure and distribution acceptance |
| Performance/endurance | Existing profilers, ordinary benchmarks, scoped idle samples and historical 1,000-cycle campaigns | Ordinary targets still fail; complete current-source memory/pacing/idle/overnight qualification |

The first automated desktop placement attempt expired without input. The second
attempt also sent no pointer/key input: the private operator refused an unusable
Foundation process-birth value. A [fresh current-source attempt](../docs/reports/macos-placement-session/README.md)
uses the corrected kernel identity and session stop scope, reaches native AX
readiness and posts a click/Return to the owned PID, but observes no placement.
It fails and is sealed without retry. Direct/manual verification is now required;
no further native input operator work is queued. Existing offscreen routing and
model tests remain separate evidence.

The [pinned Go-module declaration audit](../docs/reports/macos-module-declarations/README.md)
confirms that all five WebGPU linker-module notice gaps exist in their exact
upstream root/module-root trees, with matching selected Go source bytes. These
are absent declarations, not collector omissions. The82-component source-built
native archive bundle does not close those Go-module gaps. Applicable declarations
or a reviewed dependency replacement remain required; the distribution audit
stays pending. No runtime, SDK or Mac milestone changes at this checkpoint.

## Ordered delivery backlog

Each row ends in a stable, reviewed commit and a short result entry. Functional
fixes discovered later receive their own commit and invalidate only the affected
acceptance results. Rows can overlap only as described under execution order.

| ID | Work and concrete deliverable | Exit criterion | Desktop needed |
| --- | --- | --- | --- |
| M1 | Fix SDK command cancellation ownership in `scripts/child_process.py`; cover child creation, active communication and repeated cancellation during cleanup. Preserve caller signal policy and normal output/error behavior. | Real headless process tests show SIGTERM and SIGINT under supported POSIX main-thread/default-handler conditions cancel the owned child group and propagate failure, including the creation boundary; no live owned descendant remains. Existing SDK-script regressions pass. Worker-thread/custom-handler limits are explicit. | No |
| M2 | Close the representative application workflow using the existing public workbench/replay consumers and controls catalogue. Run the full 216-operation professional workspace record, plus command/history, worker cancellation, save/open/dirty close, dock/float/redock, two-window survival and large-data edits. Add only missing assertions or product fixes needed for these paths. | The complete record matches model, focus, semantics, selection, geometry and defined captures across owner batches 1/7/64. Native smoke variants cover 10k/1M rows with the same domain validations. The existing 18-action native smoke is not substituted for the full record. All controls in `controls.md` and linked dedicated contracts have a named acceptance scenario and gaps are resolved. | Headless work now; native integration later |
| M3 | Qualify measured 60 Hz responsiveness on representative table, tree, text/layout, canvas, input and resize workloads. Add missing actual presentation feedback before timing acceptance. Repair demonstrated release-path failures; retain CPU work budgets as optimization targets. | The [current Mac release policy](macos-responsiveness.md) passes with source/configuration-bound native presentation and input/pacing evidence. Historical 4/6 ms CPU misses remain reported; they alone do not block release or require another rendering redesign. | Quiet desktop for native measurements |
| M4 | Close memory/resource and integrated endurance gates after the functional/performance fixes stabilize. Attribute Go, native/font and framework-controlled GPU retention; repair unbounded ownership or over-budget components. | Required memory/allocation budgets pass; at least ten minutes of composed activity, 1,000 composed lifecycle cycles and one complete eight-hour visible/occluded/minimized/background soak finish with joined workers, clean tracked teardown and no unbounded retention. Full idle CPU, wakeup and present requirements pass, not CPU alone. | Quiet desktop; overnight for soak |
| M5 | Complete real Mac usability: palette placement, canvas selection/pan/zoom, docking, keyboard-only controls, text/IME and VoiceOver, menus/choosers/clipboard/file drops, tray/notifications as declared, resize and display/DPI transitions. Use the existing demo suite and probe. | The native workflow matrix below is exercised with expected domain results and cleanup. Automated routing, PID-directed native input and physical/human verification are recorded separately. Required human/display/IME cases remain open until actually performed. | Yes; some cases require a human or display hardware |
| M6 | Close the actual Go/native/Rust/font/asset component and notice inventory for shipped bytes; produce the final clean SDK and example application bundle. Verify independently from a relocated extraction without checkout paths or development-only environment variables. Document tested OS/GPU/toolchain and build/signing/notarization steps. | All dependency/rights and release gates in `release.md` are satisfied for the intended artifact; generator, public contracts and all shipped demos build, required native consumers run, and the packaged app works outside the source tree. Signing/notarization or public publication that needs owner identity/approval remains an explicit external action, never a simulated pass. | Most inventory/build work now; packaged app acceptance later |
| M7 | Audit the frozen candidate against this plan, `macos-completion.md` and every applicable P1–P11 requirement in `prototype-plan.md`. Update tutorial, support/limitations and migration notes. | Every required item has direct evidence for the candidate or justified reuse of identical affected bytes/configuration. No missing, failed, indirect or unavailable required gate is marked complete. Record the final local commit/artifact identities and provide runnable consumer instructions. | Final checks depend on earlier rows |

## Budget matrix retained from the specifications

The complete [measurement contract](prototype-plan.md#measurement-contract)
remains authoritative. This table identifies the closure work; it does not reduce
its workload, cold/warm, display, repeated-run or measurement-boundary requirements.

| Gate | Required threshold / workload |
| --- | --- |
| Table | Warm UI work p95 <=4 ms and p99 <=6 ms; fixed/variable rows at 10k/1M with at least 400 nonempty clipped cells, matched geometry. Follow `grid-performance.md`. |
| Tree | Warm UI work p95 <=4 ms; stored 10k/100k wide/deep cases and the original P6 expansion/filter/visible-node workload. The current >=64-row diagnostic alone does not qualify the original <=500-visible-node workload. Follow `tree-performance.md`. |
| General UI | UI CPU work p95 <=4/3/2.5 ms for 60/120/144 Hz; warm layout and unchanged cached text each p95 <=1 ms for 1,000 controls/runs; new 2,000-code-point paragraph shaping/layout p95 <=4 ms. |
| Startup | Warm p95 <=300 ms; cold p95 <=1,000 ms to the specified first interactive presented frame. Snapshot readiness alone does not measure this. |
| Canvas/GPU | P4 combined CPU critical work p95 <=4 ms and GPU scene+UI p95 <=5 ms for the 60 Hz tier; chrome/text GPU p95 <=2 ms at 1080p. Preserve the 4096-square image, million-sample waveform, 100k indexed primitives and custom 3D/compute cases. |
| Input/pacing/resize | Input-to-submission p95 <=5 ms at 60 Hz / <=3 ms at 120/144 Hz; input-to-presentation p95 <=33.3/20 ms; >=99% of frame intervals within 1.1 times the target period; report every >2-period stall. Resize UI work p95 <=8 ms and no application-caused stall >50 ms. |
| Baseline memory | Live Go heap <=40 MiB; process private/resident target <=150 MiB for the specified 1,000-control/font baseline without domain data; framework-controlled GPU <=96 MiB at 1080p. Attribute native/driver contributions. P4's separate tracked GPU budget starts at 512 MiB with eviction/refusal. |
| Allocation/tasks | Warm baseline <=64 KiB/frame and <=100 allocations/frame; zero idle frame work. Task progress and cooperative cancellation acknowledgment <=100 ms; worker exit reported separately. |
| Idle/endurance | 60-second settled intervals: mean CPU <0.5% of one logical CPU, <=2 application-originated wakeups/s and no repeated presents, including occluded/minimized states. Complete >=10-minute integrated activity, 1,000 lifecycle cycles and eight-hour soak with bounded post-warmup retention. |

Use the original 1920x1080 logical primary workload with roughly 1,000 visible
controls/text and an active canvas, plus the specified 2560x1440/144 Hz and
3840x2160/200% variants. Actual physical geometry/refresh must be observed.
Missing hardware, unavailable GPU timestamps or absent presentation feedback is
an open evidence item; scalar readback or a timing estimate cannot silently
replace it. Any proposed threshold revision must have an explicit decision and
recorded workload/hardware basis before it changes an acceptance result.

## Native workflow matrix

| Workflow | Required result |
| --- | --- |
| Palette to canvas | All component types; threshold/capture/visible drag and target previews, acceptance/rejection and optional snapping; transformed drop position after pan/zoom/scroll/scale; exactly one command/history entry; same-object Undo/Redo; release consumption; Escape/outside/capture loss/source removal/target removal/window close cancel with no object and no retained reservation. Keyboard placement uses the same validation. Same-window scope stays explicit. |
| Controls/workbench | Catalogue controls, focus/tab/shortcuts, modal precedence and disabled/stale refusal; property editing, selection, docking/undocking/redocking and independent windows; no duplicate command or stale mutation after removal/closure. |
| Text and accessibility | Exact Unicode/grapheme and mixed-direction editing; text-range geometry; native composition/candidate placement after scroll/focus/scale; actual keyboard/VoiceOver task completion over controls, text and virtual data. Native selector queries alone do not qualify screen-reader use. |
| Native services | Accepted/canceled save/open and dirty-close paths; same command validation from menu/shortcut/automation; clipboard representations and restoration where tests access the general board; actual external file-drop workflow and declared inbound/outbound DnD support; tray/notification capability and lifecycle behavior. An unsupported declared capability is recorded as such, never inferred from a missing test. |
| GPU/window/display | Resize, minimize/restore, occlusion, independent views, clipped/alpha-correct composition, supported monitor/DPI transitions, resource rebuild and declared device-loss policy. Distinguish controlled fault injection from actual device loss. |

Run these on the frozen candidate and retain the tested OS, scale/display,
input/assistive-tool versions and method. Older macOS compatibility cannot be
claimed from a minimum-OS load command alone. Local hardware limits and any
additional Mac configuration needed for advertised support stay visible in M7.
No automatic input-source, VoiceOver, clipboard or preference changes: use only
explicitly authorized tests and restore any authorized temporary state changes.

## Execution order and work while the Mac is occupied

For current M3 work, the [release policy's immediate tasks](macos-responsiveness.md#immediate-completion-work)
supersede architecture-first directions in the historical experiment notes below.
Measure 60 Hz presentation/input behavior before proposing another renderer change.


1. **M1 source checkpoint complete.** The production SDK cancellation fix and
   its 107-method headless acceptance are [recorded here](../docs/reviews/macos-sdk-cancellation/README.md).
   The existing private SDK archive remains historical; M6 includes the fix.
2. **M2 headless checkpoint complete.** The full public professional replay
   passes all six frozen records. The [catalogue map](control-acceptance.md)
   includes routed control regressions and named native records. The complete
   [native workspace record](../docs/reports/macos-native-workspace/README.md)
   passes at both row counts; native catalogue coverage remains partial. Physical/
   human workflows and full M2 acceptance remain open. Software and native
   evidence retain their distinct scopes.
3. **M3-A, tile culling, row fragments and resolved shaping are rejected.** The complete table
   matrix still fails. No follow-up AX, packed-upload or tile-culling microtuning
   is queued. The [composition checkpoint](../docs/reports/macos-scene-composition/README.md)
   measures 480 actual packets and the failed five-pair culling comparison.
   Smaller input alone does not improve total-frame work. The subsequent
   [row-fragment prototype](../docs/reports/macos-row-fragments/README.md) fails
   its actual reuse feasibility gate before ordinary timing. A separate M3
   architecture decision is required; no further row-cache tuning is queued.
   The independent [M6 minimum-toolchain source check](../docs/reports/macos-minimum-go/README.md)
   now passes; final SDK/rights closure remains open.
   A [fresh current private SDK](../docs/reports/macos-sdk-stage0079/README.md)
   now includes M1/stage0079 and passes actual Go 1.25.6 relocated consumption,
   all 26 default-main builds plus the tagged lifecycle helper and 28-font byte/
   metadata mapping. [Four serial native cases](../docs/reports/macos-sdk-native-stage0079/README.md)
   now pass with the same binaries/archive and confirmed owned shutdown. Broader
   native/human workflows and rights/full acceptance remain open.
   The [resolved-shaping checkpoint](../docs/reports/macos-resolved-shapes/README.md)
   then proves real CPU/GPU reuse and passes correctness, but ordinary median
   run-p95 improves only 0.074/0.076 ms, below the 0.4 ms adoption bar. Its whole
   private overlay is rejected; all twenty p95s fail 4 ms. No shape-cache or other
   small rejected-boundary tuning is queued. Complete missing tree/general UI
   qualification and independent M6 inventory work. The [fresh SDK tree cost
   boundary](../docs/reports/macos-tree-boundary/README.md) now records two ordinary
   runs of 240 measured events each and two separate diagnostics of 120 each,
   all after 128 warmups per run (480 ordinary / 240 diagnostic samples).
   Ordinary CPU/GPU-requested p95s are
   7.183709/7.217669 ms, both failing, with matched 68–69 clipped rows. The
   [implemented tree/local-layout prototype](../docs/reports/macos-tree-local-layout/README.md)
   passes correctness, race/vet/build and positive native-route checks, but is
   rejected after five ordinary pairs per mode: median run-p95 worsens CPU
   7.293207 → 7.848000 ms and GPU 7.071166 → 7.531169 ms, with 0/5 improved pairs
   each. Both median p99s regress; all twenty p95s fail 4 ms. No runtime, identity
   API, prepared dependency or SDK change is adopted. Retire this prototype
   without local-guard relaxation or row/cache tuning. M3 requires a different
   material scene-publication/rendering decision; proceed with independent M6
   shipped-component/rights closure while that remains unresolved.
4. Use an authorized desktop window for other M2/M5 native checks and ordinary M3
   measurements. Never replay a completed/human-return seal.
   The one fresh bounded placement attempt has now failed its delivery observation;
   retain its result and use direct/manual verification. Do not run another input
   operator attempt or spend another open-ended cycle building an operator.
5. The unchanged SDK now passes a [two-cycle lifecycle and four-state idle
   checkpoint](../docs/reports/macos-sdk-lifecycle-stage0079/README.md), including
   four eligible 60-second CPU observations. The original sandbox/self-RSS
   harness failure is preserved; direct execution uses the historical lifecycle
   policy without claiming IP denial. This is a bounded preflight/CPU result,
   not full idle or endurance acceptance. Extra queue admissions remain
   unattributed and one hidden Cocoa window remains unclassified; native-window
   deallocation/leak closure and loss-free delivery are open. After significant functional/performance
   fixes stabilize, rerun affected preflight evidence, then ten-minute activity,
   1,000-cycle campaign and overnight soak in that order. Each needs its own
   reviewed admission; these scoped passes do not launch a long campaign.
   Stop on a real failure; diagnose and rerun the affected gate.
6. Finish M6 packaging from the frozen passing runtime, then M7 final audit.
   A package/source difference that affects a passed gate requires the relevant
   check again. Documentation-only changes do not require rebuilding every SDK.

Live tests are stopped while the user is online unless separately authorized.
Do not remove human-return or completed-case seals. Read-only inspection,
headless regressions, source fixes, documentation and packaging preparation can
continue. Do not request desktop access again while independent work is available.

## Completed M3-A experiment and next boundary

The [matrix decision](../docs/reports/macos-grid-matrix/README.md) specifies an
owned segmented ready accessibility snapshot: construct each eligible leaf grid
once into its final detached group, serialize an ordered reference view and share
that immutable group with the existing ready baseline/cache. Keep callbacks and
guards separate. Preserve independent per-grid storage and all retirement fences.
Pending/editor/custom/namespaced/uncertain cases keep the full old route before
state mutation. Preserve nil/empty wire distinctions and custom callback count.

Prove the fast path actually runs and compare complete legacy wire, order,
actions/generations, interleaved scopes, pending, namespaces and lifecycle. Full
race/vet/native checks precede five ordinary interleaved comparison pairs.
Experiment adoption requires >=0.4 ms lower median run-p95 and >=4/5 paired
improvements without correctness, retention or tail regression. Failure means
reject and stop small follow-up tuning of that candidate. This adoption bar does
not replace the unchanged 4 ms p95/6 ms p99 acceptance matrix, including CPU mode.

A new glyph-span shader is not selected: current-DPI bitmap routing does not
establish a span bottleneck. Any later scene/resource redesign first needs actual
operation/resource-kind counts and encoded-buffer composition for the failing
fixture. Cocoa JSON parsing is asynchronously main-queued; do not attribute it
as synchronous ReconcileNS. M3-A was implemented, tested and rejected: see the
[complete experiment report](../docs/reports/macos-segmented-ready/README.md).
Median run-p95 rises 4.749206 → 4.866957 ms; all five pairs lose despite 10.82%
lower median of per-run nearest-rank p50 process allocation. Full driver race, independent frozen
wire/geometry/action/lifecycle checks and both JSON variants pass. Separate
native diagnostics prove positive segmented routing. All eleven source/test
changes and the diagnostic field are archived and reverted; no runtime change
is retained. One scoped Prism review covers the final rejection checkpoint.
The next desktop work is the pending M2/M5 workflows, followed by evidence for
any materially different M3 implementation. No small follow-up M3-A tuning.

## Rejected M3 prototype: retained managed-row paint fragments

The following is the original prototype scope, retained as history rather than
queued work. It was implemented privately to avoid repeated row glyph/path
capture on unchanged managed rows, shared by CPU and GPU renderers. This is a material
retained-work boundary, not another operation-buffer allocation adjustment.
Keep the original shader and encoding initially. Retain detached immutable
integer paint operations and instantiate/replay them in their original painter
order. Reuse requires proof from actual current content, shape, font, paint,
clip and placement inputs; a source stamp, model revision or borrowed layout
pointer alone is insufficient. Do not build a full VerifyPaint plan each frame
merely to decide whether reuse is possible.

Admission must preserve original work/byte limits, exact clip and alpha rules,
partial repaint and pending/editor/unsupported fallback. At scale 1.6875, actual
rounded row heights and fractional glyph phases can differ: do not assume every
row shares a scroll delta or permit unproved translation. Selection, edits,
source replacement, theme/font/scale/width changes and row reuse/removal must
invalidate affected fragments. Bound retained bytes/entries and font references;
clear authority references and retire private state on resize, release and close.

Use an independent uncached renderer/pixel oracle across fixed/variable rows,
selection/edits, scroll/clip edges and supported scales; compare captured replay
and real native GPU readback. Positively observe current-fixture reuse separately
from ordinary timing. Both CPU and GPU paths need a matched five-pair ordinary
comparison; adoption requires >=0.4 ms lower median run-p95 and >=4/5 improvements
without correctness, retention or tail regression for each applicable path.
The full unchanged 4 ms p95/6 ms p99 matrix remains required afterward.

Reassess feasibility after 90 minutes of implementation/investigation. If actual
input proof, current-fixture hits or material full-frame benefit cannot be
established, stop this prototype, retain the failure and make a new architecture
decision. Do not iterate into another small allocation-only candidate or rerun
unchanged comparisons. The [implemented prototype](../docs/reports/macos-row-fragments/README.md)
passes final race, pixels, native GPU, vet and build checks, but is rejected at
feasibility: GPU body-row reuse is 5% across 120 frames and absent in 108 frames,
while owned storage reaches 8,387,889 / 8,388,608 bytes. CPU app exit is zero but
its observer flag was filtered, so CPU reuse is unavailable and the combined
diagnostic controller fails. No ordinary latency comparison is run and no CPU
performance conclusion follows. Stop the whole prototype without observer retry,
cache/coordinate relaxation or further architecture branches in this experiment.
Production and stage0079 remain unchanged; no acceptance gate closes.

## Rejected M3 prototype: retained tree updates and local layout

The following is the original bounded scope, retained as history. The
[implemented result](../docs/reports/macos-tree-local-layout/README.md) rejects the
whole private overlay after its completed native comparison; it is not queued work.

The [current tree cost boundary](../docs/reports/macos-tree-boundary/README.md)
identifies roughly 3 ms of mean ordinary reconciliation plus layout/paint work.
`syncTree` currently reassigns visible-index native rows and recursively updates
its host on scrolling. Reuse the existing grid retention/local-layout ownership
infrastructure for a bounded tree experiment; this is not another rejected
AX/upload/culling/paint-fragment/shaping-cache change.

Keep at most the current admitted 500 rows, retain overlapping native rows by
stable domain key, use each departing donor once and retire surplus rows. Update
and size changed text/state; do not run full recursive style/layout for unchanged
retained children when the root's actual scene, parent, allocation, resolved
style, scale, row configuration and accepted immutable index are proved unchanged.
Equal source stamps and row counts are insufficient index identity. If the public
model lacks a safe identity, add a comparable non-retaining capability around the
existing private index token carried in the same snapshot, or use the complete
fallback; never retain the model, data or old index through that capability.
Refresh current OutlineRef, epoch and AX authority every generation: widget
retention must not preserve stale input authority. Prove fractional-DPI sizing and
queued-text cancellation; key identity alone cannot prove intrinsic size reuse.

Initially admit viewport-only changes. Pending, selection, topology, uncertain,
theme/font/style/scale/width and ownership changes take the original complete
route. Include footer status, pager disabled states, scrollbar state, inherited row selected/
disabled styles, global layout/pending fences, completed semantics/geometry,
clipping, modal/stale refusal, callbacks and lifecycle in the proof. Local layout
must preserve the root's outer allocation and complete scene painter order.

A frozen full route supplies independent geometry, software pixels, semantics,
actions and lifetime comparisons. Positively observe the native fast path in a
separate diagnostic. Required race/vet/native checks precede five ordinary
interleaved original/candidate pairs for each CPU/GPU mode. Keep the existing
>=0.4 ms lower median run-p95 and >=4/5 paired improvements in both modes,
without tail, correctness or retention regression. Reassess after 90 minutes of
implementation/investigation, with an explicit adopt/reject result before further
work. This prototype is now implemented privately and rejected before the
90-minute deadline. Its two diagnostics observe actual local-root use and CPU/GPU
routing; five pairs per mode (240 measured events each after 128 warmups) regress
median run-p95/p99, with no improved pairs. Correctness success is insufficient
for adoption. No production change follows. The full tree/general UI acceptance
matrix, unchanged table failures and every other Mac gate remain open.

## Existing execution entrypoints

Use the pinned prepared dependencies and controlled environment from
[development.md](../docs/development.md). Reuse these entrypoints rather than
introducing a second runner:

- M1: `python3 -B scripts/child_process_test.py`, followed by the affected
  `build-macos`, `release-sdk`, `check-sdk` and `package-app` Python regressions.
- M2: from `examples/replay`,
  `go test -race -tags offscreen ./workspace -count=1 -timeout=9m` executes the
  fixed workspace records. Do not run the authoring mode or regenerate goldens
  to make a regression pass. The suite must cover both row counts and batches.
- M3: the existing `scripts/benchmark-grid.py`, tree performance consumer and
  frame metrics described in [grid-performance.md](grid-performance.md),
  [tree-performance.md](tree-performance.md) and the measurement contract.
- M4: the existing workbench endurance/lifecycle/soak modes described in
  [workbench-endurance.md](workbench-endurance.md).
- M5: [placement probe](../examples/placement/README.md), professional workspace
  smoke and the service/editor/accessibility fixtures in their contracts.
- M6: `scripts/release-sdk.py`, `scripts/check-sdk.py`, `scripts/new-app.py`,
  `scripts/build-macos.py` and `scripts/package-app.py` under [release.md](release.md).
  Their publication, runtime and signing options require the corresponding scope.

`scripts/check.sh` remains the broad regression entrypoint at a stable
implementation checkpoint; it is not rerun after each documentation edit.

## Limits on investigation, testing and review

- A checkpoint must close a named backlog item or demonstrate a concrete blocker
  and the next implementation action. A new report or harness by itself does not
  count as feature or acceptance closure.
- Reuse existing tests, profiles, probes and immutable results. Add verification
  code only when an identified required assertion cannot otherwise be tested.
  Use a 90-minute investigation timebox before explicitly reassessing a failed
  approach. A timebox never waives the gate; a renderer that cannot meet the
  release target requires an implementation/architecture decision against the same goal. A provisional optimization-target miss alone does not trigger this rule.
- Test changed behavior and its real dependents first. Run the broader headless
  suite once per stable implementation milestone; repeat only for new failures,
  affected code changes or unresolved concerns. Run native measurements only on
  an admitted, stable candidate. Do not pool different modes/workloads/builds.
- Use one internal code review per significant implementation checkpoint. Use
  Prism with Anthropic Claude Sonnet 5.5 once for the completed implementation
  checkpoint; resolve material findings and review only necessary changed scope.
  Documentation-only progress updates do not trigger another implementation review.
- Keep one concise result entry: requirement, source/artifact identity, command,
  actual outcome, remaining failure and next action. Link raw logs/profiles;
  preserve failures. Do not create recursively copied historical checkpoints or
  duplicate unchanged SDK archives/binaries for every documentation commit.
- Commit stable product fixes independently. Stop broadening this milestone.
  Report completed gate IDs and remaining blockers after each checkpoint, rather
  than a percentage based on code volume or the number of passing small tests.

## Forecast and closure ledger

M3 requires actual presentation/input evidence for the owner-selected 60 Hz
release gate. Historical CPU-work misses remain optimization evidence and do
not establish a failed 60 Hz user experience. Measure the existing candidate
before selecting another rendering experiment; keep the previous rejected
experiments and their original results intact.

M1 is complete in source. M2's current workspace/text and five catalogue records
pass, with three dedicated Cocoa methods still open. The private SDK passes its
scoped minimum-Go/relocation/native checks; final candidate and rights closure
remain M6 work. The ten-minute activity, 1,000-cycle and eight-hour final soak,
human accessibility/input verification and supported display checks remain
required. Estimate remaining time from those actual results.

| Milestone | Current status | Next action / evidence needed |
| --- | --- | --- |
| M1 | Complete in source; incorporated in current private SDK | [Reviewed fix and 107 headless tests](../docs/reviews/macos-sdk-cancellation/README.md); exact helper bytes are included in the [current SDK](../docs/reports/macos-sdk-stage0079/README.md). Full final artifact acceptance remains in M6. |
| M2 | Partial; prior scoped records pass; three Cocoa methods and affected stage88 native renewal remain open | [Current workflow](../docs/reports/macos-current-workflows/README.md) qualifies native workspace 432/54 and text 68/33. [Current catalogue/audit](../docs/reports/macos-current-catalogue/README.md) qualifies five records 255/169 and maps actual assertions for 205 executed tests. Dedicated tab-close, Breadcrumb and ColorPicker Cocoa-selector methods remain unqualified; renew and review those before M2 closure. Physical/human workflows remain M5. |
| M3 | Open; measured 60 Hz presentation/input acceptance is pending | [Current table matrix](../docs/reports/macos-grid-current/README.md) misses the historical 4 ms optimization target in 33/40 runs; all 40 meet 6 ms p99. It does not measure actual presentation. Under the [owner-selected policy](macos-responsiveness.md), qualify the [implemented recorder](native-presentation.md) with real native feedback, measure the existing candidate and fix actual 60 Hz failures before choosing another renderer experiment. Prior rejected experiments and all original results remain preserved. |
| M4 | Partial; current SDK preflight and sampled idle CPU pass | [Two composed cycles/66 native checks and four eligible 60-second CPU intervals](../docs/reports/macos-sdk-lifecycle-stage0079/README.md) pass with owned process-group absence and framework-tracked retirement. Extra admissions remain unattributed; one hidden Cocoa window remains unclassified, so native deallocation/leak closure stays open. Original sandbox harness failure is preserved. OS wakeups/presentation, full memory/resource accounting, ten-minute activity, 1,000-cycle and overnight gates remain open. |
| M5 | Partial; foreground tests deferred while owner is online | Native fixtures exist, but physical/human/display workflows and palette input acceptance remain open. |
| M6 | Partial; current SDK incorporation and scoped consumption pass | The [current SDK](../docs/reports/macos-sdk-stage0079/README.md) (7,879 payload files plus `SDK_MANIFEST.json`, 7,880 total) passes actual Go 1.25.6 relocated consumer/public-contract checks, 26 default main builds plus one tagged helper and 28-font metadata mapping. [Four native SDK cases](../docs/reports/macos-sdk-native-stage0079/README.md) pass source assertions/readbacks/owned closure. The [standard-library input checkpoint](../docs/reports/macos-stdlib-inputs/README.md) now binds selected Go package source/declaration bytes and refuses their drift in owned builds; it does not approve notice applicability. Final candidate/application-bundle acceptance, broader native/human workflows, exact runtime/font/non-font/generated-asset rights and owner distribution approval remain. |
| M7 | Open | Requirement-by-requirement final audit after all preceding gates close. |

The ledger changes only when the corresponding requirement is demonstrated.
Completion is not announced with an open required row.


### Direct grid accessibility construction checkpoint

The [direct construction result](../docs/reports/macos-grid-ax-construction/README.md)
records an experiment with passing race/vet/internal semantic checks and a Prism
review, but slower native p95 in all five matched pairs. The runtime candidate
was rejected and reverted. The local construction reduction does not close M3; all unchanged
frame thresholds and the complete table/tree/application matrix still apply.
The next bottleneck investigation must use full refresh/scene capture and native
publication costs, rather than extrapolating another small isolated allocation
win into framework completion. M4–M7 and native M2/M5 remain open.

### Structural controls checkpoint

The [native structural result](../docs/reports/macos-structure-catalogue/README.md)
adds controlled tab keyboard navigation with burst/read-only lifetime fixes.
The new 50-input/42-state native scenario fails after41 inputs/35 source-asserted
observations, before document selection/close results. Preserve the sealed failure;
next record measured acquisition status/attempts without relaxing Frame success
or ordinary retries, and resolve the original disabled-layout equality oracle
using same-style restoration or independently checked state-specific metrics
before a new scope. Full structural catalogue,
physical/pixels/M2/M5/M3–M7 remain open with unchanged thresholds.

The structural native prefix above predates the subsequent Prism tab activation
and disabled-focus fixes. Those fixes pass actual Core queue/manager routing and
complete driver-root/gallery race suites. At that prior checkpoint they had not
received a new native run. The failed diagnostic attempt below now exercises
those fixes but does not complete structural qualification. Both failed scopes
remain sealed and the full structural/native gate stays open.


### Measured acquisition checkpoint

The [acquisition diagnostics result](../docs/reports/macos-acquisition-diagnostics/README.md)
adds measured-only actual source/stage/status/attempt detail without changing
ordinary retry/release behavior or strict completed-frame criteria. A distinct
native50/42 scope fails after16 inputs/15 source-asserted observations with
actual `occluded(196609)` on one acquisition attempt. Cleanup and separate whole
post-failure boundaries pass; full native validation remains open. This result
does not identify the older failure's cause. Both scopes remain sealed.
The new style-aware oracle has separate public Manual160/162DPI and41 synthetic
controls, with no complete native oracle pass. Complete driver-root (`.`)/gallery race, focused
GPU/Core regressions, build, prepare and affected-package vet pass; default Core
vet retains12 exact baseline diagnostics. That checkpoint proposed visibility-aware
Frame scheduling. The typed-terminal decision below preserves the existing
contract while an authoritative drawable-ready event remains unestablished.
Full M2/M5 and M3–M7 remain open; performance targets and release gates are unchanged.


### Typed unavailable Frame checkpoint

The [typed terminal result](../docs/reports/macos-frame-unavailable/README.md)
preserves the first-composition contract and classifies checked known surface
yields through `DesktopCompositionError`/`ErrDesktopFrameUnavailable`. No failed
frame token, automatic acquisition/requeue or readiness inference is introduced.
Cancellation identity is checked only after the actual trace terminal; capacity
and lifetime rules remain bounded. Late cancellation may race a committed reply
and does not retract prior work. Contradictory metadata and source-detail
faults fail closed. Full driver-root/gallery race, focused controls, affected vet,
build and deterministic stage86 preparation pass; default Core vet retains12
baseline diagnostics. Synthetic measured Manual results are software evidence.
No new native case runs; the earlier16/15 and41/35 failures stay sealed. Automatic
pending/resumption needs an authoritative drawable-availability event that is
not established. Continue remaining native catalogue/typed-error acceptance with
explicit availability admission and no retry loop. M2/M5/M3–M7 and all unchanged
performance/release gates remain open.

The [media/chrome catalogue checkpoint](../docs/reports/macos-media-catalogue/README.md)
completes 71 native Go FIFO inputs/42 observations for shared Image, Toolbar,
StatusBar, FormGroup, Breadcrumb and ColorPicker views. Actual focus, measured
pointer targets, overflow, accepted/rejected/disabled/stale/hidden states and
exact callbacks pass; original window/session/process retire. Public Manual
160/162 DPI routing and independently sampled immutable-image replacement pixels,
full gallery race/vet/build and 13 negative validator controls pass separately.
Initial checker/preflight failures remain preserved. This is controlled
Go 1.25.6/GOMAXPROCS2/SDK786a native routing; physical/OS input, IME/human accessibility, native
pixels/presentation and full M2/M5/M3–M7 stay open. No SDK refresh occurs.

Prism then identifies a real undispatched-owner wait on application retirement.
The final media runner cancels after Run and gives final close acknowledgement/
closed-reference validation a separate 1-second context within its 6-second join.
Two actual Application regressions discriminate the missing-cancel and canceled-
close counterfactuals. Final gallery race/vet/build and the distinct corrected
71-input/42-observation native case pass; both native scopes remain sealed.
One original Anthropic Sonnet 5.5 call and independent fix reviews are retained.


### Shared catalogue shutdown checkpoint

The [shared lifecycle fix](../docs/reports/macos-catalogue-lifecycle/README.md)
applies post-Run cancellation and the independently bounded final-close contract
to all five gallery catalogue modes. Four actual Application regressions, the
full gallery offscreen race suite, vet and native executable build pass. This
closes the known cross-catalogue consumer shutdown defect; no native application
runs, sealed scopes remain sealed and no full M2/M5/M3–M7 gate closes. Remaining
structural and text/modal native acceptance follows under a fresh reviewed scope.


### Text/editor/modal catalogue and unavailable consumer checkpoint

The [text/modal result](../docs/reports/macos-text-modal/README.md) adds an ordinary
public editor page/shared project field and dialog, plus the intended68-input/
33-state consumer. Full software160/162DPI, driver/gallery race, qualified
unchanged-root reuse, affected vet and build pass. The adapter preserves the
multiline renderer's existing required monospace family under UI typography.
The fresh native case **fails** at35inputs/19validated observations through
editor Tab on actual one-attempt occluded196609 acquisition. Typed unavailable
propagation/no retry and tracked session/owned-process cleanup pass. Disabled-
editor/modal/final-close native checks remain unqualified; the exact failed case
is sealed. No full M2/M5/M3–M7 gate closes or SDK/prepared bytes change. Continue
at the actual drawable availability/ownership boundary before another complete
structural/text/modal admission; do not infer readiness from shown state or
retry a sealed case. Earlier failed native cases retain their own unknown causes.

The [cursor ownership checkpoint](cursor-ownership.md) defines per-window native
ownership, bounded shared/per-window retention, theme synchronization and terminal
drain. Its automated fake-backend/race evidence does not close physical cursor,
DPI, visibility-mode or M4 native/endurance qualification.

The [desktop pointer mode checkpoint](desktop-pointer-modes.md) adds native-owned
visibility/relative/position operations and checked routing. Source/fake-backend
evidence supersedes the earlier status-only Hide/Show implementation gap. Cocoa
physical pointer/visibility/capture, DPI/focus transitions and M4/M5 live gates
remain open; measured 60 Hz remains the release gate and 4 ms an optimization
target.

The [event queue coalescing checkpoint](event-coalescing.md) preserves finite
scroll samples, checked raw/local movement origins and unique/deferred barriers
through queue compression. Automated race and original-code failure controls
qualify source routing only. Twelve inherited default Core vet diagnostics remain
tracked for the M7 source quality audit; physical input, actual presentation,
measured 60 Hz and all remaining Mac acceptance/distribution gates stay open.

The [Core source quality checkpoint](core-source-quality.md) clears the twelve
previously recorded default Core vet diagnostics. Ten literal sites preserve
their values with named fields; two unreachable profiling tails are removed while
normal benchmark bodies remain identical. Seven existing offscreen race methods
and benchmark compilation/listing pass. This resolves the specific source audit
follow-up from event coalescing; final M7 audit and native/Mac gates remain open.
