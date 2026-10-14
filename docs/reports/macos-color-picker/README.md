# Inline color picker checkpoint

Vellum now provides an owner-held ui.ColorPicker with RGBA byte sliders,
a bounded hex field and an immutable checkerboard preview. It returns ordinary
managed FormGroup/Slider/TextField/Image/Label nodes, so callers use existing
focus, disabled ancestry, local themes and Mac accessibility publication. The
controls gallery includes a Colors page with reset, disable and focus actions.

ColorEdit.Apply patches the caller's current accepted straight-alpha sRGB model;
queued channel edits preserve each other and latent RGB at alpha zero. Hex edits
replace the whole value. Callers can accept, normalize or reject requests and
own their validation feedback. Unchanged preview pixels retain immutable image
identity; the cache retains only the latest image. See the
[contract](../../../specs/color-picker.md) and [usage](../../controls.md).

The actual offscreen test routes a right-arrow key to the focused red slider,
verifies retained field incarnation/canonical hex and checks an independently
computed alpha/checker pixel. A second key request rejected by the owner restores
the actual native Slider value. The optional 400x460 PNG records the disabled
state after the test. Initial layout checks exposed a general adapter defect:
Slider's final native 20em minimum overrode managed minimum widths. The adapter
now applies managed styling after that default. A regression fails on the original
controls.go overlay, then passes explicit/theme sizing and removal restoration
without replacing the slider. Original failed layout and fixture diagnostics are
preserved with source/version boundaries.

Full portable race tests, the full offscreen adapter race suite, portable/adapter/
native-gallery vet and default/native gallery builds pass. The first full adapter
run hit a sandbox Unix-socket bind refusal; the authorized unrestricted run
passes in 93.51 seconds. The first gallery compile referenced diagnostics absent
from ordinary/offscreen builds; complementary diagnostic build-tag helpers now
keep the reusable gallery buildable without the native-checks tag. A first slider
regression fixture lacked required numeric bounds; its corrected original-source
oracle fails specifically at the 20em minimum-width assertion.

The frozen native build passes eight actual Cocoa selector steps in 1.31 seconds:
red set, green increment, zero alpha, whole hex replacement, invalid hex, fractional
byte rounding, owner rejection and owner normalization to opaque alpha. Seven
valid proposals and one invalid proposal have exact owner counters and completed
native accepted values. All five ready inputs require nonzero native identities
and retain them. Disabled red and a stale hidden hex target refuse actions.
Main-thread cleanup observes zero windows, native outstanding operations, pending
callbacks and AX window registrations, with callbacks stopped. No OS preferences
or clipboard are touched. These are in-process Cocoa selectors; physical
keyboard/pointer, human VoiceOver, monitor/DPI, sustained performance and broader
P6/P11 gates remain open. Vellum remains an experimental Mac alpha.

Internal source review covered the API, preview, Slider precedence, gallery model,
FIFO native admission/owner barriers, queued cancellation/executing callback join,
build tags and cleanup. Its sole final identity-proof guard is corrected and
approved. Raw checks and producer/source/prepared hashes are indexed in the
checkpoint archive; all 6204 prepared dependency files remain unchanged.
[Qualification](records/qualification.json) and
[evidence index](records/evidence-index.json) record exact review/test coverage.
