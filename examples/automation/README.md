# External native workbench automation client

This independent Go module uses only Vellum's public `automation.Call` and the
standard library. It starts the built native workbench and checks application
semantics through its explicitly enabled, private macOS endpoint. No native
adapter or third-party module is imported by the client.

Build the native application and run the client from this checkout or a prepared
private SDK on macOS:

```sh
(cd drivers/cogent && ./prepare.sh && . ./env.sh && go build -o build/vellum-workbench ./cmd/workbench)
(cd examples/automation && go run . -workbench "$(cd ../../drivers/cogent/build && pwd)/vellum-workbench" -rows=100000)
```

Use `-rows=1000000` for the million-record case. The client opens one visible native
window for about 20 seconds. Run it when the desktop is available for testing.
The application fixture disables window geometry persistence in its own process;
it does not change input sources, VoiceOver or system preferences. Edits stay in
memory, and the example exercises no file chooser, clipboard or save command.

The child is invoked with `-automation -automation-check -exit-after=20s`.
`-automation-check` requires a separate process with `-automation`, an exit duration
of 5–30 seconds and no other smoke/endurance mode. It waits for a native window
and a successful nonempty public GPU snapshot before emitting readiness. The
fixture's signal context is installed before endpoint/worker creation. On normal
exit or handled termination it separately joins the readiness worker and the
endpoint transport before reporting closure. A failed capability-descriptor write
also closes and joins the endpoint before the process exits.

The client keeps the bearer descriptor in bounded memory. It discards all raw
child stdout/stderr, including on malformed, missing or duplicate descriptors;
only its own checked scalar summary is published. Never tee the workbench's
`-automation` stdout into shared logs. The endpoint capability is not a sandbox
against the same OS account. The client inherits its child's process group from
its own caller, allowing an external test runner to retire the entire owned
group. Local failures attempt direct-child TERM, bounded wait, Kill and a bounded
final join; an unproven join is a failure, not a successful cleanup claim.

Forty serial requests cover stable component identities, the last virtual row,
selected-object DTOs, redacted document paths, typed parameter refusal, command
availability, rotate/raise, sparse Undo/Redo, unchanged state after rejection,
and stale revisions after edits and domain/selection ABA changes. It never
retries an invoke after a transport error. Every response must match the native
server's bounded terminal audit; zero exit, closure marker, socket/directory
removal and a refused post-close public call are required for success.

```sh
cd examples/automation
go test -race ./...
go vet ./...
```

These non-GUI tests use an independent in-memory application behind the real
public owner queue, Registry, Unix endpoint and Call. Six deliberate domain,
history, revision, redaction and availability faults must fail for their specific
reasons. Other controls cover fragmented/private framing, missing/wrong markers,
stream limits, audit mismatch, no invoke retry and cleanup failure paths.

This is native application **semantic API** evidence. It does not synthesize or
verify physical pointer/keyboard input, native AX geometry, menu focus, drag/drop,
IME, display scaling, presented-frame timing or retained/native memory. Those
acceptance checks remain separate. The endpoint implementation is macOS-only;
this example refuses other platforms. No cross-window UI interaction is tested.
