# Personal Mac input and accessibility acceptance

Use only the disposable Mac Widgets gallery and editor sample. Do not change
system input sources, VoiceOver, global shortcuts or display settings as part of
automation. Existing automated checks and native event helpers are separate from
human acceptance. Record candidate commit, application, method and actual result.

## Automated contract coverage

- Existing managed-editor tests exercise marked-text preview/commit/cancel,
  exact CRLF/emoji bytes, atomic history and stale/remounted input refusal.
- Existing native AX selector fixtures exercise UTF-16 ranges, surrogate bounds,
  protected entry, current text geometry and stale/disabled refusal.
- Native focus-loss/FIFO tests verify admitted composition settlement before
  refocus, preserving selection and rejecting intervening-field activation.
- The public SDK consumer combines Unicode Find/Replace, active-composition
  refusal, cancellation and one-operation Undo/Redo using exported APIs only.
- Toolbar/range/gradient keyboard alternatives and AX roles have separate routed
  regressions. These are not a screen-reader usability pass.

## Bounded native app check

1. Tab into a field; enter disposable accented/combining/emoji text, move the
   caret, select, copy/paste using only the sample, and switch focus. Confirm text
   survives without duplication. Preserve the user's clipboard when using tools.
2. On Find/Replace use Cmd-F, Cmd-G, Shift-Cmd-G and Escape; Replace All then Undo
   restores the exact original document. Closing a focused native panel returns
   keyboard input to the original live window.
3. Select a Quick Look sample. Space opens, Escape/Space closes normally, and
   Space reopens without clicking the parent. Selection must remain unchanged.
4. Use the range track's Up/Down then Left/Right, and gradient stop keyboard
   alternatives. Numeric values/selection must agree; disabled inputs refuse.
5. Open a managed sheet/popover and a native panel. Verify keyboard cancellation,
   default action where advertised, focus return, and no callback after closure.

## Human IME acceptance — pending until observed

Using an input source the owner already has enabled, compose several characters
in single-line and multiline editors; revise the candidate, commit once, Undo,
Redo, then repeat and cancel. Move focus during composition. Confirm candidate
placement follows the caret and that cancel/focus changes do not lose committed
text or create duplicate history. Record language/provider and actual outcome.
Automation must not enable an input source to manufacture this evidence.

## Human VoiceOver acceptance — pending until observed

With VoiceOver enabled by the owner, navigate fields, menu choices, toolbar
buttons, sheets, document rows, range endpoints and gradient stops. Check label,
role, current value, disabled/selected state, focus return, and ability to perform
the advertised action without pointer input. Read and select multiline text,
including emoji. Record actual usability issues; an AX tree alone is insufficient.

## Hardware boundaries

Moving a window between displays or changing actual scale/contrast/motion settings
requires separate authorized observation. The supplied 160/320-DPI offscreen
routing checks do not qualify physical displays. No Intel/other-macOS acceptance
is inferred from this personal arm64 candidate.
