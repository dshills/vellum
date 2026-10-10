# Multiline editor native adapter

10 October 2026. The experimental macOS adapter now supports the exact multiline Editor as well as TextField. G1 remains open and broad workbench expansion remains on hold. [ADR 0011](../decisions/0011-multiline-native-adapter.md) records the extension.

Patch 0012 adds an atomic compare-and-edit seam to the owned Cogent buffer. Marked text is visible to every view, while commit creates one net undo transaction. Cancellation restores original text, directional selection and dirty flags. Ordinary shared-buffer edits, undo/redo, read-only transitions, owner removal and explicit saves seal pending composition first; buffer close discards its baseline. Delayed autosaves snapshot committed text rather than provisional preedit. No-net-change commits restore original dirty flags and add no undo record. Existing external-file modification checks remain in use.

One Cocoa client is rebound between controls using stable IDs (8 for TextField, 40 for Editor). Every admitted snapshot carries its origin client and prior text. Queued updates reach their original control; stale snapshots cannot erase newer shared edits. Focus handoff commits the old marked context, waits for acknowledgement, then binds the next control. Admission refusal preserves the old context. Unsupported document publication deactivates the native context through this handoff. NUL, malformed UTF-8, oversized context and unsafe C integer/rectangle conversions are rejected before native state changes.

The editor publishes owned document-wide rune state and visible cells from its actual shaped rendering, translated to UTF-16 and Cocoa view points. Geometry excludes the gutter, tracks wraps and scrolling, includes complete tab advances, newline ranges and blank-line insertion positions, and expands split styled graphemes. Matching text revisions invalidate stale geometry. Document metadata is cached by buffer identity/revision; line-start indexes map selections without repeated prefix scans. Composition underlines use visible cells.

The editor retains its existing fixed-grid keyboard, pointer and raster cursor/selection policy. Native insertion geometry publishes one strong caret per logical offset and physical line, preferring the downstream rendered fragment at styled boundaries; secondary bidi affinity and full visual editor navigation remain future work. Native hit queries use containing character cells; zero-width newlines support range queries, not point containment. These policies do not establish native-reference editor equivalence.

## Validation

Evidence is in [20261010-editor-native](../../prototypes/foundationintegration/evidence/20261010-editor-native). Focused tests cover multiline composition, undo/redo, backward cancellation, shared-view ownership, stale updates, read-only state, exact-buffer eligibility, saves/autosaves, dirty flags, close, metadata reuse and rendered wrapping/scrolling. The scrolled raster was inspected. No golden images were changed.

- Canonical rendered editor/native comparison passes 63 queries: 28 character cells, 3 newline ranges and 32 insertion positions. Expected screen rectangles derive independently from actual physical cells, framebuffer dimensions and Cocoa view bounds. TextField geometry regression also passes.
- Native editor protocol probes acknowledge 13 updates with zero rejected application actions. Saturation refuses eight admissions; the queue closes empty. These are direct NSTextInputClient protocol calls, not live input-method sessions.
- Native client handoff tests pass, including refusal, stale acknowledgement, origin routing and inactive superclass behavior. The 1,000-cycle hidden Cogent/Metal test attaches/detaches 1,000 times and drains 11,000 admitted updates.
- Go vet, transport/input/semantics race tests, focused exact-buffer/grapheme/adapter regressions and the existing core UI package pass. A cached 10,000-line EditingState microbenchmark reports zero allocations; it does not measure editing, rendering or full-app latency.
- The full upstream lines/textcore suite does not pass: lexer Rule/node conversion panics and a 90-second textcore timeout remain. Prior unmodified pinned control evidence reproduces the parser issue; the timeout is unresolved. See [the acceptance audit](FOUNDATION_ACCEPTANCE_AUDIT.md) and the retained package output.

Prism reviews the decoded nine-file toolkit patch and the Cocoa/routing/preparation changes separately, followed by documentation/evidence. Findings and dispositions are retained under `docs/reviews/editor-native-20261010/`. Review is distinct from runtime acceptance.

## Bounds and remaining gates

The native context ceiling is 1 MiB, but the unchanged transport ceiling is 64 KiB per encoded message, 256 slots and 4 MiB queue-owned payload. Encoding now includes both previous and prospective full text; JSON escaping also counts. Transactions exceeding admission credit refuse before native mutation. This is not a large-file composition implementation. Diagnostic history retains at most 128 records and 2 MiB of text, prior-text and kind bytes; other allocations remain outside queue accounting.

Live Japanese/Pinyin conversion and candidate-window placement, full VoiceOver workflows, editor visual navigation/native-reference parity, global resource budgets and interactive performance, broader regression compatibility, Windows/Linux, trusted release and upstream acceptance remain open. This task changed no input-source or VoiceOver settings and did not replace the previously installed sealed bundle. The deliverable is the reviewed local adapter prototype and its canonical Vellum commit.
