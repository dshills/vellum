# ListBox native writability metadata

The external Mac accessibility client reported `AXSelectedRows` as writable
after the owner made a ListBox read-only. The modern selector gate and owner
already refused changes. This follow-up aligns the native metadata with that
policy using a private list-specific Cocoa class and its optional legacy
writability method. Other managed control classes retain their existing path.

The legacy method is absent from the superclass on the tested system. Delegation
therefore checks the actual superclass implementation before calling it; unknown
attributes safely refuse when there is no inherited implementation. Modern
setter admission and owner validation remain independent of metadata.

Compatible lists retain their Cocoa identity through read-only/disabled changes.
Crossing between list and ordinary control classes retires the old object and
preserves ordinary ID-based geometry, including when the Go bridge suppresses
an identical next geometry payload. Text geometry and scoped publication retain
their existing retirement rules.

Validation records distinguish automated inert Cocoa lifetimes/class transitions,
full offscreen race/vet checks, and a private native gallery inspected through
the real external AX API. The external scenario checks selected-row and host
stable-ID requests, owner rejection, read-only/disabled refusal, programmatic
clear, restored writability and unsupported-attribute survival. An original
native binary supplies the failing writability comparison.

Acceptance requires the raw attempts, source maps and review to be retained in
[records](records/qualification.json).
Diagnostic instrumentation is separate from the final production build. The
initial Objective-C overlay did not establish instrumentation; a later startup
race and an invalid direct legacy-message oracle are preserved as failed
fixtures, not product acceptance or proof of the original metadata behavior.

`LIST-NATIVE-01` is closed only for this native metadata/selection scope.
`LIST-NATIVE-02` remains open: foreground ownership could not be verified, so
native pointer/key delivery and focus reveal remain unqualified. No OS input
sources, VoiceOver or preferences were changed. Owned app/inspector children
are reaped; this is not a native/GPU allocation-zero claim. Full Mac performance,
endurance, physical/human accessibility and distribution gates remain open.

This small follow-up uses internal review after the complete Anthropic Prism
ListBox checkpoint; it does not claim another provider review.
