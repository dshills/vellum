# Standard control gallery

From the `drivers/cogent` module after preparing dependencies:

```sh
go run ./cmd/gallery
```

For the bounded native Selector/NumberField integration scenario:

```sh
go run ./cmd/gallery -control-native-check
```

The second command opens a compact view built with the gallery's ordinary public
control constructors. It uses optional `cogent.DesktopSession` from one consumer
outside the app owner. It records actual completed snapshots and 57 Go FIFO
inputs with 21 expected observations, then closes the window and emits one JSON
result. It has a 90-second scenario deadline and 6-second join boundary. A failure
exits nonzero. It checks the current Command+A binding without changing keymaps.

The authored scenario includes popup pointer/key selection, rejected selection, Escape and
outside-dismissal consumption; numeric typed, invalid, corrected, clamped,
stepped and rejected edits; disabled/refused and reenabled controls; and closure
with an open popup. Rejection/enabled flag changes are owner-posted fixture setup.
Input evidence comes from public native routing, not model setters or AX actions.

Use a free desktop. Geometry persistence is disabled for this mode. Go events do
not prove OS/physical input, IME, VoiceOver, display presentation or performance.
See [native integration usage](../../../../docs/desktop-integration.md) and the
[Mac acceptance map](../../../../specs/control-acceptance.md) for scope and results.

Current status: all four original native attempts remain failed. The last reached
24 routed inputs/11 observations before a missing-refocus guard. The corrected
scenario now adds four explicit routed refocus clicks after numeric Enter:
57 inputs/21 unchanged model observations. Its software focus/refocus regression,
gallery race checks and build pass; the revised native record has not run. Full
numeric/disabled continuation and final original-reference closure remain pending.
