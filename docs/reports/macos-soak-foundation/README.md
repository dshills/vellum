# Mac lifecycle failure diagnostics and soak fixture

The optional native window-cycle fixture now supports an idle/background soak
and records more context when a native action fails. These sources are guarded
by `darwin && cgo && !offscreen && vellum_native_checks`; ordinary SDK builds
exclude them. This checkpoint qualifies code and unit checks only. No new native
application, soak, desktop interaction or performance campaign was executed.

## Workload and measurement boundaries

`-soak` requires `-composed -cycles=1 -sample-every=1`. Four equal requested
phases arrange a visible unfocused primary window, occlusion behind the survivor,
minimization and visible background-worker completion. A bounded Cocoa census
reports up to 64 process-owned windows with frame, key, visible, miniaturized and
occlusion flags. Native samples must uniquely match the primary title and have
finite positive frame extents. Endpoint observations cannot prove continuous
occlusion, and requested arrangements do not prove their effective state.

Each quiet interval separates process CPU and primary paint-request wake counters
from owner/Cocoa queries, forced GC, memory samples and JSON output. It records
wall and monotonic durations. Short, overslept and clock-divergent intervals stay
ineligible for sixty-second CPU acceptance. Invalid/reversed CPU data is refused.
The process percentage uses one logical CPU and includes both windows and native
threads. The original <0.5% static-idle target also requires app-originated wake
and presentation evidence, which these counters do not supply.

Background samples reserve one task, allocate/checksum 1 MiB with cooperative
cancellation and acknowledge accepted completion on the owner. Worker and owner
termination are joined. The 512 MiB Go-heap/1 GiB RSS guards remain process safety
bounds. These values do not qualify professional native/GPU memory budgets.
See the [fixture contract](../../../specs/workbench-endurance.md).

The fixture logs owner close-policy requests with actor unknown. Failed native
actions include individually synchronized application cancellation, domain window
registration and window-lifetime observations. They form no atomic lifecycle
snapshot, attribute no human/OS/application cause and do not retry an unavailable
attachment or qualify an incomplete action. Forced close bypasses the policy.

## Source and checks

The nine-file work-in-progress was transferred onto committed base
`dc33d64fd11fcacda7b789f10c81001c5b9cba28`. No earlier runtime qualification or
live harness was inherited. This checkpoint additionally rejects a phase without
room for a quiet sample, rejects nonfinite frames, validates interval eligibility,
and exercises the actual allocating checksum worker with a permutation oracle.

Internal review found an unlocked inherited `WinSize()` read racing backend
geometry callbacks. The fixture now copies that value under the desktop window
mutex and unlocks before dispatching native work. A race test uses a native-free
backend record with actual shared-mutex writes/reads. Removing the lock produces
a detected data race; it does not reproduce a real Cocoa callback.

Final v2 checks bind all nine source/test/spec hashes and 6,217 unchanged prepared
files: command unit race PASS (3.766s wrapper/1.747s Go), command vet PASS (0.581s),
native-check fixture compile/link PASS (3.405s), ordinary driver build PASS (0.413s),
portable app/tasks race PASS (1.470s). Four injected faults fail their intended
regressions: phase fit, nonfinite frame, interval eligibility and geometry locking.
These are check durations. The command tests never call its native main loop or
create GLFW/Cocoa windows. V1 receipts precede the locking fix and remain historical.

Before commit, the exact ten source/spec/report files require internal evidence
review and one Prism checkpoint review using Anthropic Claude Sonnet 5.5. Raw
records, source/overlay hashes, historical failure and provider coverage must be
retained in `records/evidence.tar.gz`, indexed by `archive-index.json` and summarized
by `qualification.json`. The compiled binaries are omitted with hash/size records.

## Remaining native gate

The archived `2c87a3d`/6,216-leaf campaign completed 123 of 1,000 cycles, then failed
the 124th window's fourth native action with an unavailable accessibility lookup.
It has 3,201 completed action records plus one failed record; cause was unobserved.
Its original 55-member archive is retained with SHA-256
`3d5fc84797e3b59a8796fccd1dc86b81d27ee3b7b1f9af9a903203e9c0588fd8`.
New diagnostics cannot recover the missing historical cause. This is not a fix or
new pass for that failed campaign.

A fresh source-bound runner and native preflight, complete 1,000-cycle campaign,
actual short and overnight soak, effective-state observations, app-wide wakes/
presents, native/GPU memory and remaining P4/P6/P11/Mac release acceptance are
still required. The original isolated live harness is obsolete for this endpoint.
Live launches remain disabled while the human is online; this checkpoint changes
no input sources, VoiceOver, system preferences or general clipboard.
