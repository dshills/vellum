# Fusion-style workbench — phase32-direct-handles

Move/resize axis handles, world-axis ray projection, transient preview, optional 1mm snapping, numeric deltas and one history commit per valid release. Escape/capture loss/outside release cancels. Default off.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: Preview now retains feature-marker visibility through detached displayObjects slices. Numeric transforms reject empty selection, invalid mode and axis. Renderer and scene run on the same owner, but backing-array reuse was removed regardless. Invalid deltas report an error and retain the last valid preview.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
