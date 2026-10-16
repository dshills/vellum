# Mac native runtime and SDK checkpoints

The completed 16 October 2026 checkpoints verify same-name window replacement,
actual GPU fixtures, sampled quiet CPU and private SDK relocation. The sources
are recorded separately: lifecycle source `ddd23c6` has 4,566 tracked and 6,230
prepared files; native diagnostics and the SDK use `97aee13`, with 4,585 tracked
and 6,243 prepared files. These results advance Mac acceptance while ordinary
frame latency, broader memory budgets, physical input, human accessibility and
distribution acceptance remain open.

[Qualification](records/qualification.json) binds the results and archive.
[The index](records/archive-index.json) maps each saved artifact to a digest and
its member in [the evidence bundle](records/evidence.tar.xz).

## Same name window lifecycle

Source `ddd23c6ea06ef7baf1014a08c61d01601c059f36` completes 1,000 composed
windows, 33,000 acknowledged native/owner actions and 999 sampled same-name
replacements in 11,279.409 seconds. The fixture checks the version-2 mix of
controls, editor history, virtual grid/tree publication, modal actions, color
controls, ListBox rejection/repetition and named Canvas2D selection. Stale owner
capabilities are checked against replacement identities. These are Cocoa selector
and owner-completion checks; physical pointer behavior is separate.

There are 33,098 queue attempts and admissions for 33,000 completed action
records. The 98 additional admissions have no observed cause and are not extra
accepted callbacks. Forced-GC samples range from 30,154,216 to 88,456,608 bytes of
Go heap and 185,368,576 to 495,763,456 bytes of RSS. They pass the fixture's
512 MiB heap and 1 GiB RSS safety guards, which do not qualify the different
professional workload's memory budgets or attribute native/GPU memory.

The final main-thread snapshot retires tracked native operations, accessibility
windows, panels, callbacks and owned menu/tray targets. The owned process is
reaped and its group is observed absent. One hidden, inactive, untitled
`NSKVONotifying_TUINSWindow` with an `NSView` remains in the pre-exit snapshot;
its owner, cause and deallocation are unproven. `Menu.OriginalCaptured=false`
leaves original-menu restoration unobserved.

## Native GPU recovery and ownership

Five actual GPU scene tests and the separate native composer-ticket test pass on
source `97aee13132b41b35b7c041da7426b22948263f7a`. Coverage includes the integer
compositor, shared-drawer ordering/fallback, real text/paths, covered uniforms,
and nested timing tickets without double counting. A separate two-window
same-name check completes all 66 native/owner actions with recorded cleanup.

The input test reaches actual retained buffer headroom, five checked writes and
logical-length accounting. It deliberately attempts a checked write to a buffer missing `CopyDst`,
observes the checked native refusal and early return, restores the owned buffer,
and verifies recovery/readback without device loss. The earlier test stopped
before this branch because its encoded inputs exactly filled power-of-two
buffers. The new fixture encodes three ordered operations and adds a CPU
headroom regression. Exactly one prepared test file changes; production renderer
bytes are unchanged by this fix. The original failed native test and the queue
that stopped on it remain failed in the retained evidence.

## Quiet CPU and framework API attempts

The short four-state trace completes but has no eligible 60-second intervals.
The separate eight-minute trace completes one eligible 60-second quiet interval
per state, with the following process CPU percentages of one CPU:

| Fixture phase | CPU percent |
| --- | ---: |
| Visible | 0.0767 |
| Occluded | 0.0750 |
| Minimized | 0.0736 |
| Allocating-worker background phase | 0.0786 |

All four samples meet the CPU component's unchanged <0.5% target. Each records
zero primary paint wakes, empty-event-post attempts and surface-present attempts;
the background phase observes one worker commit. Native counters are bracketed
and bound to the owned process. They count framework API attempts, not actual OS
wakeups or display presentations. Window states are checked at interval
endpoints; the background phase does not certify OS application-background
behavior. This trace does not complete full idle or overnight qualification.

## Private SDK relocation

The `97aee13` private SDK package completes in 4.751 seconds and relocation in
46.782 seconds. All 7,381 manifest files are verified, replacements remain within
the relocated SDK, and generated standalone-consumer/public-API race, vet and
build checks pass in a path containing spaces. The archive SHA-256 is
`fce1e78ffe3b99aeb3745d6cc682ef1315ad9a446d3cbc6088f24277b037dbdb`.

Benchmark fixtures now disable process-local window geometry persistence before
application setup; normal return restores the prior process-local value after
application closure. The SDK checker requires exactly one native benchmark
report with the actual isolation flag set to true. This relocation invocation
omits `--runtime`, so it establishes compilation and consumption, not execution
of the relocated native application or public release authenticity.

## Native dependency provenance

The saved [wgpu-native v29.0.1.1 release](https://github.com/gfx-rs/wgpu-native/releases/tag/v29.0.1.1)
asset has the recorded release digest, and its Darwin arm64 static archive is
byte-identical to the pinned local dependency. The release's commit asset names
`6aed50955d934ac36049ba8d002034841633ae02`; saved tree/blob records bind seven
source declarations at that commit. The original download timeout is retained
separately from the successful retry. Release-byte equality and source
declarations do not establish a reproduced build, the executed compiler,
selected compiled features, complete component closure or redistribution rights.

## Evidence and remaining acceptance

Internal reviews approve each completed scope without unresolved findings. The
bundle contains raw traces, saved producers, source/build maps, owned-process
receipts, pure validation outputs, reviews and historical failure records. It
also includes the digest-checked SDK manifest. Private executables, the SDK tar
archive and downloaded native ZIP are excluded; their receipts/digests and local
cache locations are retained in the index. Saved producers retain their original
absolute paths, so this is an evidence archive rather than a portable replay.

The unique-name 1,000-window run and later grid diagnostic campaign are excluded
from this checkpoint. The failed eight-hour soak remains failed. Ordinary
4 ms p95 / 6 ms p99 frame targets, sustained pacing/input, full P4/P6/P9/P11,
professional and native/GPU memory, actual wake/present behavior, physical
pointer/displays, human IME/VoiceOver, broader service interoperability and Mac
distribution acceptance still require their own evidence. See
[Mac completion](../../../specs/macos-completion.md).
