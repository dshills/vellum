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

Current status: the complete 57-input/21-observation native record passes with
the inherited-state and actual multi-click focus fixes. Six preceding attempts
remain failed/sealed. This is automated native Go routing in the controlled
Go2/SDK786a CPU-scene configuration; physical/OS input and full Mac acceptance
remain open. See the [result](../../../../docs/reports/macos-control-refocus/README.md).

## Native choice catalogue

```sh
go run ./cmd/gallery -choice-native-check
```

This independent mode uses the same ordinary public constructors in a compact
900x700 requested window. It routes 30 inputs and checks 23 observations, then
closes its original window and emits one JSON result. The 90-second deadline,
6-second worker join, nonzero failures and disabled geometry persistence match
the other catalogue mode. These modes and `-exit-after` are mutually exclusive.

Checkbox, Switch and ToggleButton receive pointer/Space input. RadioGroup and
SegmentedControl share selection, skip disabled options and restore rejected
requests. Slider keyboard changes drive progress/status values. Right at its
maximum sends one clamped Change proposal; selected-choice Space sends none.
Final native routing passes with exact callback counts and terminal cleanup.
Progress checks cover semantics/layout, not native pixels; Slider drag and
physical/OS input remain separate. See the [result](../../../../docs/reports/macos-choice-catalogue/README.md).

## Native lists and action menus

```sh
go run ./cmd/gallery -navigation-native-check
```

This separate 900x700 requested view reuses the ordinary public list and two
menu constructors. It records 47 native Go FIFO packets and 41 observations,
then closes the original window and emits one JSON result. Its 90-second
deadline, 6-second join, nonzero failures, disabled geometry persistence and
exclusive mode flags match the other catalogue modes. Policy changes are
owner-posted setup; tested values flow through ordinary public callbacks.

The complete native routing record passes list scroll/focus/rejection/read-only/
disabled behavior and menu duplicate-label/check/separator/dismissal/stale-target
cases with exactly-once callbacks and terminal cleanup. Native pixels, physical/
OS input and human accessibility remain separate. See the [result](../../../../docs/reports/macos-navigation-catalogue/README.md).

## Native structural controls

```sh
go run ./cmd/gallery -structure-native-check
```

This separate bounded 900x700 requested view reuses ordinary Disclosure/SplitPane
constructors and the document-tab view. It authors 50 Go FIFO inputs/42 observations, emits one result and joins its
worker. The intended successful path closes the original window; failure cleanup
retires the session. The first native run remains failed at41 inputs/35 states. It uses the other
catalogue modes' 90-second deadline, 6-second join, exclusive flags and disabled
process-local geometry persistence. Tested state uses ordinary public callbacks;
owner-posted policies are setup only. The ordinary split page now includes enough
workspace rows to exercise actual scrolling. See the
[result and limitations](../../../../docs/reports/macos-structure-catalogue/README.md).

The structural native prefix above predates the subsequent Prism tab activation
and disabled-focus fixes. Those fixes pass actual Core queue/manager routing and
complete driver/gallery race suites, but have not received a new native run.
They retain the failed scope and do not close the full structural/native gate.
