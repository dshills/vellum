# Mac preview and editing widgets

User-approved expansion: Quick Look, Find/Replace, search scopes, gradient editor,
and range slider. This explicitly supersedes the older internal-app API freeze
for these five reusable capabilities only. Rendering/pacing and release gates stay
unchanged. Each phase has public Go consumption, specs/docs, focused regression
checks, Prism Anthropic Claude Sonnet 5.5 review and a separate commit.

## Quick Look

Optional worker-only QuickLook service; exact live parent; 1–16 stable-ID local
regular files, initial selected ID, bounded titles/paths. Actual QLPreviewPanel,
controlled responder-chain ownership, no foreign singleton takeover. Space on
DocumentList/OutlineView proposes Preview independently of activation/selection.
Close/Space/Escape completes normally; cancellation and parent/app closure revoke
native references. One owned native auxiliary panel at a time. Other platforms
return unavailable. The demo generates temporary text/image samples, never reads
unselected personal files, and deletes samples after preview closes.

## Find/Replace

Reusable compact/regular bar over an existing editor.Model, literal Unicode
search, case/whole-word options, previous/next and counts. Optional replace row;
Replace and Replace All each use the model journal, preserving Undo. Cmd-F opens,
Escape closes, composition refuses committed document changes. No regex engine.

## Search scopes

Controlled Finder-style scope bar with original-color icons, small/regular sizes,
query field, selected scope and counts. Scope IDs stable; disabled choices refuse;
retained state, keyboard arrows and focus via existing controls. No filesystem or
network search; application supplies matching data.

## Gradient well

Controlled 2–16 stable-ID color stops, preview over checkerboard, linear angle,
editable stop position/color/opacity, pointer stop dragging, keyboard/numeric
alternatives, add/remove and native Colors option. Stop edits merge against current
accepted state; stale/replaced/disabled gestures refuse. Full-color pixels remain
untinted. Apps own undo/persistence. sRGB byte interpolation, no color management.

## Range slider

Two handles on one track, ordered lower/upper values, bounds/step, paired numeric
fields, small/regular sizes, pointer capture and keyboard/accessibility alternatives.
Handles clamp at each other; live proposals, no trailing click on release, cleanup
on loss/removal/disable/shutdown. No new renderer or application history ownership.

### Native verification follow-up

Quick Look requires a visible main/key parent in the active application. The
shared panel acquires its responder controller before presentation and owns its
geometry; it is not reparented as a child window. Relinquishing preview control
terminates the session. Find Escape is a window shortcut while the Find page and
bar are active, so a focused native field cannot swallow it as its Abort key.
