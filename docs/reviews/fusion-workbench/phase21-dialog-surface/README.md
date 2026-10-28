# Fusion-style workbench — phase21-dialog-surface

Actual packaged native action-panel opening exposed the same host-dark surface mismatch. Ordinary dialogs have a separate Core Body whose colors did not inherit the managed hidden host. Its final styler now applies only existing resolved background/foreground for ordinary dialogs; absent options retain native defaults. Popovers/sheets/layout/commands/public API and global appearance policy stay. A new regression changes process-native palette after Manual initialization before dialog opening, matching the native failure sequence. Neighboring panel/sheet/popover and lifecycle race tests pass.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: The review asks to cache style resolution and diagnose errors. Reconciliation already validates theme and node style before mounting/sync; supported inputs cannot produce a resolution error here. The existing applyStyle path uses the same guarded resolution on style passes. Reading the retained record on its owner preserves existing live theme changes while open; caching at construction would introduce stale colors. This bounded extra Body resolution changes no performance gate. The earlier initialization-only diagnostic did not reproduce the native opening boundary and is explicitly not counted as evidence for this fix.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
