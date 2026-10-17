# Native workbench external semantic automation

The independent public Go client passes two native workbench cases from clean
source `781f847a2d43033d3869d7fdd2f6713656b8a803`: 100,000 and one million component
records, with 40 serial semantic requests each. Both workbenches close normally;
the client requires endpoint removal and post-close refusal. The outer executor
observes the inherited controller/application groups absent. This advances the
native application API portion of P10; full macOS acceptance remains open.

## What the client checks

The runnable [example](../../../examples/automation/README.md) consumes only
public `automation.Call` and the standard library. It uses the workbench's actual
allowlisted bindings, shared command registry and sparse history. Its queries
check selected component 17 and the last virtual record, canonical identities,
component DTO fields, redacted document path, record/reference counts and four
command availability flags. Typed wrong/unknown params, noncanonical component
identity, unavailable Undo and forbidden `file.save` refuse.

The native sequence selects, rotates and raises a component, checks unchanged
state after a stale invocation, Undoes both edits, Redoes both and rejects old
revisions after both domain/history and selection ABA changes. Finally it Undoes
the edits and checks the original component DTO and empty Undo availability.
Invokes are never retried after a transport error. Forty terminal audit records
must match the responses' IDs, operations, targets and codes.

The dedicated workbench diagnostic requires `-automation -automation-check` and
a 5–30-second exit duration, excluding other smoke/endurance modes. Each live
case uses 20 seconds. Readiness requires a native window and a successful
nonempty public `GPUCanvas.Snapshot`; it asserts no pixels, GPU backend,
presentation or timing. Signal cancellation is installed before endpoint/worker
creation. The readiness result and worker retirement are separately joined;
transport Close/Done/Err is joined even on readiness/descriptor-write failure.
Process-local window geometry persistence is disabled for the entire fixture.

| Component records | Serial semantic requests | Client result | Exit and owned group |
|---|---|---|---|
| 100,000 | 40 | PASS | Zero exit; absent |
| 1,000,000 | 40 | PASS | Zero exit; absent |

These are two native application executions plus their two direct Go controller
processes. The client launches its native child in the controller's inherited
process group. The outer executor owns that group, bounds output and time,
independently attempts retirement/reap and observes absence. The shared live
lease remains held through cleanup; source/prepared/binary/review pins matched and both checkouts were clean at the
recorded admission and terminal snapshots. The native root is sealed afterward.
No physical input, input-source/VoiceOver/clipboard or OS preference changes are
generated. The elapsed case durations include the deliberate 20-second lifetime;
they are not performance or responsiveness measurements.

## Evidence and privacy boundary

The client keeps one canonical bearer descriptor only in bounded memory. It
discards **all** raw native stdout/stderr, including on malformed/missing or
duplicate descriptors. No bearer goes into argv, environment, files, shared logs
or application DTOs. The retained controller output is one fixed, checked scalar
summary and empty stderr per case. Native readiness, per-call DTO comparisons,
audit matching and endpoint teardown are assertions in the exact reviewed client
and workbench sources; the raw native events and per-call DTOs are deliberately
not retained for independent offline replay. This report reparses scalar
summaries and reconciles their command/ownership/log hashes. It does not present
them as saved raw DTO validation or physical UI evidence.

Post-close probing uses a fresh bounded context and requires a missing-socket
error; main-run cancellation cannot produce success. Local failure teardown
independently attempts TERM, bounded wait, Kill and bounded final join, preserving
fixed safe errors. The outer owner additionally verifies group absence. Normal
endpoint cleanup is proven through the client assertion; SIGKILL/OS failure has
no graceful socket-cleanup guarantee. The same OS account is not sandboxed by
this bearer capability.

The outer producer labels terminal reporting as pending until its exclusive
final receipt is written. Seal, reporting and receipt failures have separate
controls; a reporting failure from an otherwise successful state is saved as
failure. The final receipt and successful process termination are required
alongside retained controller logs.

## Source checks and review

The implementation checkpoint passed eight installed checks: complete workbench
command offscreen race/vet, public automation race, independent client race/vet/
build, native workbench compile-only and example-shell syntax. Headless client
regressions use an independent model behind the real public Registry, owner queue,
Unix endpoint and Call. Six deliberately broken redaction, availability, virtual
identity, Undo, Redo and ABA behaviors refuse for their specific expected reasons.
Framing, stream limits, duplicate keys, missing/wrong markers, audit mismatch,
no invoke retry, post-close cancellation and child-cleanup failures are covered.
Twenty-three additional outer parser/finalization controls precede native launch.
These automated/non-GUI checks are distinct from the two native executions.

Internal source reviews resolved cleanup, worker retirement, inspect-revision,
early signal, cancellation and finalization issues. One Prism Anthropic Claude
Sonnet 5.5 invocation reviewed the exact 61,120-byte, 14-file source/docs stage:
complete coverage, one chunk/call, no cache, exclusions or truncation, and no
reported findings. Source commit and clean executable metadata identify the
runtime bytes. The subsequent evidence-only checkpoint adds no provider call.
See [structured results](result.json), [bundle manifest](evidence.manifest.json)
and [retained evidence](evidence.tar.gz). Imported executor, lease,
process/environment, Git and Prism-retirement helper bytes and their historical
non-native controls are retained under `provenance/`; the origin index binds each
original path/hash to its archive member. Host executables are fingerprinted and
omitted. These historical helper controls are not additional native executions.

Physical pointer/key/menu/drag/capture behavior, external AX/human IME/VoiceOver,
chooser/clipboard interoperability, GPU execution/presentation, ordinary frame
budgets, retained/native memory, endurance, distribution rights/source-build
closure, signing/notarization and public release remain separate. This short
semantic campaign does not complete P9/P10 professional workflows or P4/P6/P9/P11
and full Mac qualification. The previously tested private SDK archive still
identifies source `2010a33`; it is not relabelled as containing this new example.
