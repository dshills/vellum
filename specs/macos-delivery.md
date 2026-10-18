# Mac framework delivery plan

Updated 17 October 2026. The latest complete table matrix measures clean
`bd672a6` with stage0079; the current headless-qualified private SDK uses clean `abe67c6`, Go 1.25.6 and
stage0079. The last SDK native drawing/3D/imaging checks remain historical
`5eef841` evidence.
Owner: the main Vellum development chat. This is the execution plan for the
existing [Mac completion requirements](macos-completion.md), not a replacement
for those requirements or a declaration that implemented APIs have qualified.

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

Full Mac completion requires every milestone below. An alpha label, compilation,
native readback or a smaller successful fixture cannot replace an open gate.
Public upload, tagging, signing/notarization with an owner's identity, and changes
to system input/accessibility preferences are not implicitly authorized here.

## Current position

| Area | Implemented and demonstrated | Still prevents completion |
| --- | --- | --- |
| Application/API | Public Go APIs, controls/themes, commands/tasks, documents/history, docking and independent windows | Complete composed workflow and lifecycle acceptance on a release candidate |
| Data and canvas | Managed tables/trees, typed placement, retained drawing/images, compute/HDR and custom 3D | Frame targets, representative interaction, full resource accounting |
| Mac integration | Cocoa input/accessibility, menus, clipboard, choosers, drops, tray and optional notifications | Human/native workflow coverage, scale/display changes, declared service support |
| SDK | Current relocated minimum-Go SDK, all default main builds and font-byte mapping; earlier SDK native drawing/3D/imaging checks | Final clean candidate including the M1 cancellation fix, component/rights closure and distribution acceptance |
| Performance/endurance | Existing profilers, ordinary benchmarks, scoped idle samples and historical 1,000-cycle campaigns | Ordinary targets still fail; complete current-source memory/pacing/idle/overnight qualification |

The first automated desktop placement attempt expired without input. The second
attempt also sent no pointer/key input: the private operator refused an unusable
Foundation process-birth value. A [fresh current-source attempt](../docs/reports/macos-placement-session/README.md)
uses the corrected kernel identity and session stop scope, reaches native AX
readiness and posts a click/Return to the owned PID, but observes no placement.
It fails and is sealed without retry. Direct/manual verification is now required;
no further native input operator work is queued. Existing offscreen routing and
model tests remain separate evidence.

## Ordered delivery backlog

Each row ends in a stable, reviewed commit and a short result entry. Functional
fixes discovered later receive their own commit and invalidate only the affected
acceptance results. Rows can overlap only as described under execution order.

| ID | Work and concrete deliverable | Exit criterion | Desktop needed |
| --- | --- | --- | --- |
| M1 | Fix SDK command cancellation ownership in `scripts/child_process.py`; cover child creation, active communication and repeated cancellation during cleanup. Preserve caller signal policy and normal output/error behavior. | Real headless process tests show SIGTERM and SIGINT under supported POSIX main-thread/default-handler conditions cancel the owned child group and propagate failure, including the creation boundary; no live owned descendant remains. Existing SDK-script regressions pass. Worker-thread/custom-handler limits are explicit. | No |
| M2 | Close the representative application workflow using the existing public workbench/replay consumers and controls catalogue. Run the full 216-operation professional workspace record, plus command/history, worker cancellation, save/open/dirty close, dock/float/redock, two-window survival and large-data edits. Add only missing assertions or product fixes needed for these paths. | The complete record matches model, focus, semantics, selection, geometry and defined captures across owner batches 1/7/64. Native smoke variants cover 10k/1M rows with the same domain validations. The existing 18-action native smoke is not substituted for the full record. All controls in `controls.md` and linked dedicated contracts have a named acceptance scenario and gaps are resolved. | Headless work now; native integration later |
| M3 | Resolve ordinary frame cost, then qualify table, tree, text/layout, canvas, input and resize against the existing workloads. Start from retained current profiles and an ordinary release baseline; implement one measured bottleneck fix at a time. Keep CPU/GPU modes and diagnostics separate. | The applicable budget matrix below passes on the required geometry/workload with five measured release runs, complete raw samples and unchanged thresholds. Larger datasets have the required matched visible work. GPU-route requests alone do not establish actual route coverage. | Quiet desktop for native measurements |
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

1. **M1 source checkpoint complete.** The production SDK cancellation fix and
   its 107-method headless acceptance are [recorded here](../docs/reviews/macos-sdk-cancellation/README.md).
   The existing private SDK archive remains historical; M6 includes the fix.
2. **M2 headless checkpoint complete.** The full public professional replay
   passes all six frozen records. The [catalogue map](control-acceptance.md)
   includes the new routed Selector/NumberField regressions. Native M2 integration
   remains queued; [actual results](../docs/reports/macos-workspace-acceptance/README.md)
   distinguish software routing from desktop acceptance.
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
   metadata mapping. Its native execution and rights/full acceptance remain open.
   The [resolved-shaping checkpoint](../docs/reports/macos-resolved-shapes/README.md)
   then proves real CPU/GPU reuse and passes correctness, but ordinary median
   run-p95 improves only 0.074/0.076 ms, below the 0.4 ms adoption bar. Its whole
   private overlay is rejected; all twenty p95s fail 4 ms. No shape-cache or other
   small rejected-boundary tuning is queued. Complete missing tree/general UI
   qualification and independent M6 inventory work before choosing a new M3
   redesign; preserve the table failure and every other gate.
4. Use an authorized desktop window for other M2/M5 native checks and ordinary M3
   measurements. Never replay a completed/human-return seal.
   The one fresh bounded placement attempt has now failed its delivery observation;
   retain its result and use direct/manual verification. Do not run another input
   operator attempt or spend another open-ended cycle building an operator.
5. After significant functional/performance fixes stabilize, run M4 preflight,
   ten-minute activity, 1,000-cycle campaign and overnight soak in that order.
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
  target requires an implementation/architecture decision against the same goal.
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

The largest uncertainty is M3: the unchanged frame targets have repeatedly
failed. A completion date before that is resolved would be speculative. M1 is
complete in source, and M2's full headless replay and catalogue map are complete.
The full matrix, rejected M3-A/tile-culling comparisons and rejected row-fragment
feasibility are recorded. M3 now needs a separate material architecture decision
against the same targets. Do not continue tuning the rejected prototype.
The independent [M6 minimum-Go source check](../docs/reports/macos-minimum-go/README.md)
passes; final SDK and rights closure remain open. Return to the architecture
decision; rerun the full matrix only after an adopted change. Update the remaining
estimate from those results.
An eight-hour final soak,
human accessibility/input verification and required display availability must
be budgeted in addition to implementation time.

| Milestone | Current status | Next action / evidence needed |
| --- | --- | --- |
| M1 | Complete in source | [Reviewed fix and 107 headless tests](../docs/reviews/macos-sdk-cancellation/README.md); final SDK incorporation remains in M6. |
| M2 | Headless record/map complete; native integration open | [Six frozen records and catalogue coverage](../docs/reports/macos-workspace-acceptance/README.md); native 10k/1M workflow and desktop catalogue checks remain. |
| M3 | Failing / incomplete | [Full current table matrix](../docs/reports/macos-grid-matrix/README.md): all 40 p95s fail 4 ms; GPU case medians 4.71–4.82 ms, CPU 5.61–5.77 ms. Matched 10k/1M visible work scales flat. Four separate GPU cases confirm actual acceptance, no fallback. [Composition and rejected culling](../docs/reports/macos-scene-composition/README.md) now rule out smaller references alone. [Row-fragment feasibility](../docs/reports/macos-row-fragments/README.md) also fails. Next: separate material architecture decision; no further AX/upload/culling/row-cache tuning. |
| M4 | Partial | Historical lifecycle and scoped idle successes; full current candidate memory, integrated activity and overnight gates remain open. |
| M5 | Partial; live checks authorized while owner is away | Native fixtures exist, but physical/human/display workflows and palette input acceptance remain open. |
| M6 | Partial | Private SDK native consumers pass; [native notice packaging](../docs/reports/macos-native-component-notices/README.md) now binds the source-built archive declarations. [Actual Go 1.25.6 source/driver/public-contract race, vet and native build](../docs/reports/macos-minimum-go/README.md) pass; no GUI launch or full SDK qualification. Final SDK refresh, relocated/all-demo/native checks, exact runtime/asset rights closure and owner approval remain. |
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
