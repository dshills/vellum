# Mac framework delivery plan

Updated 17 October 2026. Baseline: clean `762dcbf`; the most recent private GPU
SDK and its native drawing/3D/imaging checks use clean runtime source `5eef841`.
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
| SDK | Relocated public consumers; current private SDK drawing/3D/imaging native checks | CLI cancellation ownership, final clean candidate, component/rights closure and distribution acceptance |
| Performance/endurance | Existing profilers, ordinary benchmarks, scoped idle samples and historical 1,000-cycle campaigns | Ordinary targets still fail; complete current-source memory/pacing/idle/overnight qualification |

The first automated desktop placement attempt expired without input. The latest
attempt also sent no pointer/key input: the private operator refused an unusable
Foundation process-birth value. Its kernel-identity correction is prepared but
unexecuted. These are test-operator outcomes, not placement acceptance. Existing
offscreen routing and model tests remain separate evidence.

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

1. **Next commit: M1.** Implement and test the production SDK cancellation fix.
   The private grouped test host is not that production fix.
2. **Following checkpoint: M2.** Complete the public professional replay and
   catalogue coverage using existing runners; correct actual application gaps.
3. **Then M3 is the main engineering priority.** A quiet native baseline and one
   targeted renderer/layout fix precede another broad qualification campaign.
   Analyze existing profiles and implement/test candidates offscreen while the
   user is working. Inventory work for M6 can proceed independently.
4. Use an authorized desktop window for M2/M5 native workflows and ordinary M3
   measurements. Reuse the prepared placement operator for one bounded attempt;
   if delivery still fails, record that result and schedule direct/manual
   verification instead of spending another open-ended cycle building an operator.
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
failed. A completion date before that is resolved would be speculative. The
next two deliverables are the M1 product fix and the M2 full replay/coverage
checkpoint, followed by a bounded M3 performance decision. After those, update
the remaining work estimate from the measured gaps. An eight-hour final soak,
human accessibility/input verification and required display availability must
be budgeted in addition to implementation time.

| Milestone | Current status | Next action / evidence needed |
| --- | --- | --- |
| M1 | Open | Patch the shared SDK child runner and execute cancellation boundary regressions. |
| M2 | Partial | Reuse existing domain/routing coverage; complete the full workspace record and map the control catalogue. |
| M3 | Failing / incomplete | Current ordinary frame targets remain unqualified; profile-to-fix-to-ordinary-comparison loop. |
| M4 | Partial | Historical lifecycle and scoped idle successes; full current candidate memory, integrated activity and overnight gates remain open. |
| M5 | Partial; live tests stopped | Native fixtures exist, but physical/human/display workflows and palette input acceptance remain open. |
| M6 | Partial | Private SDK native consumers pass; production CLI cleanup, final package and native/Rust/font/asset rights closure remain open. |
| M7 | Open | Requirement-by-requirement final audit after all preceding gates close. |

The ledger changes only when the corresponding requirement is demonstrated.
Completion is not announced with an open required row.
