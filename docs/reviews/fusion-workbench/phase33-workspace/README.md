# Fusion-style workbench — phase33-workspace

Persist panel widths/visibility on orderly shutdown; AppKit toolbar order/hidden/display autosave and explicit reset; native File/Edit/View menus. Invalid preferences preserved until Reset Workspace.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: Added workspaceLoaded guard: startup failure before reading preferences cannot overwrite them. Reset intentionally restores owner.order, which includes configured fixed/flexible spacers and full-color actions. Hidden-pane branches use either a valid two-child split or center directly; all four combinations pass ui.Validate.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
