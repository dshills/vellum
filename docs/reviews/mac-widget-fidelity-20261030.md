# Internal review — macOS widget fidelity

Scope: shared MacTheme defaults, combo/secure/font/path composition, numeric
stepper geometry, managed button elevation, full-color inline spacing, menu row
sizing and pointer/keyboard highlight ownership. Reviewed the final source diff,
existing style resolution/styler order, relevant native Button/Switch defaults,
new/updated regression assertions and the fresh AppKit/gallery screenshots.
No external review service or source transmission was used.

Findings addressed:

1. Content minimums exclude padding/border. Reducing only theme padding leaves
   composites inflated; compact their embedded minimums and preserve form insets.
2. ButtonQuiet retains backend elevation allowance. Remove it only for explicit
   bordered controls with a managed palette, retaining focus outline reservation;
   ordinary backend controls and checkboxes keep their existing treatment.
3. Native SVG artwork already has a spacer. Add the new inline gap only for
   full-color artwork with a label, after which normal caller styles still apply.
4. Painting Hovered and Focused simultaneously can highlight two menu rows.
   Transfer focus on enabled pointer entry; render focus/active only. Disabled rows
   cannot focus or highlight; keyboard traversal after pointer movement is covered.
5. Geometry fixtures must use the gallery's embedded-input recipe. Their declared
   software scale is not evidence of changing the desktop's backing scale.

Verification and live limitations are recorded in the
[comparison report](../reports/mac-widget-fidelity-20261030/README.md). Focused race
checks, geometry/containment tests, retained draft/focus behavior, actual popup
routing/retirement, menu branches/checks and gallery tests passed. New menu test
checks a single highlight, disabled hover, pointer→keyboard transition, the chosen
stable identity and dismissal in both themes. Existing stepper tests cover
pointer halves, bounds, rejection and retained state; live checks separately show
upper/lower and keyboard changes. Packaged binary signature verification passed.

No exported API declaration, native dependency patch, rendering/pacing policy,
distribution approval or release gate changed. No blocking code findings remain
within this correction's scope. Clipped-edge physical targeting and remaining
custom-control material/shape differences are explicitly unqualified, not passed.
