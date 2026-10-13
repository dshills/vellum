# Managed image previews

`ui.Image` reuses immutable `canvas.TiledImage` pixels as a passive retained
control. Contain centers all pixels, cover centers and crops, stretch fills both
axes. Fit changes painting, not the layout box. Preferred size caps the long edge
at 256 device-independent pixels, one minimum derives its partner, and max/style/
grow determine the box. Source/fit/style updates retain the widget. A named image
exposes AXImage/alternate text; an unnamed image is decorative. Neither adds an
action or Tab stop. See [catalog](../../controls.md), [contract](../../../specs/controls.md)
and [source-image contract](../../../specs/images.md).

The leaf scales immutable source through the ordinary scene painter using full
content coordinates and a composed visible-content/ancestor/opacity mask. It owns
no resized bitmap or cover-sized intermediate. Scene packet preflight safely
refuses this transform/custom-mask route; MeshCanvas direct GPU texture/tile
uploads have a separate contract. Removal drops the widget source reference while
caller-owned data remains usable.

`ui.DecodeImage` reads caller-owned PNG/JPEG data, bounded to 32 MiB plus one
refusal byte. Direct format configuration validates 1..8192 axes before decode,
then normalizes/copies immutable straight-alpha pixels. Retained pixel bounds
differ from temporary decode/conversion memory and reader time. Large decoding
belongs on an application worker with caller-owned reader cancellation. No path,
URL or chooser lookup occurs.

## Verification and review

Full portable/offscreen adapter race, image/gallery race3, vet and native gallery
BUILD ONLY pass on all sixteen final root leaves and unchanged 6,198 prepared
identities. The final full adapter execution takes
89.08 seconds. Actual Manual frames check
independent centered-fit colors, cover overflow, straight-alpha blend, immutable
replacement/incarnation, real wheel mapping and viewport clipping. Two windows
share a source at logical DPI320; retiring one preserves the other. Direct fixtures
cover size/max/restoration, source release, alternate-text/decorative semantics
and mask composition. Inert Cocoa checks AXImage from an actual managed descriptor.
Decode fixtures check PNG/JPEG normalization, ownership, bytes, direct dimension
refusal, unsupported/corrupt data and reader errors. One inspected cover PNG is
archived with capture/source hashes; it is software visual evidence.

Three localized mutation overlays compile and assertion-fail: removing the image
mask paints outside the cover box; reanchoring to the clip loses scrolling;
bypassing dimension preflight enters decode. The same final fixtures pass race3.
Earlier unmasked scroll mutation returns raw0: its parent integration did not
expose image-owned clipping, so that oracle cannot prove the property. Independent
cover overflow now tests the image-owned mask. These are new-code fault seams,
not full original-production builds.

Internal source review approves exact final16. Prism runs once using Anthropic
Claude Sonnet 5.5, complete sixteen-file coverage, 3 chunks,
35,287 input/2,739 output tokens, raw0/no findings.
Production bytes remain unchanged throughout later fixture corrections; source
review and execution scopes remain distinct. No repeated provider run or broad
Mac completion is inferred.

Two initial native attempts fail on a wrong Manual explicit-Name assumption and
a bilinear stripe-boundary sample. Final fixtures use the declared Manual name
and interior colors; semantic Text fallback has separate direct/native checks.
A first raw0 capture omits its file because child environment filters the supplied
variable. The preserved original runner and corrected explicit destination produce
the inspected PNG on a new count1 run. These remain historical attempts.
25 receipts/50 logs preserve 7 raw exit1 attempts.
Exact selected source families and mutation inputs are archived; the historical
map lists intermediate fixture gaps. Binaries/caches and full compiler/SDK/native
dependency closure are excluded.

Separate evidence review binds the [checkpoint](records/checkpoint.json),
[archive](records/evidence.tar.xz) and [member index](records/evidence-index.json).
Physical Mac display/input/VoiceOver, native/GPU memory, sustained performance,
endurance, distribution rights/signing/release and full Mac completion remain
open. Native gallery compilation does not launch it.
