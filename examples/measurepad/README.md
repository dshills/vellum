# Measure Pad

A small local measurement notebook grown from `scripts/new-app.py`. It uses only
public Vellum application/UI, quantity, history, document and driver APIs. There
are no direct Cogent Core imports and no framework additions.

Prepare native sources and source drivers/cogent/env.sh, then:

```sh
(cd examples/measurepad && go run .)
```

Give a measurement a name and length, choose default units, and click Add.
Explicit units override the selector: `2 in`, `25 mm`, or `2 * 25 mm`.
Rows show millimeters and inches. Select a row to delete it. Undo/Redo restores
measurements; focused text controls receive editing shortcuts first.

Save/Open use native file panels and versioned `.vmeasure` JSON files. Save before
opening another notebook: Open refuses dirty documents rather than discarding
changes. Closing a dirty document offers Cancel or Discard and Close. Cancel
returns to the notebook so you can Save. File operations temporarily disable
application actions. Canceling a pending file operation may require closing again
once completion retires. No file is written automatically and no recovery/autosave
or cloud service is implied. Keep important measurements explicitly saved.

Limits: 100 measurements, 128 bytes per name, positive finite lengths up to one
million mm, 64 KiB files, 100 history entries. It is a utility example, not a
precision metrology or CAD system. Human IME/VoiceOver remain unqualified.

Checks: `go test -race -tags offscreen ./...` and `go vet -tags offscreen ./...`.
Tests exercise conversion, invalid admission, identity/history and actual atomic
file round trips. Native appearance/input observations are separate evidence.

The starter exercise found that the existing public APIs suffice. Applications
still wire document prompts, command readiness and status explicitly. The starter
stays minimal; its independent model/test templates avoid diagnostic dependencies.
