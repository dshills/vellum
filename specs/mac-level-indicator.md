# Mac-style level indicator

`ui.LevelIndicator` is a retained, owner-held public Go composite. It uses
existing labels, immutable images and ordinary managed buttons; it introduces no
native node kind, scheduling policy or rendering backend. It is custom rendered,
not an AppKit `NSLevelIndicator` wrapper. Fixed `Dark` selects local light/dark
colors without changing operating-system appearance.

## Capacity

`Mode` is `LevelContinuous` or `LevelDiscrete`. `Min`, `Max` and `Value` must be
finite, `Min < Max`, and `Value` must lie within the closed range. Fraction
arithmetic uses float64, including ranges that span both finite float32 extremes.
Capacity is passive: `Change`, `AllowClear` and `Unrated` are invalid in these
modes and the component adds no focus targets.

An optional absolute `Warning` threshold changes green fill to amber; optional
`Critical` changes it to red. Each must be within the range. If both are set,
warning must not exceed critical. Comparisons include the threshold itself;
critical takes priority. Visible status text and accessible descriptions convey
Normal, Warning or Critical as well as color. Disabled meters use muted fill.
Capacity descriptions include current value and minimum/maximum range.

A continuous track represents the exact proportion. Discrete mode defaults to
10 segments and accepts 1–40; a partly occupied segment is lit (ceiling of the
fraction times the segment count). Zero empties all segments. Segment count must
be zero outside discrete mode. The preview is a fixed 240×14 logical-pixel
rounded track, backed by an immutable 480×28-pixel image. Unchanged inputs reuse
the same image; changed inputs replace it without mutating old descriptions.
The component retains only its most recent preview.

## Rating

`LevelRating` requires `Min=0`, integral `Max` from 1 through 10, integral `Value`
in range, no segment settings and no thresholds. Filled/outlined stars are
ordinary buttons with a stable `LevelStarID(root, value)`. Existing pointer,
Tab and Return/Space routing supplies accessibility and keyboard interaction.
There is no special arrow-key or drag-to-rate interaction.

`Change(LevelEvent)` proposes a new rating. The application decides whether to
accept by updating `Value` and `Unrated`. A nil callback is read-only. `Disabled`
mutes the stars and disables all actions. A proposal equal to the current
accepted state produces no callback, including repeated Clear after acceptance.

The zero button (`LevelZeroID`) explicitly assigns a **rated zero**.
`AllowClear` adds `LevelClearID`, which proposes `Unrated=true, Value=0`.
Unrated is accepted as a description only when AllowClear is set and Value is
zero. It displays “Unrated” and empty stars. No star click implicitly clears a
rating. Each action has an accessible name that identifies its proposed value
and the current rating; the zero and clear descriptions explain the distinction.

Actions verify the current owner-held identity, mode, range, clear policy,
callback, enabled state and valid configuration at delivery. Rendered disabled
or read-only actions cannot become active without a new description. Changing
component identity or structural configuration, or returning an invalid
replacement view, invalidates the prior action generation. No hidden model value
or asynchronous work is retained.

## Demonstration and verification

The private gallery's Level indicators page shows continuous/discrete capacity,
warning/critical thresholds, an editable five-star rating, explicit zero/clear,
rejected proposals and disabled controls. A shared sample-capacity slider changes
both passive meters. The helper is `newMacLevelGallery(dark).view()`.

`ui/level_indicator_test.go` checks actual source pixels for capacity fractions,
threshold colors, discrete gaps, dark/disabled palettes and immutable preview
reuse. It also checks offset/extreme bounds, controlled acceptance/rejection,
zero versus unrated, duplicate suppression, keyboard focus ordering, callback
revocation, stale/disabled/invalid actions, input bounds and namespace collisions.
These are automated semantic/source-image checks; they do not constitute
physical pointer, native focus-ring, Retina display or human VoiceOver evidence.

Known limitations: fixed preview width, fixed local colors, font-rendered stars,
button-based rating navigation, and no automatic system appearance/accent/locale
tracking. Native packaging and visual checks are separate gallery checkpoints.
