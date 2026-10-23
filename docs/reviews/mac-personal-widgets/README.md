# Mac personal-use widget checkpoints

Each source checkpoint uses Prism with Anthropic claude-sonnet-5-5, complete
staged-diff coverage, and focused race/routing tests. Native app observations are
recorded separately after packaging. No release/pacing/rendering policy changes.

## Color well

Root UI race suite and offscreen ColorWell/ColorPicker managed tests passed.
`color-well.json` reports a medium-confidence stale-session concern. Retained as
intentional: opening must advance the session to revoke closed-panel callbacks;
old button descriptions must not toggle a newer panel session. Owner dispatch
reconciles the view after activation, before further native editing. The managed
routing regression opens through Return, commits hex through the live field, and
closes through Escape; component tests cover old sessions and read-only state.
The suggested removal of the opening epoch would permit callbacks from a prior
closed description to reach a new editor. No source change made for this finding.

## Token field

Focused UI race tests, managed committed additions and keyboard chip removal, and
gallery model tests passed. Prism completed with zero findings. Completion menus
reuse ComboBox; no claim of per-keystroke filtering or Mail-style inline editing.

## Font picker

Focused UI race tests, managed Up-arrow size changes, committed family changes,
preview typography and gallery tests passed. Prism's catalog-iteration suggestion
was applied and retested. Both reviews also flag silent refusal of typeface events
from a prior family; this is deliberate stale-event protection, covered by UI
tests. Those actions must not edit a different family's typeface or emit user
validation errors. Accepted callbacks are followed by owner reconciliation, which
installs current-family controls. No remaining actionable finding.

## Path control

UI race tests, icon artwork tests, managed overflow-menu navigation and gallery
model tests passed. A focused managed test confirms disabled read-only rows
propagate Disabled to both icon and label; explicit child flags are redundant.
Prism then identified misleading read-only overflow help: corrected to expose
the omitted ancestor labels without an action hint, with a new UI regression.
Final review flags redundant navigation in a two-item path. Retained deliberately:
the current segment consistently offers its ancestor menu, even when that sole
ancestor also has a direct button. This is optional UX preference, not a failure
of controlled navigation or lifecycle. No remaining actionable finding.
