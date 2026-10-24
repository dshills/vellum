# Mac sheet

Owner-held ui.Sheet composes a modal Dialog card, retained children, title and
explicit Cancel/primary actions. It has no move handle, generic Close button or
resizing affordance. The parent window is blocked by the existing modal scene;
other native windows remain independent. Escape follows existing modal dismissal.
Unhandled Return invokes the current enabled default child; text editing retains
precedence. Accept confirmation/cancellation by setting Open false. Confirmation
can be rejected while leaving the sheet open for validation. Closed/disabled or
replaced component identities revoke prior action closures. Native default-action
callbacks additionally retain the exact modal scope to reject queued old sessions.
ReturnFocusID optionally restores an enabled visible Button in the parent scene.

The surface uses app-local Mac light/dark styles, existing scene composition and
accessibility routing. It is a same-native-window managed modal, not NSWindow's
beginSheet API, does not animate attachment and does not implement nested-sheet
stacking policy beyond existing Dialog behavior. No rendering/pacing/release-gate
changes. Routed software checks are separate from final packaged live checks.

Render and mutate a Sheet only on its application owner. Concurrent View calls or
cross-owner/window sharing are unsupported, as with the other retained composite
widgets. Disabled/nil-Confirm default buttons are valid and inert; validation must
allow a temporarily unavailable action while input is being corrected.
