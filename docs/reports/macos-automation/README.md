# Mac semantic automation checkpoint

Vellum now has an explicitly enabled Mac-local semantic endpoint. Applications
register typed queries/actions, provide a revision identity and start a private
Unix socket with a random capability. Ordinary imports/workbench startup open no
endpoint. `-automation` enables the workbench's named domain bindings and emits a
credential descriptor only to its invoking process.

The endpoint serializes domain work through Application.Post, checks expected
revisions, rejects malformed/stale/unknown requests and retains eight combined
transport/owner credits until accepted work skips/returns or its owner closes.
Socket/directory modes are 0600/0700. Response JSON and scalar audit retention are
bounded; callbacks own their allocations/time and explicitly redact their DTOs.
Cancellation is cooperative after handler entry, without rollback/deduplication.
Request contexts retire on connection completion; applications deliberately use
their application/window lifetime when admitting jobs that should outlive an RPC.

Workbench inspection reads selected or stable virtual component IDs without a
render/owner scan. Selection uses the shared validated UI selection function;
rotate/raise/Undo/Redo reuse command availability and document history. A literal
redacted document path accompanies count/availability metadata. Source, selection
and reference epochs prevent stale ABA requests. Epoch exhaustion refuses before
mutation. Native outline group selection retains its prior noncomponent behavior.
These DTOs describe application state, not current native frames or accessibility.

Twenty automation test functions cover real headless Mac sockets and injected
policy errors, including credentials/permissions, strict wire parsing, typed
actions, stale revisions, queue/credit retention, disconnect, half-close/second
message cancellation, window-scoped retirement, panic teardown, redaction, audit/
response limits, constructor/copy refusal and owned path replacement. Two actual
workbench domain tests and an independent public typed consumer also pass. Root
race/vet, the complete offscreen driver suite before the bounded correction, final
affected-workbench race/vet, native workbench compilation and the 28 SDK producer
tests pass. Final root/consumer checks cover the transport correction. Native
compiled binaries are never run by this checkpoint.

The final private SDK verifies all 6904 identities with trusted digest
`40705fd4dc37504741b5d76f69789760d6113618ecc572a0b96b306e11e35f2c`.
The extracted path contains spaces, all four module replacements remain inside
the SDK, and external module verification plus generated consumer race/vet/build
and public API scenarios pass. All 26 final source hashes match the shipped
manifest; all 6172 prepared bytes remain equal to the preceding checkpoint.
Native runtime is explicitly not executed; this is not independent authenticity,
redistribution rights, signing/notarization or full Mac qualification.

Internal V1 review found a client-dial deadline gap, zero-Registry panic and audit
publication after EOF. V2 fixes them with pre-dial cancellation, constructor/self
guards and pre-EOF terminal audit. One Prism run used Anthropic Claude Sonnet 5.5
with complete 25-file/five-chunk coverage and raw exit 1/one medium finding
4086271e3b54a6eb: transient accept errors retired the endpoint. Wrapped resource/
aborted/temporary accept errors now retry with cancellation-aware 5 ms through
1 s capped backoff; permanent errors remain observable through Server.Err.
Independent V3 review accepts that exact fix; V4 approves the final UI-preservation
leaf. No second provider review was run for the bounded corrections.

Initial development receipts remain distinct: workbench/consumer compile fixes,
two new-test attempts copying an active Server mutex, and a resolved manifest
reader schema assumption. Those failures are not final qualification. Copying
active synchronizing values remains unsupported; the final fixture snapshots
shared resources under lock without copying the live mutex. Original failing
source variants for every development receipt are not all frozen.

This checkpoint launches no native UI/input/performance campaign and changes no
clipboard, input source, VoiceOver, display or power preferences. Live checks stay
deferred until the user explicitly says the computer is free. Broader professional
automation/replay, human/native accessibility/input/services, sustained performance,
GPU/DPI/memory/lifetime and distribution acceptance remain open.

[Checkpoint](records/checkpoint.json), [raw evidence](records/evidence.tar.gz)
and [member index](records/archive-index.json) bind this headless source/tooling scope.
