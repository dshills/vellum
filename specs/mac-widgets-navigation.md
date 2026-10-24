# Personal Mac widgets: next batch

This user-authorized widget expansion adds Sidebar, Sheet, CollectionView,
ShortcutRecorder and LevelIndicator. It leaves rendering, pacing and release
policy unchanged. Every phase has focused tests, Prism Anthropic Claude Sonnet
5.5 review and a commit; final private bundles include all five.

## Sidebar

Owner-held Sidebar composes bounded sections (32, 256 total items), Disclosure
and ListBox. Stable item IDs are globally unique. Options support bundled icons
and short trailing badges on multiple-selection ListBox source rows only. Selection is controlled and singular;
section expansion is controlled. Disabled, collapsed, removed or replaced items
revoke old proposals. Tab traverses section lists; arrows navigate within a list.
Multi-item/select-all requests are ignored. Section headings remain reachable.

Rows reuse Mac source-list active/inactive colors and native activation delivery.
Fixed light/dark sidebar surfaces match the personal gallery; they do not add OS
appearance/accent discovery, translucency or AppKit NSOutlineView wrappers.
