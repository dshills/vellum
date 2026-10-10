# Single-field Cocoa input prototype

A separate continuation of the [remediation fixture](../remediation/README.md). Results and limits are recorded in [NATIVE_INPUT_RESULTS.md](../../docs/reports/NATIVE_INPUT_RESULTS.md). The previous prototypes are unchanged. This is a local macOS arm64 experiment, not a public Vellum API or a production backend.

## Build and run

Prerequisites match the remediation fixture: Go 1.27.1, Cargo/Rust, Xcode/macOS SDK, Python 3 and `patch`. Put Cargo on PATH. Locked dependencies may need network access.

```sh
./run.sh test
./run.sh build
./run.sh run -report "$PWD/evidence/my-live.json" -exit-after 240s
./run.sh run -input-probe -lifecycle -exit-after 20s \
  -report "$PWD/evidence/my-protocol.json"
./run.sh run -native-input=false -exit-after 10s
```

The runner creates `nativeinput.mod` and verifies pinned module hashes before applying patches 0001–0004 to ignored `build/upstream`. Patch 0004 applies **after** the retained remediation patches; it adds single-field composition transactions, a layout/render notification, and visible current-caret bounds. Generated module files and the copied module are disposable. The module cache is never patched. The native bundle identifier is `org.vellum.nativeinput`.

## Native ownership and routing

`inputbridge` attaches to the exported GLFW Cocoa view after AccessKit has attached. A reusable Objective-C subclass is applied only to that view; there is no global method swizzling or GLFW source change. State is held in an associated native object. All attachment, updates, protocol methods and teardown run on Cocoa's main thread. Native callbacks copy strings/ranges into owned Go values and enqueue unique events; they never read widgets. Go's UI path validates UTF-16 scalar bounds before applying edits.

When the representative field is active, printable keys go through `interpretKeyEvents` and native insertion. Application shortcuts and unhandled commands forward once to the original GLFW path. Other widgets retain that path. Escape during composition cancels before Cocoa can commit a dead-key accent and before Cogent can revert unrelated earlier edits. Marked updates form one undo transaction; unmark commits it; cancellation restores the original text and selection. An authoritative value replacement clears composition.

The native client keeps optimistic text/selection across callbacks. A monotonic sequence and UI acknowledgement prevent older publications from overwriting newer native input. Geometry comes from the latest Go layout/render and invalidates Cocoa's cached character coordinates; move/resize/screen-change notifications also invalidate them. The visible current insertion caret is converted from physical scene pixels to view points, then through AppKit to screen coordinates. Its actual range is zero-length, with zero width and the rendered caret height. Empty visible bounds report an unavailable rectangle. This does not implement arbitrary text-range geometry or point hit testing.

Input is detached before AccessKit and before view destruction. Teardown restores the view's original class, removes associations/observers and deletes the Go handle. Unexpected subclass order is asserted rather than silently restoring an unrelated class. Attachment remains after the first show/focus: a supported before-show hook, other native subclass combinations and sustained lifetime stress are still required.

## Verify

With the existing U.S. input source: set Component value to `A`, place an unselected native caret after it, then send Option-E and E without switching applications between keys. Expect `Aé`. Native **Record text** saves an owned snapshot and current-caret protocol query without committing the field. Command-Z returns `A`; Command-Shift-Z restores `Aé`. Option-E followed by Escape must preserve `Aé`. Select `👩‍💻` in `A👩‍💻Z`, type `b`, and expect `AbZ`; Undo restores the sequence.

The `-input-probe` harness calls the actual client's protocol methods with Japanese/Chinese strings, explicit replacement ranges, cancellation, an invalid surrogate-interior range and a burst of three callbacks. It is a deterministic integration test, **not live Japanese/Chinese IME input**. Its JSON includes native selection, marked range, substring and current-caret coordinates. `-lifecycle` records six public resize requests and caret queries. Record text also allows checking a long field's horizontal scrolling.

For a live Japanese/Chinese test, use configured Romaji/Pinyin sources, verify provisional text, select candidates, commit/cancel, and check the candidate window after scrolling/resizing. The recorded live run passes Romaji/Pinyin commit, Undo/Redo and cancellation, including Japanese candidate cycling after horizontal scrolling. Separate candidate-panel placement is unverified. Report input environment/source fields are captured at first show; each recorded native probe includes the current source. VoiceOver must be exercised as an actual workflow. The runner never enables input sources, toggles VoiceOver or changes system preferences.

## Remaining scope

Single field, wrapping disabled, logical selection only. Marked text has a distinguishing underline; clause-specific attributes and visual bidi carets are not implemented. Multiline text, arbitrary AX range geometry, native range discovery, screen-reader completion, target-platform coverage and long queue/lifetime stress remain open. Action delivery remains lossless but unbounded. The native subclass and four local toolkit patches are unaccepted extension seams; upstream-supported integration is required before committing to reuse. Existing font-memory and formal performance gates remain open.

For settings changes during manual verification, record and restore the original input sources and VoiceOver state. macOS can also change Globe-key behavior and add Dictation languages automatically; verify those after removing temporary sources. Use `-exit-after 0` only for supervised runs and close with **Finish**.
