# Current SDK lifecycle and short idle preflights

The unchanged private SDK from clean `abe67c6`, built with actual Go 1.25.6 and
stage0079, passes a fresh bounded two-cycle composed lifecycle preflight on this
Mac. It uses the already compiled tagged helper from the [SDK consumption
checkpoint](../macos-sdk-stage0079/README.md). The calling checkouts are clean at
`f875857` at terminal observation; changes since the SDK source are documentation
only. No SDK, source, prepared dependency or executable is rebuilt or modified.

## Lifecycle result

Two same-name composed window lifetimes complete 66 native Cocoa-selector
owner/model checks, with 1,000 fully contained dense labels in each lifetime.
Standard controls, controlled text, list acceptance/rejection, native tree
selection/disclosure, modal open/close, breadcrumbs and managed-canvas selections
complete. The replacement lifetime refuses the stale WindowRef, old callbacks
and old native resources. Framework-tracked window entries, callbacks, file panels and native
resources retire; workers join. Native Cocoa-window deallocation is separate. This is owned selector testing, separate from
physical pointer/key input and the still-failed placement operator attempt.

The run takes 22.150208 seconds, exits zero, reaches complete log EOF and has
process group 56617 observed absent. Its shared live-test lease remains held
through cleanup and terminal identity checks. There are 67 queue attempts and
admissions for 66 completed checks. Admissions do not count accepted callbacks;
the additional admission does not prove a duplicate callback or its retry cause.
The extra admission here and in the short trace remains an open anomaly: its
per-admission outcome and cause are not attributed. The fixture's passing model/
native completion predicates do not qualify loss-free delivery or a protocol
with exactly one admission per action.

Post-forced-GC samples span 20,159,288–59,211,880 bytes of Go heap and
187,416,576–350,568,448 bytes of RSS. They satisfy only the fixture's 512 MiB Go /
1 GiB RSS safety guards. This two-window composed lifecycle workload differs from the specified
baseline, so no baseline gate result follows. RSS is process-resident memory and
cannot isolate native-heap usage. Native/font and GPU retention attribution
remains open. One hidden, inactive, untitled Cocoa window
remains in the terminal snapshot. Its source and deallocation remain unclassified,
so native-window lifetime/leak closure is open. Process-group absence and tracked
retirement do not prove zero Cocoa windows. Menu restoration is also unqualified.

## Preserved harness failure

The first fresh run, with an added IP-denying sandbox wrapper, exits one after
1.865683 seconds before cycle zero: `fork/exec /bin/ps: operation not permitted`.
The local `/bin/ps` is setuid-root. The fixture's initial self-RSS subprocess is
incompatible with that wrapper; this is not lifecycle evidence. Its complete raw
failure, source/SDK identity and observed absence of process group 55643 remain
preserved under a sealed stop scope.

The independently reviewed fresh continuation removes only that incompatible
wrapper, retaining the exact helper, SDK, arguments, validators, watchdog and
owned cleanup. Direct execution matches the historical lifecycle fixture policy.
It makes no enforced network-denial claim. No production change or threshold
relaxation follows, and the failed run is not relabeled as passing.

## Short four-state preflight

The separate 40-second trace completes in 52.132136 seconds, exits zero, has
empty stderr and complete stdout, and retires process group 57375. Strict native
observations accept visible, fully occluded, minimized and allocating-worker
background phases. All 33 initial Cocoa-selector checks complete, with 34 queue
attempts/admissions; two worker results commit in the background phase.

Each phase has two one-second quiet samples. All eight samples match the actual
owned PID and exact versioned counter/read brackets. There are zero measured
primary paint wakes, empty-event-post attempts and surface-Present attempts.
The observed one-core process CPU range is 0.0668–0.2129%, but there are zero
eligible 60-second intervals: the idle CPU target remains unqualified. API
attempts do not count actual OS wakeups or compositor presentations. Native
window state is observed at transition/interval endpoints, not continuously.

The reused pure validator accepts the unchanged actual trace and rejects seven
headless malformed copies: wrong owned PID, decreased counter, invalid read
bracket, missing transition, incorrect occlusion, duplicate action and missing
terminal result. These are validator controls, not additional product/native
execution evidence. Current SDK activity validation and prior pure payload
functions are bound to the exact reviewed plan; no old producer or source-bound
receipt validator is rerun.

## Eight-minute trace and eligible CPU observations

A separate reviewed 480-second four-state trace completes in 492.374140 seconds,
exits zero, reaches complete log EOF with empty stderr and retires owned process
group 58240. The four observed phase durations total 480.006621 seconds. One
background worker result commits. Its 33 native selector checks each have one
attempt/admission; strict target/cover states and all terminal predicates pass.

| Observed state | Actual quiet interval | One-core process CPU | Primary paint wakes |
| --- | --- | ---: | ---: |
| Visible | 60 seconds, eligible | 0.075553% | 0 |
| Occluded | 60 seconds, eligible | 0.079166% | 0 |
| Minimized | 60 seconds, eligible | 0.079136% | 0 |
| Background worker | 60 seconds, eligible | 0.076492% | 0 |

All four actual monotonic/wall intervals meet the 60-second eligibility rules
and the sampled process CPU target below 0.5% of one logical CPU. All four
versioned activity samples match the actual owned PID and exact counter/read
brackets, with zero empty-event-post and surface-Present attempts. These are
per-interval observations, not five repeated performance runs, actual OS wakeup
counts, compositor presentation or full idle acceptance. The background label
means allocating-worker completion, not OS application background.

The final proof rechecks both clean calling checkouts at `f875857`, each exact
6,260-file stage0079 dependency vector, all 22 controller/helper/predecessor pins,
the unchanged 7,879 SDK payload files plus manifest, exact executable and saved
29,816,117-byte SDK archive. All four owned test groups, including the original
failed attempt, are observed absent. The SDK remains bound to source `abe67c6`;
the later documentation commit does not relabel its provenance.


## Evidence and limitations

All cases have independent fresh result/log/stop scopes and are sealed after
completion. Internal reviews cover the exact plans, raw outcomes and final
checkpoint. Prism reviewed the three documentation files with Anthropic Claude Sonnet 5.5
in one provider call. Its two retained medium findings
prompt explicit open-anomaly and Cocoa-window qualifications here and in both
summary specifications. The original CLI exit one and findings remain preserved;
internal review closes the wording changes without another provider call. Records retain the original failure as well
as successful continuation, complete raw output, owned receipts, exact plans,
validators/executor/lease sources and SDK input hashes. Executables, fonts and
the separately preserved SDK archive are excluded from the evidence tarball.

No input source, VoiceOver, clipboard or preference is changed. Geometry
persistence is disabled locally in the fixture. These checks use real native
windows and GPU readback, but do not establish physical/human interaction, OS
application background, continuous window state/deallocation, actual OS wakeups,
compositor presentation, native/GPU memory, 10-minute composed activity,
1,000 lifecycle cycles, eight-hour soak, older macOS or the required display
variants. Full M2–M7, ordinary performance and Mac acceptance remain open. The
short preflights admit no automatic long campaign; final M4 qualification follows
stabilized functional/performance fixes and its own reviewed plan.
