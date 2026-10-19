# AppKit hidden-window reference

A standard AppKit text client created one persistent hidden `TUINSWindow` without
linking Vellum or GLFW. The fixture's own windows and views expired, while that
additional window remained after two fixed lifetime cycles. This supports an
AppKit text-service lifetime candidate for the historical hidden-window anomaly;
it does not identify the historical `NSKVONotifying_TUINSWindow`, prove Vellum
native deallocation, or close M4.

The reference ran against the guarded source101 baseline
`8ba9ca617d5c01d40fd9785a2896b8153bdc4772`. It is separate from Vellum's native
window-cycle acceptance and the later Core102 integration. Both the refused
first attempt and the successful revised protocol remain sealed. Compact hashes
and observations are in [evidence.json](evidence.json); raw compiler/runtime logs,
admissions and the independent review remain in the development checkpoint
`macos-m4-appkit-reference`, including its `v2` directory.

## Attempts and observations

Version 1 treated a false `setActivationPolicy:` return as refusal and exited
with status 1 before its first inventory or explicit fixture window cycle. It
produced no stdout observations. Its source and failure evidence remain intact;
there is no retrospective pass or inferred initial policy state for that run.

Version 2 reads the actual policy first, calls the setter only if necessary and
requires prohibited-policy readback. In both actual processes, every inventory
reported `Before=Desired=Readback=2`, `SetterAttempted=false` and
`SetterReturned=null`: policy was already prohibited and no setter was attempted.

The two fresh-process modes each completed 15 records: an initial inventory, six
inventories per cycle, a terminal inventory and one protocol result. The windows
were hidden and nonkey. Plain mode uses `NSView`; text mode uses `NSTextView` with
owned synthetic text, a local first responder, an input-context getter and local
insert/mark/unmark/discard operations. No input context/source activation,
physical events, personal app data, preferences, clipboard or VoiceOver change
is part of this reference.

| Observation | Plain mode | Text mode |
|---|---|---|
| Initial windows | 0 | 0 |
| Maximum sampled windows | 1 | 2 |
| Owned window weak reference | Expired after each pool drain | Expired after each pool drain |
| Owned view weak reference | Expired after each pool drain | Still live at pool/100 ms; expired by 500 ms, both cycles |
| Terminal windows | 0 | One persistent `TUINSWindow` |

The text-service window first appeared after the first local text-client actions,
not in the initialized baseline or first `created-hidden` inventory. Its same
process-local window number persisted through cycle 2 and terminal sampling. It
was untitled, hidden, nonkey, nonminiaturized, nonopaque, alpha 1, level 0, frame
`[0,0,500,500]` and ordered index -1, with `NSView` content. Neither fixture-owned
weak marker identified it as the fixture's own window or view.

Runtime metadata identifies its class image as
`TextInputUIMacHelper.framework/Versions/A/TextInputUIMacHelper`, bundle identifier
`com.apple.TextInputUIMac`, and superclass chain
`TUINSWindow → NSWindow → NSResponder → NSObject`. No KVO wrapper appeared in this
reference. A class family, code-image origin or matching frame does not establish
cross-process object identity or managed ownership of the historical Vellum
window. All sampled windows and classes remain counted; there is no whitelist or
count subtraction.

## Protocol, provenance and checks

The source fixtures are archived byte-for-byte for review:

- [Version 1](fixtures/appkit_reference_v1.m), including the refused policy check.
- [Version 2](fixtures/appkit_reference_v2.m), including policy readback and the
  complete two-mode lifetime protocol.
- [Copied production serializer](fixtures/native_window_snapshot.h), identical to
  `drivers/cogent/native_window_snapshot.h` at the guarded baseline.

The `.m` fixtures are Vellum-authored code with their MIT SPDX notices. The
version 1 refusal receipt and raw stdout/stderr are hash-bound in `evidence.json`. The
serializer is copied from this repository and covered by the root
[MIT license](../../../LICENSE); its original bytes are preserved. Apple SDK
headers and binaries are not copied into this checkpoint. SDK/clang identities,
runtime image metadata and raw-record hashes are evidence, not redistribution or
rights approval.

The actual executable was built with Apple clang 21, ARC and the selected Xcode
macOS SDK. Independent Mach-O inspection found only AppKit, Foundation,
CoreFoundation, libobjc and libSystem dynamic dependencies. There is no Vellum,
GLFW or WGPU dependency. Snapshots execute on native main, use the production DTO
serializer and record every process window and the complete bounded superclass
chains. Observer autorelease pools retain no window/view references beyond a
sample. The two drain intervals are fixed at 100 ms and then another 400 ms;
they do not poll until a desired object disappears.

Limits are two cycles per process, 64 windows, 64 runtime classes, 16 superclass
levels, 512 KiB per JSON record, 8 MiB per output stream and a 30-second native
process watchdog. Actual stderr was empty. The shared live lease, owned process
retirement and sandbox guards remained in force. All six build/probe/native
command groups retired and were observed absent. The root wrapper recorded
successful source/SDK and geometry comparisons before/after the native cases.
It did not serialize the run-local geometry dictionaries; the independent audit
also verified clean canonical source101, its frozen SDK and the matching current
geometry baseline. Later T integration is separate from those runtime guards.

The independent audit binds the actual sources, compiler, executable, 15 records
per mode, class metadata, weak-reference timeline and cleanup receipts. This
repository checkpoint packages that evidence; it performs no new build or
native run. Protocol completion and fixture weak expiry do not certify a native
leak test, endurance, GPU memory, Vellum lifetime correctness or full Mac/M4
acceptance. A separately bound fresh framework comparison remains necessary to
identify and evaluate the historical anomaly.

Prism reviewed the complete shared 23-file, 95900-byte packet with Anthropic Claude Sonnet 5.5 in one call/chunk, without exclusion or truncation. One retained documentation warning is independently dispositioned as non-actionable; its own message acknowledges no defect, and the serial/paced figures and gate labels match the native records. No finding was discarded by the provider. This checkpoint and the separate Metal policy experiment are committed independently. Raw native logs were locally audited and bound by hashes; they were not wholly sent in the provider packet.
