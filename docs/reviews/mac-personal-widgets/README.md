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
