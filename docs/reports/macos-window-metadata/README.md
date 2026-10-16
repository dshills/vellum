# Sampled Cocoa window metadata

Opt-in native diagnostics now record runtime window/content-view class names,
the signed Cocoa window number and explicit metadata availability. Class names
are limited to 256 UTF-8 bytes; invalid/oversized metadata refuses the snapshot.
The existing 64-window and wire bounds remain. No history registry or retained
native object is introduced. See the [contract](../../../specs/native-window-lifecycle.md#sampled-cocoa-window-metadata).

The actual serializer and the Go DTO are shared with an inert Foundation
fixture. Two targeted race tests pass for original boolean state, exact signed
number boundaries, missing views, UTF-8 byte limits and legacy absent metadata.
Two isolated faulty copies fail the same Go regression at the intended assertion.
The complete offscreen adapter race suite and vet pass. Native-check tests and
the services/window-cycle commands compile; those executables were not run.
Existing Cocoa deprecation and duplicate Objective-C library warnings remain
in the raw logs.

[Qualification](records/qualification.json) and [archive index](records/archive-index.json)
bind the tested source endpoint, receipts, raw logs, controls, producers and
internal source review in [the evidence archive](records/evidence.tar.xz).
The initial baseline clone/copy was an inline tool command; its source/prepared
vectors are preserved, with no prebound saved stage producer claimed. Producers
retain their original private paths. Compiled native binaries are excluded.

Runtime class queries and native-check partial-decode cases remain unexecuted.
The old hidden untitled Cocoa window is still unidentified. Class names describe
types and Cocoa may reuse window numbers; these fields establish neither managed
ownership, lifetime identity, deallocation nor a leak. Historical raw records
stay unchanged. Native sampling, current-source lifecycle, performance/memory,
overnight, physical/human and distribution acceptance remain open. No live UI,
GPU application or OS setting was used. This is a small internally reviewed
checkpoint between significant Prism reviews; it does not complete the Mac goal.
