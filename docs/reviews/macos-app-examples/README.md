# Native application checkpoint review

Internal independent reviews preceded the external checkpoint review. The
reviewers covered native chrome/view/layer ownership, zero-policy compatibility,
root clipping, fixed geometry restoration, surface styles, actual text alignment,
model bounds and history, native field settlement, and example layout.

Material findings were corrected before freezing the implementation:

- Native header overlap: each demo now reserves the measured 52Dp band.
- Host cleanup on error: runner defers unwind before the outer executable exits.
- Enum-only alignment: completed glyph pixels and button content now align;
  selection/link geometry uses the same origin. Final positive label growth
  expands actual width while wrapping measurement retains GrowWrap.
- Calculator exponent editing and rational growth: generated numeric tokens are
  atomic on replacement/delete, token/exponent/value budgets bound computation.
- Pending native text before structural/history actions: Notes and Reminders
  serialize public asynchronous settlement, preserve committed Change ordering,
  bound pending actions, and clean up on window lifetime cancellation.
- Reminders changed-draft Redo: refuses nonmutating and retains the redo branch.
- Saved Calculator geometry: finite explicit fixed Dp overrides saved size/max;
  the public Width/Height documentation states this exception.
- Client/fullscreen restoration: AppKit client-view coordinates and cached
  position ownership are corrected without replacing the GLFW content view/layer.
- Intrinsic-width demo content: completed geometry and native captures check the
  layout instead of accepting descriptor growth flags as proof.

Review evidence is retained in the external macos-window-chrome and
macos-app-examples artifact directories. Reviewer execution is distinguished
from root-owned tests/native checks; a source review alone grants no native,
performance, physical pointer, distribution or full Mac acceptance. See the
[checkpoint report](../../reports/macos-app-examples/README.md) for those limits.
