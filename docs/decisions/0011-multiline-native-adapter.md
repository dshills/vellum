# ADR 0011: Bind the multiline exact editor to the native text client

Date: 10 October 2026. Status: accepted for the experimental macOS foundation; G1 remains open.

Extend the owned Cogent exact buffer with an atomic expected-text transaction seam. Share preedit with all views while keeping one composition owner and one net undo group. Seal composition before unrelated mutations, save and undo; cancel restores its baseline only while ownership and expected text still match. Close discards the baseline. Autosave snapshots committed text while preedit is active.

Bind the existing Cocoa window client to a stable control ID and carry origin ID plus previous text in every admitted update. Drain and acknowledge old-client composition before changing the binding. Reject stale compare-and-edit updates and unsafe publications. Unsupported documents relinquish native activation through the same handoff. Keep the bounded transport and insertion-only compatibility APIs.

Capture cells and insertion stops from actual visible rendered editor fragments, retaining whole-document rune/UTF-16 mappings, tab advances, newlines, viewport clipping and revision invalidation. Cache document metadata and use its line-start index for selection mapping. Publish one strong insertion stop per offset/physical line; secondary bidi affinity is outside this seam. Keep the editor's existing fixed-grid keyboard/pointer/raster cursor policy, explicitly retaining full editor visual navigation as an acceptance boundary.

Install patch 0012 only after pinned-input hash verification. Review decoded toolkit sources and native integration separately with Prism, and preserve focused/native tests plus known broader failures. See [the implementation report](../reports/MULTILINE_NATIVE_ADAPTER.md). No live IME/screen-reader, production performance, other-platform, distribution or upstream acceptance claim follows from deterministic protocol checks.
