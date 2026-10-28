# Fusion-style workbench — phase39-recent-revert

Recent Studies bounded MRU list persisted in workspace preferences, native File actions and shared guarded Open. Explicitly confirmed Revert to Saved reloads exact current path; cancellation/failure preserves edits. Added prompt epochs and document-session integration tests. Clipboard shortcuts defer to focused text controls.

Prism used Anthropic Claude Sonnet5.5 with complete text coverage and retired
review processes. Original findings and receipt remain unchanged.

Findings: Confirmed-open/revert dispatch now uses captured pending path and both cases are explicitly handled. Regression exercises real document.Session reload/cancel, plus recovery dialog metadata. recordRecent runs only on Err==nil completion, after Session validated/installed data, and ignores empty/invalid paths. Successful reverts deliberately update MRU.

Focused checks are recorded in checks.json. Automated routing/geometry evidence
does not establish physical pointer, human accessibility or CAD-scale performance.
