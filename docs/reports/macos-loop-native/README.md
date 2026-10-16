# Mac event-loop native follow-up and soak diagnostics

Four native queue cases, five composed windows and an eight-minute four-state
idle fixture pass against binaries built from clean `82c0c5e`: 4,479 tracked
files and 6,229 prepared dependency leaves were verified before/after the builds
and native runs. These checks provide native acceptance for the event-driven
Mac queue source checkpoint. They remain distinct from the earlier failed
eight-hour soak and from full Mac completion.

## Completed native checks

| Case | Observed result |
| --- | --- |
| Queue normal | 100 callbacks, 11,000 requested wakes, terminal queue and refused late creation |
| Queue panic | Original panic identity retained, one callback, terminal queue and refused late creation |
| Queue idle | Three delayed callbacks wake indefinite native waits; no deliberate event flood |
| Queue stop | Worker-only stop wakes an indefinite wait without a callback; no deliberate event flood |
| Composed windows | Five windows, 165 completed owner/native actions, 168 attempts/admissions in 58.99 seconds |
| Four-state idle | Four 120-second phases, each with one eligible 60-second CPU interval |

Queue idle/stop also exercise 1,000 ignored late posts after native termination.
Window retry causes are unobserved; queue admission is not callback acceptance.
Each owned controller held the shared native lease through terminal group cleanup
and checked stop markers, exact binary/source identity and bounded logs. The
five-window run retires tracked windows/services/callbacks/AX entries/file panels;
one hidden inactive untitled 500×500 Cocoa window remains. This does not establish
zero Cocoa windows, menu restoration, native/GPU memory budgets or physical input.

| Idle state | One-CPU percentage | Primary paint wakes |
| --- | ---: | ---: |
| Visible, unfocused | 0.06716% | 0 |
| Occluded | 0.02137% | 0 |
| Minimized | 0.03754% | 0 |
| Background allocating worker | 0.04734% | 0 |

The total phase observation is 480.007 seconds. Each percentage is one process
CPU delta across one eligible 60-second interval, including runtime/native
threads. The background phase commits one reserved worker result; it does not
mean the OS application is backgrounded. Window states are sampled at endpoints.
Application-wide wakeups and GPU presents are unavailable. These observations
meet the sampled CPU threshold only; they cannot establish full idle acceptance
or attribute a gain to the event-loop change.

## Earlier overnight failure

The separately frozen older-source eight-hour attempt terminated with exit 1
after 9,084.17 seconds: `soak occluded state drift`, completed 0 of 1 cycles.
It completed the 7,200-second visible phase and six occluded samples, for 30
eligible 60-second intervals overall. Sampled CPU ranged from 0.1333% to 0.5611%
of one CPU, with no primary paint wakes. Minimized/background phases did not run.
The original validator correctly rejects it. Its failure receipt, raw streams,
partial trace, heap profile, producer/source bindings and tracked/outer teardown
remain retained without relabeling it as overnight qualification.

The failed interval's exact endpoint states were not logged by that executable.
A bounded read-only powerd observation shows display assertions labeled
`Codex Computer Use interaction` near the terminal failure; their origin and
causal effect are unknown. This is neither proof of a framework defect nor an
explanation of state drift. No OS input source, VoiceOver, clipboard or preference
changes were made by these test controllers. No global machine inactivity is
claimed.

An internally reviewed policy amendment admits only this exact known failed,
retired predecessor for independent follow-up scheduling, with a fresh group
absence check after acquiring the shared lease. It leaves all overnight/full-idle
qualification flags false. Original success-only harness inputs and failed
control attempts remain preserved. The amended native and measurement harnesses
pass 36 and 17 CPU-only controls; actual held-lease launch refusals also pass.

## Native performance attribution

Two separate optimized fixed-million-row CPU-renderer processes provide 360
profiled samples and 120 nested-source diagnostic samples, each after 128 warmup
steps. Diagnostic UI-work p95/p99 are respectively 7.209/7.864 ms and
7.153/7.637 ms. These are excluded from ordinary warm target qualification.

In the detailed run, each ticket has one scene source. Source raster p95 is
2.348 ms and source copy p95 is 1.163 ms. Summed raster/Draw elapsed is 68.12%;
summed copy/Draw elapsed is 31.86%. These intervals nest; independent percentiles
must not be added. The 10,337,424 logical RGBA bytes and one image API attempt per
ticket do not measure physical GPU transfers, successful uploads or GPU time.

The retained whole-process profile spans 12.11 seconds and 5.92 seconds of samples.
`runtime.cgocall` accounts for 70.61% flat samples, including stacks through native
WaitEvents and accessibility updates. Go profiling cannot resolve the native
work below those calls or attribute elapsed waiting as CPU. Those overlapping
stacks are not a measured frame-stage budget. Keep validation intact and use the
native profile plus source timers to investigate raster/copy and accessibility
snapshot work before choosing an optimization. The earlier 40-process ordinary
P6 campaign still misses the unchanged 4/6 ms targets.

## Failure evidence preservation

The new `vellum.soak.interval-failure` event preserves the captured start/end
target and cover states, query errors, UTC boundaries and separate predicates
before returning query failure or state drift. Queries stay outside quiet CPU
intervals. Text and JSON are bounded, nonfinite numeric diagnostics remain
representable, original errors survive output failures, and successful intervals
emit no failure event. Missing/ambiguous paired windows fail closed. This improves
future diagnosis; it does not fix or identify the cause of the original drift.

Six new test functions cover both endpoint failures, error identity, UTF-8/output
bounds, nonfinite values, UTC conversion, missing timestamps, strict occlusion
and successful no-emission behavior. The candidate's inert native-harness race
tests and vet pass. Native results above predate this diagnostic addition; they
do not qualify its new event in a live failing window run.

Checkpoint acceptance requires raw evidence, source/build vectors, executable
digests, controllers, validators, historical versions, tests, attribution and
internal reviews in `records/evidence.tar.xz`, indexed by
`records/archive-index.json`. Executable binaries remain private.
`records/qualification.json` must identify the exact reviewed endpoint and
provider result. The contract is
[workbench endurance](../../../specs/workbench-endurance.md).

Current-source 1,000-cycle endurance, a successful overnight trace, full P4/P6/P9/
P11, app-wide wake/present and native/GPU memory budgets, physical pointer/display
and human IME/VoiceOver workflows, interoperability and Mac distribution remain
open. Public release, signing and notarization are separate work.
