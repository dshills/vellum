# Inline color picker

`ui.ColorPicker` is an owner-held reusable component whose `View()` returns an
ordinary FormGroup with four Slider fields, one TextField, a Label and an Image.
It adds a public UI component without a new Node kind or native color-panel
backend. Retain the component on one application owner; do not concurrently
render or mutate an instance from independent owners.

The accepted `Color` is straight-alpha sRGB `color.NRGBA`. Channel proposals use
`ColorEdit{Channel, Byte}`. Apply a proposal to the **current** accepted model via
`edit.Apply(picker.Color)` to preserve other queued channel edits and RGB bytes
when alpha is zero. A hex proposal replaces all four bytes. The callback may
accept, normalize or reject any proposal. Ordinary managed reconciliation
restores the accepted values after a native commit. Nil Change is read-only;
Disabled disables the entire group. Invalid input calls Invalid when supplied;
the caller owns ValidationMessage and decides when to clear it.

Sliders cover 0..255 in steps of one. Finite fractional requests in that range
round to the nearest byte before delivery. Nonfinite/out-of-range component
requests are invalid. The native adapter may refuse invalid numeric input before
it reaches the component. ParseColorHex accepts #RGB, #RGBA, #RRGGBB and
#RRGGBBAA in either letter case with surrounding whitespace, bounded to 32 input
bytes before trimming. Three/six digit forms imply alpha 255. ColorHexString
always returns uppercase #RRGGBBAA. Named colors are unsupported.

The component owns ID + "/color/". ColorFieldID(ID, channel) identifies the red,
green, blue, alpha and hex inputs for normal focus requests. Ordinary ui.Validate
detects collisions with caller nodes. The group/image/labels are passive; only
its five enabled inputs enter normal focus order. Supported Mac semantics derive
from those managed controls; this is an inline group, not an NSColorPanel.

The preview is an immutable 32x32 opaque image with 8-pixel checker squares at
240 and 176. Each displayed byte is the rounded straight-alpha blend with that
background. The retained picker caches only its latest image. Unchanged colors
reuse it; changed colors apply an immutable patch. Transparent RGB edits retain
latent model bytes and may reuse identical rendered pixels. No global cache,
image-history chain, file, clipboard or OS setting is involved.

Explicit Slider size styles and theme overrides run after the backend's final
minimum-width default. Removing overrides restores that default while retaining
the widget. Validation includes portable format/merge/cache/refusal tests,
actual offscreen routed arrow input and checker pixels, and a bounded Cocoa
selector campaign for accepted, rejected, normalized, disabled and hidden edits.
These scoped checks do not qualify human VoiceOver, physical input/display,
sustained performance or the broader P6/P11 acceptance gates.
