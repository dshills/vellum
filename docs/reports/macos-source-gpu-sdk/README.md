# Private source-built GPU SDK overlay

The SDK builder now accepts an explicit ARM64 macOS static GPU library and packages
it with owned Go linkage, exact-byte provenance and local module replacements.
Generated applications can consume that capsule after the SDK moves to a directory
containing spaces. The implementation is opt-in and private; `--public` refuses
this overlay even when the other public gates are approved.

The interface and its limits are in [the SDK specification](../../../specs/sdk.md).
The source-built dependency's earlier GPU and workbench checks are recorded in
[the source-build report](../macos-source-built-gpu/README.md) and
[the workbench consumer report](../macos-source-gpu-framework/README.md).
Those earlier native checks do not qualify this SDK archive or a generated application.

## Implemented and checked

- Three coupled packaging options select the library, expected SHA-256 and JSON
  provenance. Bounded reads retain the exact bytes; malformed archives, foreign or
  non-object headers, duplicate keys and nonfinite JSON numbers refuse. The object
  check observes thin ARM64 Mach-O headers and does not validate full Mach-O semantics.
- The four-file capsule is recorded in the SDK manifest. Eleven driver/native-example
  module overlays disclose source and archived hashes. The root and portable modules
  retain their source manifests. Generated apps receive an explicit replacement
  because Go does not inherit replacements from dependency modules.
- Generation and checking reject capsule/hash/overlay drift, orphan or dangling
  capsules, symlinks and unexpected files or directories. The candidate checker
  requires a native ARM64 Mac before executing SDK code.
- All 68 tests pass across the GPU overlay, SDK builder/checker and archive parser.
  Fixtures cover all-or-none/public refusal, real relocated Go-module resolution,
  wrong-origin rejection, tar integrity, malformed/foreign library payloads,
  provenance overflow, host refusal and capsule layout/integrity boundaries.
- A real private SDK packages 7,436 file hashes and 89 Go inventory modules. Its
  library contains 541 observed ARM64 object headers. Relocation verifies every
  manifest hash, five dependency replacements inside the moved SDK, generated
  consumer race/vet/build and public API scenario race/vet. No native app was
  executed in these SDK checks.

## Artifact and evidence identity

The tested SDK is a precommit private snapshot based on framework
`a274e8aa9305f4b458b2d263e2a1c3d3d7b6a623`, with `source.clean: false` and the
explicit overlay. It is not an archive built from the subsequent clean implementation
commit. The archive is 29,397,418 bytes with SHA-256
`1777dc32f97e12056b2931f5d6d21a114a0ad191a1415c35d260b414c0d3c05d`.
The static library is 37,401,760 bytes with SHA-256
`786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.

[Structured results](result.json), [the evidence index](evidence.manifest.json)
and [the compressed evidence bundle](evidence.tar.xz) retain source copies,
regressions, exact input inventories, owned-process logs, review records and failures.
The SDK archive and static library are fingerprinted exclusions rather than bundled
binaries. The manifest verifies captured bytes against caller-supplied expectations;
it does not independently authenticate their author or origin.

## Failures and review

An initial regression exposed a single-line `require` statement omitted by the
module selector; the selector was fixed. Internal review then found JSON exponent
overflow and unrecorded physical capsule acceptance. Both were fixed with regression
coverage, and bounded directory inspection was adopted.

The first real relocation check passed generated-consumer race/vet/build, then
failed because the harness denied the local Unix socket required by an API test.
That failure remains recorded. A revised test policy passed one Unix-socket
round-trip and eight direct IPv4/IPv6 TCP/UDP refusal controls before the fresh
relocation check passed. These controls do not establish an air gap or exclude
indirect networking through a system daemon. Cocoa deprecation and duplicate
`-lobjc` linker warnings remain in the passing logs.

One fresh Prism review with Anthropic Claude Sonnet 5.5 covered the exact eight
source/specification files (41,126 diff bytes), with no exclusions or truncation.
It returned three retained medium findings and CLI exit 1; no findings were
discarded or suppressed. Independent internal review refuted each on the actual
production path: the framing parser already normalizes index names; framing and
matching captured-byte SHA-256 checks precede the object walk; and valid default
manifests without a capsule already return no candidate. Invalid or oversized
manifest refusal is intentional. A focused 14-case probe corroborates these paths.
The raw provider report and per-finding dispositions are retained. No second
provider call or production source edit followed this review.

The generated report and evidence packaging receive internal review after the
source checkpoint review. The source review does not claim coverage of later
report files or prove runtime behavior.

## Remaining scope

Inspect a retained SDK consumer's link map and dynamic dependencies, then execute
that exact application through an independently reviewed native check. Build and
qualify a fresh SDK from a clean implementation commit. Native-library, Rust,
standard-library, font and other component rights/notice closure remain separate;
additional provenance fields are caller declarations, not attestations. No global
module installation, OS setting change, signing, notarization or public release
was performed. Overall Mac performance, memory, endurance, physical-input,
accessibility and human IME/VoiceOver gates remain open.
