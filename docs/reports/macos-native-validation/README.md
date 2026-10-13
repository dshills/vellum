# Native snapshot integer validation

Native snapshot validation now reads exact integral NSNumber values directly,
avoiding repeated decimal formatting of common generation/incarnation fields.
Decimal/floating and unknown encodings retain the preceding strict digit-string
path, including UINT64_MAX boundaries. The candidate passes the live stored
400-cell Cocoa fixture, focused repeated race checks, the full adapter race suite
and vet. Full Mac and latency qualification remain open.

## Diagnostic basis

Two optimized fixed-10k processes at commit
`f3caa2e668839395f0e7ed66f57d39f8e0941df7` capture separate default/GPU CPU
profiles, with 128 guarded warmups and 120 events each. A third GPU process has
240 profiled events and an owned five-second native sampler admitted after warmup.
These 480 profiled observations are excluded from ordinary latency qualification.
Source/prepared receipts bind 710 selected source and 6,198 prepared identities;
they do not establish a complete compiler/SDK/cache closure.

Go profiles attribute substantial time to C calls, including the native event
loop. The native sample distinguishes waiting from main-queue accessibility work
and exposes NSNumber string formatting inside snapshot validation. It also shows
Metal buffer-write paths. Nested/overlapping stack counts are not added into CPU
time. External sampling perturbs the observed process and does not measure GPU
completion, transfers, presentation or input-to-photon latency.

## Change and correctness

The candidate changes one native helper and its test leaf. Signed integral
encodings reject negative values; unsigned integral encodings preserve their
complete encoded width. CFBoolean values remain refused. Decimal/floating and
unknown encodings still use the preceding digit/range check. Snapshot structure,
monotonic native epochs, scoped pending/completed commits, identity, callbacks,
geometry and retirement follow their existing contracts.

The boundary matrix runs first against the original implementation and passes,
then passes with the candidate in the focused race suite repeated three times.
It covers zero/positive fields, integer-width transitions, values around 2^53 and
2^63, UINT64_MAX, overflow, fractional/exponent values, booleans and nonnumbers
through actual Foundation JSON parsing. The complete offscreen adapter race suite
and vet pass. The optimized candidate separately passes actual stored 400-cell
Cocoa value/frame/identity/action/closure acceptance.

## Local benchmark

The fixture validates one complete JSON snapshot containing 400 cells, 50 rows
and one host (451 nodes). Five sequential original observations span
2.241–2.258 ms/op, with median 2.249 ms/op. Five selected candidate observations
span 1.928–1.961 ms/op, with median 1.936 ms/op. This is a local validation-only
observation, not a controlled full-frame speedup or sustained guarantee. Native
allocation is not measured by Go's reported allocation counter.

Earlier candidate benchmark batches overlapped race checks or a build; they are
preserved as exploratory records. The selected final batch runs after those jobs
terminate. Background host isolation is not continuously established. No ordinary
P6 campaign is added here; the preceding checkpoint's failed 4/6 ms results remain
the latest ordinary qualification. Native main-queue validation lies outside the
recorder's exact UI-work total, so this result cannot be subtracted from it.

## Evidence and review

Internal source and diagnostic-evidence reviews precede one bounded Prism review
with Anthropic Claude Sonnet 5.5. The [checkpoint](records/checkpoint.json),
[archive](records/evidence.tar.xz) and [member index](records/evidence-index.json)
preserve exact source variants, receipts, diagnostics, benchmarks and reviews.
See the [accessibility contract](../../../specs/accessibility.md).

Human VoiceOver/input, physical display/GPU cases, sustained performance,
integrated P11/endurance and distribution/release remain open. These fixtures
require no input-language, VoiceOver, clipboard or persistent preference changes.
