# Native titlebar customization: personal app check

Implementation checkpoint: `a7354ea`, packaged version18, macOS26. Focused race
and vet receipts are kept with the private app delivery, separately from these
native UI observations. No human physical-pointer/VoiceOver qualification or
release performance campaign is claimed.

## Verified in the packaged application

- The content-area Customize Titlebar button opens AppKit's actual palette.
  The native item's context-menu Customize Toolbar entry opens that same palette.
- The authorized native mouse helper moved Save from third to first, removed Open
  by dragging out, and re-added it by dragging the original full-color palette
  item back in. Customization did not increment the sample command counter.
- Palette Done restored command admission: Save then incremented once.
- A fixed space was added through the native palette's accessible insertion
  action. The process was quit/relaunched. The new process retained Save-first,
  omitted Open and the identical leading spacer geometry. Counter reset to zero.
- All native actions were removed through AppKit's accessible Remove actions.
  The content-area button still opened the palette. Dragging the native default
  set back into the toolbar restored all six actions and original spacing.
- Dark appearance independently omitted New, quit/relaunched and retained that
  omission. The light native toolbar retained its own default set. Dark New was
  re-added through the palette's accessible insertion action, then ran once after
  palette exit. Dark palette artwork retained color and used the native dark panel.
- After removal/reinsertion/restoration, narrowing light from668 to307 points
  exposed native overflow with enabled Redo and disabled Delete. Redo reached the
  owner once. Width was restored; the active toolbar screenshot retained full color.
- A managed content customization dialog disabled every native command and
  removed reorder availability. Done restored five enabled actions and disabled
  Delete without changing the action counter.

Both appearance settings were restored to their default item sets for handoff.
No content-toolbar JSON preferences, document data or global OS settings changed.

## Observed limits

The native macOS26 compact context menu offers Remove Item and Customize Toolbar;
its palette does not offer a display-mode picker. This implementation retains
that native icons-only behavior. The content toolbar display variants remain
independent. Live persistence evidence covers order, omission and spaces; native
hidden-toolbar recovery and alternate display attributes are not separately
qualified by an OS display/visibility picker here. Empty-action-toolbar recovery
was tested directly.

Customizable attachment requires macOS15+ for documented palette KVO and still
supports one registered standard-chrome window opening. AppKit owns the saved
configuration format; no custom settings schema or native window delegate was
introduced. Tests for queued stale actions, idempotent samples, managed modal
precedence and retirement are deterministic owner-routing evidence.
