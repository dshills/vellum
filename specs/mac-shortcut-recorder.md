# macOS shortcut recorder

A reusable, retained `ui.ShortcutRecorder` renders a compact rounded managed
button with macOS modifier glyphs in Control–Option–Shift–Command order. It uses
the existing light/dark Button theme, focus outline and inactive-window style.
Applications retain the component on one owner and accept `ShortcutEvent.Value`
by assigning `Value` in `Change`. The recorder never silently applies proposals.

## Interaction and validation

- Clicking, Return or Space on the focused button starts recording. The visible
  “Type shortcut…” prompt marks readiness. Keyboard chords arriving while the
  activation is still being published are consumed; start typing after the prompt.
- A chord ends recording. `Shortcut.Validate` requires Command, Control or Option
  for printable/named keys; F1–F24 may be unmodified. Shift alone cannot turn an
  ordinary letter into an application shortcut. Modifier-only keys keep recording.
- Canonical key names are uppercase printable ASCII plus Space, Return, Tab,
  Escape, Delete (backspace), ForwardDelete, Left, Right, Up, Down, Home, End,
  PageUp, PageDown and F1–F24. This first version does not record non-ASCII keys,
  multi-stroke sequences or physical-keyboard-layout-independent bindings.
- Plain Escape cancels without a proposal, including a queued start before the
  recording prompt appears. Tab/Shift-Tab cancels and allows normal
  focus traversal. Plain Delete/Backspace clears through the same validation and
  controlled Change path, including when the focused recorder is idle.
- Application `Validate` may reject conflicts, reserved actions, or clearing.
  `Invalid` reports the error; accepted `Value` remains intact. Errors are owned
  and displayed by the application, as demonstrated in the gallery.
- `Shortcut.Label` supplies macOS key glyphs; `Shortcut.Chord` supplies existing
  command-registry/Cogent modifier spelling. Neither method registers a command.
- Nil Change makes the recorder read-only/disabled. `Cancel(session)` only affects
  that session, protecting a later recording from stale cancellation.

## Native ownership and lifecycle

One native recorder binding per managed button publishes immutable owner-held
snapshots and at most one focused binding per native scene. Recording intercepts
focused key chords before window and after-focus application commands, including
conflicting Undo/Save/Quit assignments. Outside recording, commands retain normal
routing. Modifier-only and pending-settlement input cannot escape to commands.

Focus loss, native window deactivation, disablement, tab/disclosure hiding,
dialog-scope retirement, source replacement/removal and window closure retire the
recording capability. Native input carries its originating session and lifetime;
queued stale input cannot invoke a replacement component. Model delivery uses
`postControl`/the originating WindowRef and does not inspect another window’s
native tree or focus. Capture is local to the focused active native window; no
system-wide keyboard monitoring, input-source changes or global registration is
used. A refused cancellation retains one bounded capability for owner cleanup;
existing adapter refusal/error policy is preserved. macOS-reserved shortcuts which the system does not deliver cannot be
recorded. Applications remain responsible for their command registry and policy.

## Demonstration and evidence

The Mac Widgets Shortcuts page exposes Duplicate/Export editors, visible conflict
feedback, disable controls and reset. Assigning the other action’s chord proves
conflict rejection. Command-Q/Command-W are reserved by the sample.

Focused portable tests cover canonical keys, glyphs/chord conversion, controlled
accept/reject/clear behavior, stale sessions, disabled/read-only identity and
validator-driven retirement/restart. Focused offscreen native tests cover actual
managed binding, command precedence, exactly-once proposals, Escape/Tab/Delete,
focus/activation changes, retained modal scene transfers, actual modal keyboard
handlers, hidden/disabled/replaced/closed capabilities and idle
command restoration. These are automated routing/ownership evidence. Physical
keyboard verification and live system-reserved shortcut delivery are separate
native checks; they are not implied by offscreen tests.

Accessible Press uses the same owner-held source/session capability and requests
focus through the existing pending-focus bridge. It never writes native focus
from model delivery. Captured handlers are revoked on component replacement or
modal retirement before they can begin or focus a replacement recorder.
