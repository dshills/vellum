# Fusion-style workbench — phase3

Reusable experimental small-scene viewport with depth, picking, orbit/pan/zoom and a real orientation cube plus keyboard alternatives.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: 592a8ba76f4859f2: private view() was only called after HandleInput validated Model; added a defensive nil guard. 4bced99071e01a6a: gesture metadata now updates only after accepted Begin, refuses nested/stale sequences and clears cube metadata on release/cancel; regression includes restored camera. da7ef4047b1f261c: verified existing GPUResources.Buffer compares the complete descriptor (gpu_resources.go:71), replacing and releasing old handles on size change (84–90). Helper-owned buffers must not be released by programs; GPUView frees them after Program.Release. Native GPU readback and resized mesh acceptance will be checked in the integrated demo.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
