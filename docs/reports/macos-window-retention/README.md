# Mac window retention: portable fix, native memory gate failed

Closed portable windows now cancel the original lifetime, substitute one shared
canceled marker and discard the per-window cancel function. Historical identity
tombstones remain to reject reopening and stale delivery; this does not bound
historical ID storage or caller-held objects. The 1,000-cycle regression preserves
captured-context and descendant cancellation, completion revocation, survivor
independence and stale-work refusal. The parent implementation runs all cycles
but fails the retention assertion (1,001 cancel functions / 1,000 closed contexts).
The final portable app/tasks/document/history/workbench race suites pass three
times. The affected native consumer race tests and vet pass, and pre-profile
opt-in and offscreen driver race suites pass. Diagnostic sources are build
guarded and opted in; no ordinary native startup is performed by these unit tests.

The native consumer mounts a label, text field, checkbox, button and GPU rectangle
in each fresh child while a survivor remains. OnShow precedes owner readback;
each probe returns an immutable buffered acknowledgment. PostWait alone is queue
admission. An original quick-two process returned exit 0 with admission-only
checks and remains unqualified. A corrected quick-two process passes, checking
viewport/pixel, context cancellation, completion/stale-work refusal, atomic
composer withdrawal, locked registry removal, visible/AX inventories and retired
canvas refusal. Registry retirement is not native or GPU deallocation.

## Production 1,000-cycle gate failed

Two corrected production processes stop at 375 of 1,000 cycles when sampled RSS
exceeds the 1 GiB safety guard: 1,112,784,896 bytes without profiling and
1,102,462,976 bytes with profiling. Live Go heap is about 388 MB. Both return raw
exit 1 after normal zero-inventory cleanup; neither is a 1,000-cycle pass.
Post-GC samples every 25 cycles do not bound unsampled peaks. Self ps RSS is
resident pages, not physical footprint or native/GPU attribution. An optional
exclusive new-file heap profile is capped at 8 MiB compressed; failures remain
failures. The retained profile is approximately 74 KiB. Pprof attributes 353.66 MB
(90.77% of sampled in-use space) to embedded-font parsing through new shapers.
Allocation-site attribution alone does not establish the retaining reference.

## Unadopted cache experiment and native footprint

The global glyph-mask cache has unbounded strong font-face keys. A one-file Go
overlay outside the prepared tree tests FIFO whole-face eviction with 32 face
keys and 8 MiB of mask data. Prepared source remains unchanged. This candidate
completes 1,000 cycles with raw exit 0 and empty stderr: heap reaches about 65 MB
by 25 cycles and ends at 65,709,896 bytes, while RSS still rises to 933,494,784
bytes. This is diagnostic evidence for a cache intervention, not adopted source
or a production memory pass. Parsed fonts, allocator overhead, other caches and
caller-held masks are outside its stated bounds; meaningful cache boundary,
concurrency and pixel tests and owned-patch adoption remain necessary.

A separate bounded 800-cycle candidate process completes with exit 0 before the
quiet-host authorization expires. Vmmap queries only its Popen-owned PID at 250
and 750 cycles. Physical footprint is 7.9 GB then 14.3 GB, and IOSurface is 7.2 GB
then 13.5 GB (as reported in macOS human-readable units); CoreAnimation and native
malloc also grow. These observations expose significant native graphics retention
despite the sampled RSS guard passing. They do not identify a precise allocator
owner. An earlier diagnostic recipe fails its remaining-window assertion before
launching an app; its raw exit 1 and original recipe are retained. The successful
800-cycle diagnostic does not replace the failed production 1,000-cycle gate.

## Review, provenance and remaining gates

One Prism review uses Anthropic Claude Sonnet 5.5, covers eight pre-profile files
in three chunks, consumes 22,381 input / 2,419 output tokens and returns raw exit
1 with one medium finding. Internal inspection refutes the unsynchronized-registry
claim against pinned source: FindName holds renderWindowGlobalMu and IsClosed
loads an atomic composer pointer. Names stay fixed in this fixture; no arbitrary
rename or other mutable-field safety guarantee follows. The raw result remains
exit 1. Internal exact-source review also covers the later diagnostic profile
extension without another provider call. The source certificates, disposition,
all process logs, original/cache overlays, profile, vmmap output, version groups,
and 6,174 unchanged prepared file hashes are retained in the indexed archive.
Binaries, app bundles, caches and desktop snapshots are excluded. Evidence covers
selected source and prepared identities, not the full compiler/cache/SDK closure.

All live processes finished before 2026-10-13 10:07:09 UTC. Further live/input and
performance checks are deferred pending user availability; code, offscreen,
review and commit work continues. This campaign changes no input sources,
VoiceOver, notification permission, clipboard or user preferences. Unique-title
geometry saving is disabled only in the fixture. Full P11 component/endurance,
native/GPU memory, latency, human IME/VoiceOver, Finder drag, cross-display and
public distribution gates remain open. Mac framework completion is not established.
