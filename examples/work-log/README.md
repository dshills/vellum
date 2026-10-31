# Work Log: an exported-SDK consumer

A small standalone notebook application for dated work entries: record what you
changed, what you learned and the next task. It uses the document starter's
editing, Find/Replace, Undo/Redo, Save/Open/Save As and recovery behavior. No
framework control or public API is added. Initial example notes are retained;
create your own dated entry with New Note. Delete examples only if desired.

Generate from an exported SDK with a separately checked SHA-256:

```sh
python3 -B examples/work-log/create.py --sdk /absolute/Vellum-SDK.tar.gz \
  --sha256 EXPECTED_SHA256 --output /private/tmp/my-new-work-log
```

The new output holds the verified extracted SDK, `Work Log/` application sources
and CONSUMER.json provenance. It refuses existing outputs and outputs within this
checkout. Harmless path aliases are resolved before the containment check. Failed
generation preserves its partial output for diagnosis; choose a fresh output
for retry. The generated Go module replaces Vellum and native dependencies with
paths inside that extracted SDK; no original checkout module is used.

In the generated Work Log directory, source the extracted SDK's
`drivers/cogent/env.sh`, then use the Go version declared by that SDK. Remove
an ambient GOROOT if it points to another Go installation; the selected Go binary
should determine its own GOROOT:

```sh
unset GOROOT
go test -race -tags offscreen ./...
go vet -tags offscreen ./...
go build -trimpath -o ../work-log .
```

Package with that SDK's scripts/package-app.py using name `Vellum Work Log`,
identifier `com.dshills.vellum.worklog.internal`, version `0.1.0` and ad-hoc sign
identity `-`. This is a private consumer, not a native Vellum release asset.

The generated module's stable identity selects separate Application Support
storage. For synthetic acceptance use `-data-dir /absolute/private/test-profile`;
never point it at Notes, Documents or another app's repository. Keep the SDK
alongside generated source, or deliberately update all replacements when moving.

Acceptance: create a dated entry, edit Unicode text, Undo/Redo, save, Quit,
reopen and confirm exact content; Open the saved notebook and confirm selection.
Record source, SDK hash, executable/package identity and any limitations. No
multi-day use, human IME or VoiceOver claim follows from a short smoke test.
