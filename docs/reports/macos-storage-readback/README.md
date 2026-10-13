# Mac storage readback ownership

Storage staging reused a buffer based on the GPU allocation size, allowing a
16→8→16 shape change to retain an eight-byte staging buffer. Matching GPU-buffer
upload paths also skipped repair of missing staging. Batch mapping published
callback statuses into an unsynchronized slice and admitted members before
checking the complete batch. Reads could silently truncate a short destination.

Patch 51 checks shape operands before multiplication/slicing, live buffer
size/usage and known device/queue identities. Staging reuse checks staging's own
size; successful replacement retires the old reference and refusal preserves it.
CreateBuffer, SetFromBytes and WriteDynamicBuffer propagate staging errors and
repair matching-buffer paths. A valid upload can precede a staging refusal; GPU
content is not atomically rolled back. Legacy buffer-backed StorageTexture roles
retain their prior route; real textures use their separate layout/copy contract.

Full batch preflight precedes mapping. Foreign-thread callbacks capture only
buffered per-attempt status channels. After synchronous Poll, callback delivery
has a 30-second bound, which does not bound Poll itself. Failure cancels all attempted
exact buffers; refused cleanup quarantines staging until explicit ReleaseRead.
Late callback delivery cannot mutate owner state or replacement staging. Values
share one serialized, initially-unmapped owner; external native mappings and
copied native ownership are outside this contract.

Reads require this API's completed mapping and exact full-byte destination,
including dynamic padding. Short-target refusal preserves the mapping for retry.
Range/unmap refusal quarantines it before another read. Vars.Config retains
native layout validation and releases its temporary layouts immediately rather
than leaving them to finalizers; pipelines create independent layouts.

## Verification

Final GPU/gpudraw suites pass three race repetitions; the complete offscreen
adapter race suite, vet, eleven preparation tests and native consumer compilation
pass. Fault kernels cover staging allocation/admission/refusal, checked arithmetic,
foreign callback delivery, partial mapping, cleanup refusal, missing/late callback
and short-read quarantine. These are controlled Go factories, not hardware faults.

Nine tiny real headless storage cases verify 32 words, 512 dynamic bytes including
padding, two staging reference retirements, automatic repair through all three
upload routes, short-read retry, unmapped-range refusal and late-nil/duplicate/
mixed-device batch refusal with first-member reuse. The existing eight-dispatch
compute fixture additionally checks 32 shader result words and its resource
retirement. No native windows, surfaces, input or presentation are used.

The unchanged parent dependency with two test-only overlays fails six selected
native cases and eight assertions. The corrected original-compatible refusal
probe fails four Go-side assertions using an original-three-file overlay on the
final prepared dependency; new helper/unit-test files are disabled. It is not a
full parent-checkout execution. The first refusal probe accidentally reached an
invalid native device and Rust aborted before remaining assertions. Its raw exit 1
and complete log remain preserved; no physical GPU fault or passing acceptance
is inferred. The probe was corrected to encounter the nil member first in Go.
Earlier candidate/owned source versions remain separately identified.

## Review and evidence

Internal candidate review caught refused-unmap state remaining readable; the fix
quarantines it and has a regression that forbids a second range access. Final
internal review binds all eight source leaves, both static pin maps, independent
patch replay and seven decoded/prepared outputs. Prism ran once using Anthropic
Claude Sonnet 5.5, complete eight-file coverage, 3 chunks,
28150 input/6108 output tokens, raw exit 0 and no findings.

[Checkpoint](records/checkpoint.json) binds the source/evidence reviews and
[archive](records/evidence.tar.gz) with its [member index](records/evidence-index.json).
Raw failures, source variants, inert review snapshots, producers and receipts are
preserved; binaries/caches are excluded. All 6,195 prepared identities are bound,
not a complete compiler/SDK closure. Actual temporary-layout invalidation is not
independently counted. Native/GPU heap reclamation, physical faults, sustained
performance/endurance and full Mac acceptance remain open. Desktop work remains
deferred while the human uses the computer; no settings or input sources changed.
