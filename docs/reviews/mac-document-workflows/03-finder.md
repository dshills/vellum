# Finder delivery and native Quit integration — internal review

Reviewed optional native facade callback, cold-launch AppleEvent capture,
queue bounds/ownership, path checks, retirement, save/recovery routing and signed
package metadata. No root stable API or dependency patch changes. Examples
reject multi-file batches; startup fallback is deterministic. Private build22.

Review found AppKit Quit bypassed the Go Quit hook by calling the upstream
delegate directly. Added opt-in selector interception, returning Cancel while
app decisions proceed, restoring the original IMP on retirement. An inert
native regression proves interception and restoration without opening windows.
The Go path retains forced-shutdown checks.

Focused race tests passed: bounded Finder delivery/readiness/retirement, native
Quit interception/restoration, Go Quit coalescing, notebook dirty Cancel/Save/
Discard, cold-open precedence, recovery refusal, invalid requests, window
restoration, Forge study/recovery paths. All 35 packaging tests passed, including
document-type validation and signed metadata tamper refusal. Actual Finder
cold/warm launch and Command-Q remain final native checks, separate from this
automated evidence. No external review service was used.
