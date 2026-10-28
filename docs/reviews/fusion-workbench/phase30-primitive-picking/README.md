# Fusion-style workbench — phase30-primitive-picking

Pick preview cylinders by their capped elliptical surface, with nearest-hit order and focused ray regression coverage.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: Degenerate-bounds concern is guarded twice: Scene.Pick calls Scene.Validate (which calls Box.Validate), and hitObject returns immediately when Ray.Hit rejects invalid bounds before computing radii.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
