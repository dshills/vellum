# Bounded native input and owned desktop lifetime

This follow-up uses pinned Cogent plus patches 0001–0009 and the separate deferred bitmap/static-outline go-text patches from `../fontresources`. Earlier fixtures remain historical controls. `../foundationcontracts/mailbox` gains one non-consuming `Ready` accessor.

`inputbridge` computes prospective snapshots and reserves transport credit before native text, selection, composition history or sequence changes. A full or oversized queue causes NSBeep and leaves those states unchanged. Successful mutations commit the reserved snapshot synchronously on Cocoa's main thread. Allocation failure is process-fatal; this prototype does not recover from Objective-C exceptions or Go out-of-memory. Native attachment, mutation and detachment are serialized on the Cocoa thread. Accepted messages drain after detach; closing a window discards its native composition context while preserving already-admitted updates for application handling.

The fixed ring limits 256 ordered messages and 4 MiB of queue-owned payload, with 64 KiB/message. Tickets transiently own encoded copies in addition to queue reservations; serialization/decoded values/native strings/application history and the toolkit's unrelated events are outside that payload budget. The pump enqueues at most one wake token, drains at most 16 actions per UI turn and rearms if deliverable work remains. It holds payload credit through processing. A separate one-slot latest channel coalesces caret state. Other action refusals are logged; refused actions never reach domain handlers. The event deque does not receive native/action payloads. The fixture retains at most 128 full native input snapshots and 2 MiB of text/kind bytes, reports evictions, and releases discarded values. Other bounded-duration diagnostic histories are not a production retention policy.

Patch 0008 exposes a desktop `App.NativeAttach` extension, initialized before any windows are created. It attaches after GPU creation but before `Show`, supplies main-thread focus notifications and closes the attachment before application/GPU/view cleanup. Hook methods cannot reenter UI or call RunOnMain. Creation failures close partial attachments and release surface/drawer/view resources without registering a window. `StartHidden` constructs desktop GPU windows for lifecycle testing; other drivers do not implement it. Close is idempotent.

Patch 0009 replaces deferred device release with owner-serialized polling and explicit queue/device release. Queue.GetQueue carries independent references. The original 1,000-cycle run grew from 69 to 458 MB RSS; corrected runs are recorded separately. Hardware/driver-private allocations remain outside explicit byte accounting.

Build on macOS with `./run.sh`. Defaults use private temporary cache/build directories; callers may explicitly reuse trusted caches. Cargo build needs network on first use. Run:

```sh
./build/integration-fixture -input-probe -saturation -exit-after=14s -report=evidence/native-lifetime.json > evidence/native-lifetime.log 2>&1
./build/lifecyclecheck -cycles=1000 > evidence/lifetime-release-1000.json 2> evidence/lifetime-release-1000.log
./build/lifecyclecheck -cycles=5000 > evidence/lifetime-release-5000.json 2> evidence/lifetime-release-5000.log
python3 verify.py
```

The fixture corrects baseline Prism findings concerning Intel NSRect dispatch, NSTextInputClient protocol registration, signed range arithmetic, UTF-8 scalar boundaries, finite semantic geometry, indexed traversal, accessible edit/commit parity and timer counter ownership. Historic files and their original evidence are preserved.

These are deterministic native protocol and real hidden Cogent/Metal resource tests. They do not validate live candidate-panel placement, complete VoiceOver workflows, Windows/Linux adapters, full visual bidi navigation or trusted distribution. G1 remains open.

The final native build includes deferred font resources. Its measured full-app memory is recorded separately from earlier unpatched-font native runs and from the isolated shaper corpus.

The separate caret slot coalesces publication triggers; the UI owner then publishes current text, selection, activation and geometry together. Stale acknowledgement rejects that complete publication. Backward anchor/focus direction is preserved through cancellation. External edits invalidate an obsolete whole-document cancellation baseline rather than permitting Escape to erase newer application edits.

Toolkit-generated pointer/key/shortcut events remain outside this mailbox. Ordering/fairness between bounded batches and forwarded legacy shortcuts still requires integration tests; the demonstrated lossless contract covers admitted snapshot/action messages, not every toolkit event path. The adapters and hooks remain proposed local extensions pending upstream acceptance.

## Remaining acceptance audit (10 October)

After `./run.sh` prepares the owned dependencies, run `./check_foundation.sh [output-directory]` on macOS. It builds a separate offscreen probe and the native lifecycle probe, checks vet/race tests, and validates their evidence. It deliberately exits 1 because this bounded audit cannot establish G1; exit 2 from the audit CLI identifies a harness error. It neither changes input sources nor starts VoiceOver. The CLI's exit 0/1 concerns only the measured ordering/buffer assertions, while its JSON separately records the overall gate and unavailable checks.

The ordering probe uses the toolkit's actual Source/Deque and the fixture pump's placement, inserting 32 native records before a legacy shortcut. It proves native FIFO delivery but reproduces the shortcut overtaking the second batch. The long-text probe checks exact buffers and logical grapheme primitives, not interactive layout. Bidi uses the field's actual handler and shaped boxes; it does not certify visual caret/selection behavior. The font probe queries 64 sorted family names for one Latin glyph and reports unique resolved faces; it is neither 64 distinct faces nor a global memory bound. Native queries directly exercise Cocoa methods in hidden views, not candidate-panel placement.

See [the current acceptance report](../../FOUNDATION_ACCEPTANCE_AUDIT.md) and `evidence/20261010/` for the measured outcomes, broader regression failures, live-test limitations and restored settings. Original evidence remains intact.
