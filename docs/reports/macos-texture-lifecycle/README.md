# Mac GPU texture ownership and headless readback

Owned patch 0049 publishes a texture only with a complete live view. Configuration
reuse requires an owned live pair on the same live device/queue with the required
usage bits; image/render transitions recreate incompatible allocations. Device
changes retire old texture/view/sampler/staging references. Empty borrowed state
is cleared even without a texture, and created pairs explicitly acquire owned
mode. Failure no longer leaves an incomplete pair eligible for the format shortcut.

Readback configuration explicitly releases replaced staging buffers; the original
relied on Go finalizers after overwriting its pointer. Same-device allocation
refusal preserves the old buffer/layout. Four-byte single-layer/single-sample
RGBA8/BGRA8 staging validates row/image fields and Go map arithmetic. Copy/mapping
refuse stale or invalid staging, and mapping never allocates uncopied replacement
data. GrabImage skips reading a refused copy. Raw ReadData retains native channels;
ReadGoImage normalizes detached BGRA bytes to NRGBA order with alpha/transfer bytes
unchanged. Headless admission returns an error for a missing adapter and retires
GPU/device admission references on failure; the shared instance remains owned by
final shutdown. Callers serialize allocation, copy/map/unmap and retirement.

## Exact final checks

Fresh preparation yields 6,187 files: three existing files change and four are
added relative to the original 6,183. Seven decoded outputs match their pins.
Final GPU/gpudraw offscreen race suites pass three repetitions. The full native
adapter offscreen race suite, GPU/gpudraw vet, native consumer compilation and all
11 preparation tests pass. Six final execution receipts bind unchanged eight
source leaves and all 6,187 prepared files, including the headless GPU execution.
The window-cycle binary is compiled only; its existing duplicate -lobjc warning
remains. No native window, surface, presentation, input, clipboard or preference
operation occurs in this phase.

The explicit headless GPU race fixture passes five cases: image→render→upload
usage changes retire two prior texture references; 32 staging replacements retire
all 32 held references; device replacement retires texture/view/staging and creates
new-device staging; two 3x2 RGBA/BGRA copy/map captures preserve exact native raw
bytes and produce the expected NRGBA channels. Each uses 512 padded staging bytes
for six opaque pixels. Old Go wrappers stay strongly held while IsValid becomes
false, establishing explicit reference release rather than finalizer activity.
These observations do not prove native/physical GPU heap reclamation.

Counted factory tables cover complete texture/view admission, nil/invalid/partial
results, preserved staging on refusal, five headless admission outcomes and 1,000
staging replacements with two owned handles only during replacement. Pinned
TryCreate APIs already clean their own native error references; partial-return
cases are defensive shared-control tests. Layout, device-change refusal and
four-format channel/alpha vectors accompany the seven wrapper-only regressions.

## Original failures and earlier versions

The original texture configuration fails all six invalid/borrowed same-format
cases and the empty-borrowed retirement assertion. The original-compatible GPU
fixture replaces only texture.go with its exact 9a34670 bytes atop final prepared
sources and current admission/test code. It fails usage reuse, staging retirement
at the second allocation, device-change cleanup and BGRA image channels. Its RGBA
case passes six pixels. Neither failed staging loop completes 32 replacements.
That baseline remains raw exit 1 with its overlay identity; it is not a final
prepared-source pass or a complete parent-tree execution.

The initial V4 sandbox GPU request sees zero adapters and panics in the old
NoDisplayGPU nil-adapter path before any fixture subcase. Its raw exit remains 1.
V5 fixes admission and registers fixture teardown before initialization. The final
GPU-access check passes all five cases. V1/V2 overlays, V3/V4 owned preparations
and V4 code passes retain their exact earlier source/prepared groups and are not
relabeled as final. Twenty receipts preserve raw exits and forty output logs;
four exits are 1. Frozen candidate and intermediate source copies are archived.

Internal review approves eight final source/spec/pin leaves and independently
replays seven decoded outputs. One Prism review using Anthropic Claude Sonnet 5.5
covers six patch/registration/doc leaves in three chunks, returns exit 0 with no
findings, and uses 27,243 input / 5,189 output tokens. Two static pin maps are
checked internally. The indexed archive includes exact selected source/prepared
maps, original overlays, producer scripts, source certificates, provider records
and raw outcomes. Binaries/caches/desktop artifacts are excluded. Compiler/cache/
SDK full closure is outside this evidence.

## Remaining Mac acceptance

This qualifies the stated code-level and tiny native headless pixel/reference
checks. It does not qualify GPU/native heap reclamation, physical device faults,
all readback formats/HDR/alpha, presentation, sustained frame/input latency or
endurance. The earlier IOSurface/physical-footprint growth still needs a quiet
live-window rerun. Desktop/input/performance tests remain deferred while the user
uses the computer. P11, broader Mac interaction/display, distribution and full
framework completion remain open.
