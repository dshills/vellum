# Clean private SDK native consumer

A private SDK built from clean framework commit
`3da753ed4e49ef18ca5ef2377c171efb970b46a1` is consumed by an independently
generated application after relocation into paths containing spaces. The retained
build links the exact packaged ARM64 GPU archive and has no WGPU dylib dependency.
The exact retained app then passes a separately admitted native smoke: verified
rectangle pixel, three GPU render/readbacks and normal owned closure.

The [SDK interface](../../../specs/sdk.md#private-arm64-mac-gpu-candidate) and
[implementation checkpoint](../macos-source-gpu-sdk/README.md) describe the explicit
private overlay. This check adds clean committed SDK, artifact/linkage and native
consumer evidence; it does not replace the earlier dirty precommit archive identity.

## Build and origin evidence

The SDK is 29,397,406 bytes with SHA-256
`82f63da6eb5023da88e531b5766676ed94367c043a4340d7e459e27d26e15600`.
Its manifest records 7,436 file hashes, `source.clean: true`, the framework commit
above and the explicitly disclosed private module overlays. All hashes reconcile
after extraction. All 7,436 payload hashes also exactly match the previously
race/vet/build/API-tested precommit SDK; only source and untracked-file manifest
metadata differ. That comparison does not claim a rerun of the earlier checks. The generated app has its own six starter files and module;
five replacement origins resolve inside the relocated SDK. Build discovery records
486 packages and fingerprints their selected Go/cgo/embedded inputs. Existing
admitted input hashes cannot be replaced during discovery.

The app is a thin ARM64 `MH_EXECUTE` binary built with Go 1.27.2. Its observed
minimum macOS is 13.0. The 60,180,818-byte executable has SHA-256
`0323d08b9c0a5ffce4cea6e9efe422d5fe6aed832e6ebde700197a684d398c17`.
The same owned compiler invocation produced its retained 12,994,659-byte link map,
SHA-256 `c711459eaf061deb466aeb0ff6472b153013d047d9553cf6cec7e858294b8004`.
The map resolves 141 entries to exact members of the SDK's static library,
SHA-256 `786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.
The binary embeds the SDK GPU-module replacement, and its inspected dynamic
libraries contain no WGPU library dependency. These are observed build and linkage
facts; matching hashes do not authenticate the build author or establish rights.

All eight build/generation/inspection commands finish successfully with complete
bounded logs and owned process groups recorded absent. The compiler emits 1,113
bytes of retained warnings: two Cocoa accessibility deprecations and duplicate
`-lobjc`. This is not a warning-free build.

## Native check

The retained executable runs once with `--benchmark --idle=1s --samples=3`.
It exits 0 and its owned process group is absent. The fixture emits exactly one
243-byte ready record and no stderr, with geometry persistence disabled and
a 630 × 344 GPU image. Startup from application entry to verified GPU readiness
is 179,153,834 ns. The three readbacks take 1,754,000, 1,761,541 and 1,430,542 ns.
The complete owned invocation, including harness work, is 4.389 seconds. The
reported Go heap is 62,731,672 bytes; this is not a native/GPU memory measurement.
All admitted file hashes and the exact SDK member set remain stable through closure.

The fixture disables window-geometry persistence for its process, opens the normal
native app, waits for an initial GPU Snapshot, and checks only the permissive center-pixel
predicate `G >= 100, R <= 150, B >= 80`. Three further Snapshot calls must
succeed. It emits one ready record with their synchronous render/readback
samples; no image-wide pixel equality is asserted. It then closes
after a one-second quiet interval. Its verified GPU image is evidence of that
source path, not proof of compositor presentation, physical input or ordinary
frame latency. The sample durations are host observations without a performance
qualification threshold.

## Review and retained evidence

Independent internal review covers the fixed build plan, retained build results,
separate native admission and final results. Twenty-seven guard/finalization
controls pass without launching a child or GPU. Review corrections reject dangling
output symlinks and prevent source discovery from rebasing previously admitted
hashes. Both preexecution plan amendments are retained. A shared lease, global and
local stop markers, fresh output paths, input checks at command boundaries,
watchdogs, bounded streams and owned process retirement constrain the run.

The native test policy allows local Unix IPC and refuses direct IP sockets. The
preceding checkpoint's one Unix round-trip and eight IPv4/IPv6 TCP/UDP refusal
controls support that policy; this is not an air-gap claim or a claim about
indirect networking through system daemons.

[Structured results](result.json), [the evidence index](evidence.manifest.json)
and [the compressed evidence](evidence.tar.xz) preserve the controller and controls,
source inventories, actual generated starter files, SDK metadata, build and native
receipts/logs, link map, selected static-member identities and internal reviews.
The SDK archive, executable and full static library are fingerprinted exclusions
from the report bundle; the local workspace checkpoint retains the principal
artifacts separately. Go caches, temporary compiler files and system tool bytes
are omitted. No new Prism call was made for this evidence-only checkpoint. The
preceding implementation retains its separate Anthropic Claude Sonnet 5.5 review
and raw findings/dispositions.

## Remaining qualification

This is private compatibility evidence for one native starter on this Mac.
Professional and broader UI workloads, endurance, native/GPU memory, physical
pointer/keyboard/displays, human IME/VoiceOver and interoperability acceptance
remain separate. The ordinary frame targets remain unqualified. Native/Rust,
standard-library, font and other component/rights closure is still pending;
caller provenance is a declaration. No global module installation, input-source
or VoiceOver change, developer-identity signing, notarization or public release
was performed. Overall Mac framework acceptance remains open.
