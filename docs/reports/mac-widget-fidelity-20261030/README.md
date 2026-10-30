# macOS widget fidelity comparison — 2026-10-30

Compared a fresh Mac Widgets gallery from main b5ed1830 against actual AppKit
controls built locally on the same macOS 27.2 desktop. This is a representative
fidelity maintenance pass, not qualification of every widget or a native-wrapper
migration. Reference source is [NativeControls.swift.txt](NativeControls.swift.txt).
Copy it to `NativeControls.swift`, then compile with
`xcrun swiftc -target arm64-apple-macos27.0 NativeControls.swift -o native-reference`;
use the existing private app packager to create a reference bundle. Its Light/Dark
controls affect that process only.

## Findings and corrections

| Area | Observed difference | Correction |
| --- | --- | --- |
| Regular controls | Buttons/fields too tall; strong outlines | Compact vertical padding, subtler light/dark borders |
| Composite inputs | Family/typeface/password controls inflate their containing forms | Reduce content minimums; retain embedded editor insets and the shared focus outline |
| Quiet actions/path | Push-button outlines and reserved backend elevation add visual weight and height | Flat path segments and reveal action; remove backend shadow/offset reservations from explicitly bordered Mac palette buttons |
| Numeric input | Oversized arrow button and example content minimum | Compact trailing stepper and demo/font-size fields; keep numeric validation/routing |
| Color artwork | Full-color image touches inline label | Six logical pixels between image and label; explicit caller gap still wins |
| Popup menus | 24-pixel content minimum plus padding makes 32-pixel rows; keyboard focus and hover can paint two blue rows | Use a 16-pixel content minimum; pointer entry transfers focus, and only the focused/active enabled row paints the highlight |

Actual AppKit regular field/button bounds were 26 points; its small/large buttons
were 22/30 points. Final live Vellum font family row was 29 points including its
shared outline; font size was 27. These are comparisons, not new release budgets.
The deterministic geometry test permits font/rounding variation and checks
containment at widths 480/660/960, scale 1/2, and both themes. Composed secure-entry
fixtures use the same zero-outer-padding recipe as the gallery; child editors own
the text inset. Explicit application style overrides remain supported.

## Evidence

All images are exact saved CUA screenshots from this run, inspected after saving.
Reference window is inactive in the first image: use it for geometry, not claims
about active accent colors. Samples contain only synthetic demo values.

- [AppKit controls](03-appkit-reference-inactive.jpg) and [native popup](04-appkit-popup.jpg).
- [Before standard controls](01-vellum-themes-before.jpg) / [after](05-vellum-themes-after.jpg).
- [Before inputs](02-vellum-inputs-before.jpg) / [after](07-vellum-inputs-after.jpg).
- [Dark controls](06-vellum-dark-after.jpg), [single keyboard highlight](08-vellum-menu-keyboard-after.jpg).

## Verification

Pinned Go 1.25.6 and prepared native sources were used. Passing focused checks:

- `go test -race ./ui` and `go vet ./ui`.
- Native `go test -race -tags offscreen .` filtered to Mac field/number/input/personal
  geometry and disabled routing, shared focus, combo, secure, path, font, full-color
  button and icon-above-text regressions.
- After the final menu-height correction: native `go test -race -tags offscreen .
  -run 'Test.*(Menu|Combo)' -count=1`.
- Native `go vet -tags offscreen . ./cmd/gallery`; gallery race tests filtered to
  input, Mac, theme and toolbar cases.
- Private gallery packaged with explicit ad-hoc signing; `codesign --verify
  --strict --deep` passed. No signing identity or notarization was used.

Separate desktop evidence with the owner-approved app-restricted mouse helper:
Save preview advances the action count once; font stepper upper/lower halves change
16→17→16; Down changes 16→15 and Up restores 16. Wheel input scrolls the form.
The unit popup opens aligned to the full field; a pointer click selects Inches and
dismisses it. Down/Return selects Centimeters; Escape dismisses the dark selector.
Light/dark controls and full-color toolbar sizes were visually inspected.

CUA scrolling failed to reach the form, so those attempts are not pointer evidence.
Some helper targets at a clipped lower-edge popup and the stepper's bottom border
missed; only subsequent observed value changes count as successful targeting.
Physical selection at the clipped popup edge remains unqualified; the fully
visible popup and keyboard route passed. Native display-scale changes were not
performed. Existing build24 installations and user documents were untouched.

## Remaining fidelity limits

Custom controls still differ from AppKit materials/animation and some shapes,
including checkbox/switch/slider artwork. This pass does not replace their native
driver classes or claim pixel identity. Current screenshots also show the existing
shortened path labels and app-specific form grouping. Human VoiceOver/IME usability
and exhaustive widget-by-widget qualification remain outside this comparison.
