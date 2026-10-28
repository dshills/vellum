# Fusion-style workbench — phase40-overlay-cursor-repair

Live found an incompatible implicit GPU bind-group layout when handles enabled. Use cube-layout group bound to main camera; release with renderer. Smoke now enables handles and checks nonblank readback. Reduce per-window macOS original cursor artwork from32 to20dp; no global pointer settings changed.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: Prepare now retires old groups before resource-generation recreation and releases partially initialized groups on failure. Uniform buffers keep fixed64-byte descriptors within a generation; all three groups rebuild together. No separate main-camera replacement occurs on resize. Live copy/paste/duplicate passed before this rendering-only repair; clipboard formats restored byte-for-byte.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
