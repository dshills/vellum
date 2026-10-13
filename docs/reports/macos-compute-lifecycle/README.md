# Mac compute ownership

ComputeSystem owned a separately requested GPU device but did not release it at
shutdown. A refused command Finish also left its command reference live. Patch50
now detaches owner state, abandons pending recording, retires child resources and
releases the private device and independently owned queue. The borrowed adapter
remains live. Empty/repeated shutdown and absent-pass submission are safe.

Checked construction and pass admission publish complete live state. Submission
uses the same ownership helper as graphics: pass before Finish, buffer before
command, cleanup on success or refusal, and old bind-group retirement on both
paths. Caller End and serialized ownership remain required. The pinned native
BeginComputePass has no Try API and can panic; acquired-command cleanup unwinds
without converting that native panic to an error or promising broad recovery.

## Verification

Two original-compatible offscreen tests fail three assertions in the original
compute/device files: empty End and Release panic, and a missing device retains
its queue field. The repaired code passes nine counted admission scenarios,
pass-factory panic unwind and existing submission failure kernels. Final GPU and
gpudraw suites pass three race repetitions. The complete final offscreen adapter
race suite, eleven preparation tests and native-window consumer compilation pass.
Vet passed before the fixture-only teardown correction; production sources are
byte-identical afterward. The first adapter run failed only because the sandbox
refused a temporary Unix socket; an authorized final retry passes. A native build
attempt without the required fixture tag fails before compilation; it is retained
as a failed attempt, with no executable/native acceptance claim.

The final real headless GPU fixture runs eight four-integer shader dispatches and
checks all 32 output values through storage readback. Eighteen recording wrappers,
five child references and the private device/queue are invalid after retirement.
An active recording is abandoned and both borrowed adapter/device remain live.
A separate actual partial-device case retires its still-live queue. Held Go
wrappers establish explicit reference retirement. Original compute.go overlaid
on the then-current prepared dependency fails private-owner retirement with the
same 32-word shader workload in the preceding fixture version. Current device
cleanup and test helpers remain in that baseline; it is not the final fixture
version/prepared map or a complete parent-checkout execution. Earlier empty-pass
headless observations remain separately bound historical checks.

## Review and evidence

Internal review checks all eight selected repository leaves, the static pin maps,
patch replay and six decoded outputs. Prism ran once with Anthropic Claude Sonnet
5.5, six-file coverage, two chunks, 16,176 input/5,073 output tokens and raw exit 1.
Its single medium finding flags hash files absent from the selected provider
payload. Both files are included in the full staged checkpoint; internal replay
and eleven preparation tests verify them. The finding is resolved without a
second provider call. Internal review then corrected fixture-only failure teardown
order; the final patch/hash registration and native/race/build checks bind that
delta. All decoded production sources remain identical to the Prism input.

[Checkpoint](records/checkpoint.json) binds the source/evidence reviews and the
[archive](records/evidence.tar.gz) with its [member index](records/evidence-index.json).
The archive includes failed attempts and separate source versions, raw receipts,
logs and producers. Binaries/caches are excluded. All6191 prepared file identities
are bound; this is not a complete compiler/SDK closure. No desktop windows,
surfaces, input, settings changes or presentation tests ran. Physical GPU faults,
native/GPU heap reclamation, sustained performance and Mac completion remain open.
Vars.Config temporary binding layouts and storage-value staging/map integrity
remain follow-up audit items; this checkpoint does not qualify those paths.
