# Bounded native input and owned desktop lifetime

This follow-up uses pinned Cogent plus patches 0001–0012 and the separate deferred bitmap/static-outline go-text patches from `../fontresources`. Earlier fixtures remain historical controls. `../foundationcontracts/mailbox` gains one non-consuming `Ready` accessor.

`inputbridge` computes prospective snapshots and reserves transport credit before native text, selection, composition history or sequence changes. A full or oversized queue causes NSBeep and leaves those states unchanged. Successful mutations commit the reserved snapshot synchronously on Cocoa's main thread. Allocation failure is process-fatal; this prototype does not recover from Objective-C exceptions or Go out-of-memory. Native attachment, mutation and detachment are serialized on the Cocoa thread. Accepted messages drain after detach; closing a window discards its native composition context while preserving already-admitted updates for application handling.

The fixed ring limits 256 ordered messages and 4 MiB of queue-owned payload, with 64 KiB/message. Tickets transiently own encoded copies in addition to queue reservations; serialization/decoded values/native strings/application history and the toolkit's unrelated events are outside that payload budget. The integrated `OrderedPump` reserves one payload-free deferred FIFO node per ordered admission. The deque merges only adjacent ready nodes from the same owner, up to 16 actions per UI turn; legacy events separate groups. Payload credit remains held through processing. Cancellation removes metadata nodes, and close removes this owner's queued nodes before the quiescent-owner bounded shutdown drain. The original single-wake pump remains a historical control. A separate one-slot latest channel coalesces caret state. Other action refusals are logged; refused actions never reach domain handlers. The event deque does not receive native/action payloads. The fixture retains at most 128 full native input snapshots and 2 MiB of text/prior-text/kind bytes, reports evictions, and releases discarded values. Other bounded-duration diagnostic histories are not a production retention policy.

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

Toolkit-generated pointer/key/shortcut events remain outside this mailbox. The measured cross-path ordering failure is fixed by deferred FIFO admission, with tests for earlier/later native records around shortcuts. General scheduling fairness/latency and native shadow-state synchronization remain unverified. The adapters and hooks remain proposed local extensions pending upstream acceptance.

## Remaining acceptance audit (10 October)

After `./run.sh` prepares the owned dependencies, run `./check_foundation.sh [output-directory]` on macOS. It builds a separate offscreen probe and the native lifecycle probe, checks vet/race tests, and validates their evidence. It deliberately exits 1 because this bounded audit cannot establish G1; exit 2 from the audit CLI identifies a harness error. It neither changes input sources nor starts VoiceOver. The CLI's exit 0/1 concerns the local ordering/buffer assertions (and a failed visual handler raises a harness error), while its JSON separately records the overall gate and unavailable checks.

The ordering probe uses the toolkit's actual Source/Deque and the fixture pump's placement, inserting 32 native records before a legacy shortcut. It now observes all 32 earlier updates before the shortcut, across two batches of 16; an explicit legacy control still reproduces the earlier failure. The long-text probe checks exact buffers and logical grapheme primitives, not interactive layout. Bidi now lays out and renders the field before checking its strong-caret arrow policy. Native geometry covers visible contiguous grapheme fragments, insertion stops and screen-point hits, including wrapped lines. The font probe queries 64 sorted family names for one Latin glyph and reports unique resolved faces; it is neither 64 distinct faces nor a global memory bound. Native queries directly exercise Cocoa methods in hidden views, not candidate-panel placement.

See [the current acceptance report](../../docs/reports/FOUNDATION_ACCEPTANCE_AUDIT.md) and `evidence/20261010/` for the measured outcomes, broader regression failures, live-test limitations and restored settings. Original evidence remains intact.

The [input-ordering fix report](../../docs/reports/INPUT_ORDERING_FIX.md) and `evidence/20261010-ordering/` supersede the initial ordering failure while preserving its evidence. `Stats.Wakes` for the ordered path counts created metadata tokens; `native_batch_events` records merged audit dispatches. G1 remains open.

The [visual text fix report](../../docs/reports/VISUAL_TEXT_GEOMETRY_FIX.md) and `evidence/20261010-visual-text/` supersede the initial bidi/geometry defects for the native macOS TextField. Run `./build/integration-fixture -geometry-probe -exit-after=5s -report=build/rendered-native.json` after preparation/build to compare native rectangles and point hits with independently converted rendered cells. The multiline code editor adapter is now [implemented and tested](../../docs/reports/MULTILINE_NATIVE_ADAPTER.md); full editor visual navigation and broader G1 gates remain open.

## Multiline native adapter

Use `-editor-probe -geometry-probe -exit-after=5s` for rendered Editor range/insertion queries, or `-editor-probe -input-probe -saturation -exit-after=14s` for deterministic composition and queue checks. Stable client identity and prior-text comparison route pending updates safely across focus changes. Native whole-document context is capped at 1 MiB; both text snapshots count toward the 64 KiB encoded-message admission limit. Editor keyboard/pointer navigation retains its fixed-grid policy. See the [adapter report](../../docs/reports/MULTILINE_NATIVE_ADAPTER.md), ADR 0011 and `evidence/20261010-editor-native/`.
