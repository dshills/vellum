# Fusion-style workbench — phase23-appearance-parity

Ordinary dialog chrome now shares the existing resolved content surface/ink, and Close uses the existing theme Button preset. Forge keeps its intentional light design under either native host palette. Ribbon actions expose labels as tooltips. Light/dark native palette dialog tests and neighboring menu/overflow tests pass; the app geometry fixture now switches the native palette at the correct post-initialization boundary. No global settings, popup/sheet behavior, public fields or command policy change.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: Renamed the local Close variable to avoid builtin shadowing and documented its theme-only button descriptor. Applying dialog padding, growth, radius or text styles to Close would be incorrect; ui.ResolveStyle is already called by applyStyle with ui.Button and the retained current theme, and the regression checks the resulting Close foreground/background on both host palettes. Node-specific dialog colors still style its Body/Scene. The existing explicit Button preset is intentional.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
