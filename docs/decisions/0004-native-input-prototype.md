# ADR 0004: native composition is implemented; retain the foundation hold

Date: 9 October 2026
Status: Provisional; local single-field implementation demonstrated; G1 remains open

## Context

The [remediation decision](0003-remediation-feasibility.md) prioritized native input. A separate Cocoa client now delivers marked text, native replacement, commit/cancel and current-caret coordinates to a Go TextField. Real U.S. dead-key input, Undo/Redo and native emoji selection replacement are verified. Live Japanese Romaji and Chinese Pinyin conversion, composition Undo/Redo and cancellation work. Candidate cycling after horizontal scrolling works; separate candidate-panel placement remains unverified. Deterministic protocol calls also cover explicit replacement and burst updates. See [results](../reports/NATIVE_INPUT_RESULTS.md).

The client subclasses only the exported native view and retains the original path for other input. Patch 0004 adds field transactions and visible-caret/layout notifications. The native class is restored before AccessKit/view destruction. UTF-16 validation and sequence acknowledgements protect range integrity and avoid stale native text. The first live cancellation test found a fall-through bug; the corrected test preserves preceding edits.

## Decision

Keep D as the experimental candidate, A as control and B as fallback. Complete visual candidate-panel placement and actual VoiceOver navigation/speech workflows before workbench expansion. With explicit user approval, Romaji/Pinyin and VoiceOver were temporarily enabled and original settings restored. The runner does not change macOS preferences. Value/selection actions and typing work with VoiceOver active, but automated VoiceOver navigation/invoke keys yielded no verifiable result; no screen-reader completion is claimed.

Seek a supported toolkit native-input and before-show/lifetime seam, including upstream acceptance of necessary field changes. The current Objective-C interposition is a disposable feasibility technique with backend/class-order assumptions, not a production support commitment. If providing those hooks requires major platform/text replacement or enduring unsupported ownership, compare B on the same requirements.

## Consequences

Marked text now crosses the native/Go boundary and the representative field has composition transaction behavior. Current-caret coordinates are asynchronous with rendered layout and limited to visible single-field insertion. Multiline/bidi, arbitrary native ranges, real candidate-window and assistive-technology completion, bounded transport, other platforms, distribution and memory/performance gates remain open. No upstream issue/PR was posted and G1 is not accepted.
