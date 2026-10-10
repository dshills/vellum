# Disposable remediation trial

This continues the [foundation fixture](../foundation/README.md), retaining the original as a baseline. Results are in [REMEDIATION_RESULTS.md](../../REMEDIATION_RESULTS.md). These local Cogent patches have not been accepted upstream and are not a production fork or public Vellum API.

## Reproduce

Use the same prerequisites as the foundation fixture: Go 1.27.1, Rust/Cargo, Xcode/macOS SDK, Python 3, `patch` and `install_name_tool`. Put Cargo on PATH. Network access may be needed for locked dependencies.

```sh
./run.sh test
./run.sh build
./run.sh run -report "$PWD/evidence/my-native.json" -exit-after 240s
./run.sh run -idle-sample 60s -exit-after 67s -report "$PWD/evidence/my-demand.json"
./run.sh run -on-demand=false -idle-sample 60s -exit-after 67s \
  -report "$PWD/evidence/my-periodic.json"
./run.sh run -warm-spelling -exit-after 6s -report "$PWD/evidence/my-spelling.json"
./run.sh run -lifecycle -trace -exit-after 18s -report "$PWD/evidence/my-resize.json"
```

`prepare.py` verifies the original hashes, copies the exact pinned Cogent module into ignored `build/upstream`, then applies the three reviewable patches. It never changes the module cache. Re-running prepare replaces that disposable copy, including manual edits there. `remediation.mod` and `remediation.sum` are derived module files; the runner regenerates them. All native builds use `-modfile=remediation.mod`, so they cannot silently fall back to the unpatched toolkit. The bundle has identifier `org.vellum.remediation` and ships the relative AccessKit dylib as in the baseline.

## Patch boundaries

1. **Graphemes and editing state:** existing go-text segmentation supplies logical rune boundaries. Character keys move/delete by grapheme; replacement, word movement, pointer-derived caret positions and selected ranges normalize to boundaries. Raw insertion/completion offsets keep rune semantics. `EditingState` is a read-only UI-thread snapshot with logical caret/selection and current scene-pixel caret bounds. `SetEditingSelection` validates cluster boundaries. Pixel snapping uses a preceding logical boundary; precise nearest visual caret choice, bidi navigation and multiline editing are not established.
2. **Lazy spelling:** editor construction no longer loads the spelling dictionary. The first potential spelling operation or explicit line check initializes it. `-warm-spelling` explicitly requests a line check, providing the memory control. It defers optional work; it does not shrink the dictionary when used.
3. **Requested painting:** existing toolkit behavior remains the default. This fixture opts into `SetOnDemand(true)`. Scene invalidation, deferred callbacks and native/input events request paints; requests coalesce. The earliest explicit deadline handles caret sprites, resize settling and an in-flight render retry. Shutdown cancels its timer. This is a feasibility patch: arbitrary animation, popup/sprite mutation paths, global style changes, resource loss, cross-platform and long stress coverage remain open; no general frame-pacing promise is made.

## Native text checks

Inspect the real macOS AX tree. The field has an AccessKit TextRun with UTF-8 byte lengths per selectable grapheme; native UTF-16 ranges map to validated character/rune boundaries. The adapter rejects malformed, NUL-containing or oversized (>255 UTF-8 bytes per cluster) representations instead of truncating them. Multiline ranges and text-run geometry remain absent. The caret snapshot is not wired to the Cocoa IME client.

Set field value to `é`, focus the single-line field, and inspect fresh state. Use the native selected-text-range action to place an **unselected** caret after the sequence. Record it using the native **Record text** action: expect cursor/anchor/focus `2/2/2`. One actual Backspace must leave empty text. Repeat `👩‍💻` at `3/3/3`, and `AB` at `2/2/2` (expect `A`). Undo should restore the deleted combining grapheme.

For `A👩‍💻éZ`, native selection of the emoji should report rune anchor/focus `1/4`; deletion must leave `AéZ`. From the end, successive Left keys reach offsets `6`, then `4`; Shift-Left selects the emoji with anchor/focus `4/1`. Record after each key with a fresh native observation. Forward Delete from the unselected offset `1` must remove the whole emoji.

Inspection and recording use `EditingState().Text`. Cogent's ordinary `Text()` method commits edits and clears selection, so using it in a semantic getter corrupts editing behavior. It remains appropriate only for an intentional commit. Every UI action should be followed by a fresh native state observation before dependent actions.

## Measurements and remaining gates

Idle samples start after three seconds settling. Activate AX before the interval, then leave the initial non-text focus and static fixture unchanged. CPU is process Getrusage time; `PaintSchedulerWakes` counts the patched window loop's tick/request dispatches, not all OS/runtime wakeups. Canvas callbacks are not compositor presents. The static fixture retains the same multilingual corpus and system font fallback.

The report includes forced-GC HeapAlloc, current resident size, spelling initialization, semantic/action counters and editing snapshots; it writes a sampled heap profile. The font cache is warm/uncontrolled. These are diagnostic intervals, not the full five-run acceptance workload. Real IME composition, VoiceOver completion, text geometry, before-show native attachment, bounded action transport, Windows/Linux and clean distribution remain open.
