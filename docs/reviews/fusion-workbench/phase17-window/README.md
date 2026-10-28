# Fusion-style workbench — phase17-window

App-only whole-window polish: independently scrolling inspector in short panes, bounded horizontal orientation scroller, accurate caption for existing projection state and unchanged theme factory shared with tests. Nine measured app-description geometry cases cover1440x920,1000x650,1000x420 at160/240/320DPI. Only the scene Extension is replaced by a measured placeholder in offscreen layout tests; no GPU/physical-pointer evidence is claimed there. No feature/API/backend/release policy changes.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: Review suggested loading embedded font fixtures once; this is implemented with immutable shared fixture data, detached by the existing Manual constructor for each subtest. The new test is deliberately offscreen-tagged and scripts/check.sh explicitly runs the Forge module with that tag, so the claim that normal CI would omit it is false. Final focused geometry and app/vet/API checks pass.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
