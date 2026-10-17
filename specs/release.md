# SDK versioning, compatibility and release artifacts

`release/version.json` and the public `vellum.Version` constant identify
`v0.1.0-alpha.2`. This is an alpha packaging/API version; no Git tag or remote
release is created by these tools. The owner selected `github.com/dshills/vellum`
and the MIT license for Vellum's own code. The native module is
`github.com/dshills/vellum/drivers/cogent`; example module paths also match their
repository directories. Root and module-local LICENSE files record this choice.
The third-party/native distribution audit remains pending. The Go module graph notice inventory does not establish
complete embedded Rust/native binary/font redistribution rights.

Release metadata also records the experimental, explicitly opted-in Mac semantic
automation protocol version. It does not change the alpha SDK version or target
qualification status; source revision identifies this added public contract.

## Mac distribution input inventory

`scripts/distribution-inventory.py --output <new.json>` records the actual SDK
source-file inventory, verified Go module notice graph and the local macOS
workbench's selected Go package inputs. It creates no native window, executes no
application and changes no clipboard, input-source or OS preferences. Collection
requires Darwin with cgo enabled and uses the same controlled child environment,
graph download/verification and command deadlines as the SDK producer.

Every SDK source file is size/SHA256 recorded, including prepared dependencies.
Resolved `EmbedFiles` are recorded separately with package/module identity and
font classification. Listed cgo source files, linker flags and framework
references are retained. Module-owned conventional native library filenames in
declared `-L` search directories are hashed as candidates for each `-l` reference.
Direct module-owned native file arguments are also hashed as candidates. Missing
or external native files/search paths, `-F` framework searches and unparsed
options (including `-Wl,` forms) produce explicit unresolved-input records.
Framework names retain unverified resolution and do not imply system ownership.
That observation does not establish which file the final linker selected, dynamic
dependency closure or a prebuilt archive's embedded Rust/C component provenance.
Installed Go declaration candidates are recorded from GOROOT and its parent;
they do not establish a complete toolchain component audit.

Nearby license/notice paths are declaration candidates, not evidence that their
terms apply to an asset. All embedded asset/native rights remain review-required,
and the distribution audit remains pending. The JSON neither approves rights nor
updates release gates. Module notice records preserve existing pinned version/
checksum provenance. The pin manifest and every shipped source-file hash bind
the observed prepared bytes; collection does not regenerate private replacements.

Hashes stream with a 1 GiB per-file bound. SDK and build-asset/cgo source lists
each have a 20,000-entry admission bound, checked before surplus file hashing.
Notice collection uses bounded streaming metadata mode and retains no notice
contents; the SDK producer's existing notice-copy mode is separate.
Sources must be regular files without traversal
or source symlinks. Byte identities are reread, SDK source names are re-enumerated
and selected Go build metadata is resolved again before return; observed drift
refuses the report. Go cache/staleness flags are excluded from that comparison.
These sampled checks are not filesystem isolation or an atomic source snapshot.

Output must be new and outside the source checkout. Publication uses the existing
exclusive hardlink/copy mechanism and its completion/checksum semantics.
`--include-untracked` explicitly records whitelisted private precommit files;
otherwise source names come from tracked Git entries. Source revision and
cleanliness describe capture, not a later commit. Equivalent input bytes,
configuration and source state produce equivalent reports across relocated
module/cache paths for the recorded conventional linker paths. Unknown absolute
linker flags remain literal evidence rather than being guessed or resolved.

## Recorded native release assertion

The [16 October release identity check](../docs/reports/macos-native-release-identity/README.md)
authenticates the exact retained `wgpu-macos-aarch64-release.zip` for
`gfx-rs/wgpu-native` tag `v29.0.1.1`, including its URI commit subject and release
repository/owner/release IDs. It uses the pinned official CLI/Sigstore policy,
signed timestamp and normal TUF root/metadata/target checks with captured public
responses and fresh caches. The verifier has no network fallback. All 27 negative
controls pass against the final reviewed source; previous failures and the
three-finding Prism review with dispositions are retained separately.

The prior byte comparison relates the matching local static-library member to
that same authenticated ZIP. This is not an executed-build/compiler/features
attestation, a reproduced source build, complete component/rights approval or
authentication of a Vellum SDK. Existing distribution gates remain unchanged.

## Private source-built GPU consumer

The [pinned-source GPU consumer checkpoint](../docs/reports/macos-source-built-gpu/README.md)
builds the exact WGPU source/header commits with recorded Rust/compiler/features,
then passes private static-only Go consumers and nine headless GPU roots with
23 subtests. Two Rust stripping warnings and the failed initial consumer attempt
remain preserved. The candidate is not adopted into installed libraries or the
SDK; complete component/rights closure, broader native UI/presentation, older
macOS runtime, performance/memory/endurance and full Mac release remain open.

## Alpha.2 window lifetime migration

Alpha.2 deliberately breaks the alpha.1 string-scoped application APIs to avoid
retaining every historical window name. Keep `OpenWindow` receipts, pass exact
`WindowRef` values to close/post/dirty APIs, and split application-wide completion
reservations from `ReserveWindowCompletion`. Closed logical names may be reopened;
old references and controllers remain retired. See [migration and limits](window-lifetimes.md).
This version change does not certify native runtime, performance or distribution.

## API compatibility

The root package identifies the version; focused `app`, `ui`, `command`, `tasks`,
`workbench`, `canvas` and `virtual` packages own framework contracts. The native
Cogent driver remains a separate module and an experimental implementation.
Alpha exported APIs may change. Changes require updated specifications, consumer
scenarios and release metadata; compatible stable-release promises start only
after that contract is deliberately adopted. A source revision/dirty status in a
private archive distinguishes builds sharing the same alpha version.

The independent `examples/sdk-contract` module exercises public controlled views,
style omission versus explicit zero, command availability/shortcuts, window-scoped
worker completion and docked/floating session snapshots. It uses ordinary module
replacement declarations and `GOWORK=off`, without reaching into implementation
internals. Root tests compare `vellum.Version` to release metadata.

## Deterministic private SDK

`scripts/release-sdk.py --output <new.tar.gz>` packages whitelisted framework
sources (including root Go APIs), driver patches/bootstrap, examples, developer
scripts and selected docs/specs, plus the complete prepared Cogent/typesetting
source trees. It obtains source names from Git; private review builds include
nonignored untracked whitelisted files only with explicit `--include-untracked`;
default private builds package tracked names only. The manifest records every
shipped untracked source name and dirty source status. Public
artifacts include committed tracked source only and reject the private untracked
opt-in. Public gating also rejects `assume-unchanged`/`skip-worktree` index flags,
which could otherwise hide changed working-tree bytes from Git status. Build outputs and review/research
artifacts are omitted. Files and dependency paths cannot contain symlinks.

Notice collection retains conventional `OFL`/`OFL.txt` font declarations alongside
LICENSE/COPYING/NOTICE files. OFL entries remain supplemental asset notices;
they do not satisfy a Go module's own-license gate. Owned native patch 0063
restores the exact Noto Sans v2.015 upstream OFL text beside the pinned subset
fonts without changing their bytes or embed pattern. Its source revision, blob
and SHA256 are recorded in the patch README. Collection does not certify the
historical font-subsetting provenance or approve the distribution audit.

Notice collection first downloads and verifies the external module graph in a
temporary copy of the driver module files, with absolute owned replacements.
Cold and warm caches therefore collect the same available declarations; missing
source directories fail inventory creation. Canonical go.mod/go.sum files stay
unchanged. A downloaded module without its own license declaration still requires
manual review, and the public distribution audit remains separately gated.

Archive order, file ownership and timestamps are normalized. Every regular file
is SHA-256 recorded, including copied third-party notices, the inventory and the
private-build marker. Only `SDK_MANIFEST.json` excludes itself. The manifest
records release metadata, Git revision, source cleanliness and private/public
scope. Archive creation publishes an exclusive same-filesystem link when supported,
with an exclusive-copy fallback for unsupported hardlinks. That fallback is not
atomic: the output is complete only after successful return, and a process crash
can leave a partial file. Producer checksum/JSON evidence is emitted only after
completion; readers must wait for that success and verify integrity. Failed copies remove
only their owned partial outputs; existing files and dangling output symlinks
are never overwritten. Output paths must be outside the source checkout, including
paths through symlinked ancestors, so pending archives cannot enter the source inventory.
SDK archive output permissions are normalized to `0644`.

The archive contains prepared source replacements; external Go modules and native
libraries still come from their pinned module graph/cache/network. It is a
relocatable development SDK, not an offline redistributable runtime installer.
Reproducibility means equal source/notice bytes, executable modes, release
metadata and recorded Git source state produce the same normalized archive; a shared alpha version alone cannot imply identical
source or dependency inventory coverage.

## Verification and relocated consumers

`scripts/check-sdk.py --archive <tar.gz> --sha256 <trusted digest>` verifies exact
manifest/file coverage and member checksums, rejects duplicate/traversal/link/
special members, and bounds compressed input to 512 MiB, individual members to
64 MiB, expanded payload to 1 GiB, physical/decoded headers to 20,000, tar metadata
to 1 MiB and the manifest to 8 MiB. The root go.mod is bounded to 1 MiB and
its declared module must match validated release.module metadata before any
archived code executes. It streams body hashes and preflights physical
headers before extended metadata parsing. Only regular files and one bounded
path-only PAX header per file are accepted; size/sparse/global/GNU framing
extensions are rejected, so physical and decoded framing agree. It then extracts its immutable verified byte snapshot to a new SDK
path containing spaces. It generates an external application with the extracted
`scripts/new-app.py`, checks the four required replacement directories match the
exact SDK layout (root, `drivers/cogent`, and both prepared dependency trees), verifies the external Go module cache, then tests/vets/builds with `GOWORK=off`. It also tests/vets the independent
public API scenarios. Success cannot retain an original-checkout replacement.

Without an independently trusted digest/signature, the manifest checks internal
integrity rather than artifact authenticity. Executing extracted SDK code requires
that digest or explicit `--allow-self-attested` for an owner-created private build;
the report records `digest_verified_against_expected` as false for that opt-in. Child processes
receive an explicit toolchain/cache/native/display environment allowlist that
drops unrelated API/cloud/CI token variables. Desktop session and GPU selection
variables are supplied only to the explicit native runtime launch, not to generator,
Go tests/builds or producer children. Producer Git commands use the same controlled
environment and deadline; ambient `GIT_*` redirection cannot select another repository
or index. Producer/checker commands force
`GOENV=off`, `GOWORK=off`, local Go toolchains, `GOFLAGS=-mod=readonly`, the default
public Go proxy/SumDB and empty checksum/insecurity/private-module bypass lists;
ambient modfile/overlay/Go configuration cannot replace the recorded SDK source. This is not a sandbox: archive code
still runs with the caller’s filesystem and any explicitly preserved build settings;
self-attestation is for the caller’s own private source, not unknown downloaded code.
Native cgo overrides are fixed: cgo is enabled and only the pinned Darwin rpath
flag is explicitly allowed; ambient compilers/flags/allow-patterns, `GOROOT` and `PKG_CONFIG_PATH` are dropped.
The standard target compiler must be on PATH. HTTP(S) proxy/no-proxy and TLS CA
settings remain available for ordinary network transport; Go proxy/SumDB and
verification-bypass settings remain fixed. Proxy URLs may contain credentials and
TLS paths may identify private transport configuration; these are intentionally
preserved for ordinary dependency network transport and are not claimed removed.
`HOME`/`USERPROFILE` are retained for ordinary tooling; filesystem privileges still
allow archive code to read files such as home-directory credentials. This is explicitly
unisolated execution, not a promise to protect those files or transport credentials.
This verification profile requires the public Go proxy/SumDB or an already populated
verified external module cache; offline/corporate-mirror overrides are unsupported.
Do not use it where contacting the public dependency services is prohibited.
Only execute SDK code you already authorize with those privileges. Generator commands have a 60-second
deadline, Go build/test/graph commands 600 seconds, native runtime 30 seconds,
and app dependency inspection/signature commands 60 seconds.
SDK producer/checker and app packaging child commands run in owned POSIX sessions, and timeout or
interruption kills that process group before bounded pipe/reaping cleanup. Windows
uses fixed-argv `taskkill /PID <owned-pid> /T /F`; failure to confirm cleanup fails
acceptance with the original timeout retained as its exception cause. An already
exited parent does not prove its descendants exited: if taskkill cannot confirm
tree termination, cleanup remains unknown and the check fails. Real POSIX grandchild cleanup is tested locally; Windows cleanup has
transport fixtures only and remains target-host acceptance unexecuted. This does
not isolate arbitrary code or catch descendants deliberately leaving the owned
group. Successful/nonzero normal command completion keeps ordinary subprocess
semantics, rather than promising to revoke separately launched background work.
A timeout fails acceptance and never emits a success report. Validation, archive
parsing, filesystem and failed-child-command errors also emit structured
`vellum.sdk.check_failed` JSON to stderr with a nonzero exit and no success JSON. `--runtime` additionally launches the
relocated GPU example and requires its verified pixel/readback report. Default
checks create no native windows and do not change clipboard or OS preferences.
A runtime check records its host only; Windows/Linux qualification remains a
separate target gate.

## Public release gates

The namespace and framework-license selections are complete and recorded in
`release/version.json`. Third-party licenses and notices remain unchanged; this
MIT selection applies to Vellum's own code and does not approve redistribution
of dependency/native/font components.

`--public` fails until the owner has selected a non-placeholder module namespace,
a framework license is selected and present, and the third-party/native
redistribution audit is approved. Dependencies without their own conventional text license declaration at the
module root or a top-level license directory also block public packaging. Source
files and test/example fixtures cannot satisfy that gate; nested third-party
notices can be retained as evidence but do not establish the module’s license. Public native sources are freshly regenerated in owned temporary directories
from pinned module versions/checksums and validated patch inputs/outputs. Go
module verification checks the actual cached archive and extracted directory
hashes before and after regeneration, and the public license graph verifies its
external module cache too. Bootstrap subprocesses use that same controlled environment; ignored mutable prepared trees cannot
enter a public SDK. The license graph uses these freshly verified replacements.
Public source must be committed and clean,
and its Git state must remain unchanged through packaging. The declared root module must match release metadata. Namespace changes must
be applied consistently to both Go modules, generators, examples and metadata.

The module inventory copies declared license/notice files and records checksums,
missing records and scope. Locally replaced native modules retain their
pinned source version/checksum in the inventory, rather than losing those values
when Go reports a directory replacement. It is evidence for an audit, not legal approval or a
complete native/Rust/font component inventory. Owner approval must be based on
actual distribution artifacts and applicable notices before marking that gate
approved. SDK release/check tools do not choose a license, push, tag, sign,
notarize or publish. App signing requires the explicit identity described below.

Release acceptance also tracks native accessibility/input, actual target runtime,
interactive services, device/DPI/lifetime handling and measured performance gates
in their separate specifications. Alpha packaging does not erase those limits.

## Private application packaging

`scripts/package-app.py` packages a prebuilt executable into a macOS `.app`,
Windows ZIP or Linux tar.gz using explicit name/identifier/version/target values.
`--version` supplies three numeric components for `CFBundleShortVersionString`;
`--build-version` supplies one to three numeric components for `CFBundleVersion`
(default `1`). SDK semantic release metadata remains in `APP_MANIFEST.sdk_release`
and never enters numeric Apple bundle fields. The owner increments build versions
for distribution. These formats follow Apple’s
[release version](https://developer.apple.com/documentation/bundleresources/information-property-list/cfbundleshortversionstring)
and [build version](https://developer.apple.com/documentation/bundleresources/information-property-list/cfbundleversion)
contracts.
It refuses existing outputs, mismatched binary formats and app names with
control/reserved Windows characters, device names, leading dashes or trailing
spaces/dots, and the reserved `APP_MANIFEST.json` layout filename. macOS outputs require
a `.app` suffix. All layout directories/executables use `0755`, other files use
`0644`; ZIP/tar metadata and gzip timestamps are normalized independently of host
umask and builder identity. Formats and dependency closure are validated again
on the staged executable bytes before any signing or publication. macOS packaging uses
`otool` to reject non-system dependencies until an audited bundled dependency
closure exists; the current statically linked example can be packaged without
prototype helpers. Windows/Linux layouts are not proof that dynamic native
runtime dependencies are complete or that the app ran on those targets.

macOS bundle publication reserves the output exclusively and fills that directory.
It is non-atomic: successful command return and its JSON report mark completion.
Normal exceptions remove the owned partial directory, while forced termination or
power loss can leave partial output requiring inspection/removal before retry.
The tool refuses existing output rather than overwriting that partial state.

The app manifest marks the artifact private, target runtime unverified, native
redistribution audit pending and notarization absent. macOS signing occurs only
when `--sign-identity` is explicitly provided and is verified at both the staging
and final bundle locations; `-` means
ad-hoc signing. Default packaging does not sign, notarize or publish. Binary
source checksums precede any explicit signing transformation. Application signing
and redistribution require separate owner decisions and target acceptance.


## Native linker-map observations

The distribution collector optionally accepts both `--binary` and `--link-map`.
They must be supplied together, have matching supported thin 64-bit Mach-O/map
architecture and match the selected Go build architecture. The collector records
artifact size/hash identities and the linker object-table count. It matches
absolute archive paths only to verified module-owned linker candidates, then
hashes exact selected member payloads and records original archive ordinals.
Relative archive paths remain unresolved because the linker's cwd is unbound;
unknown absolute archives also remain unresolved without reading their contents.

The bounded reader supports BSD extended and short ar names. It rejects GNU
long-name tables, thin archives, malformed/truncated headers or alignment,
symlinks, nonregular inputs, surplus names/members/bytes and missing/ambiguous
selected members. Admission opens nonblocking before checking regular-file type,
so a FIFO with no writer cannot block. Payloads stream in 64 KiB chunks and are
never extracted or executed. Limits are 1 GiB/archive, 10,000 members, 4 KiB/name,
128 MiB/map and 20,000 object-table entries. Mutable stream metadata refuses;
collection repeats observations and retains the existing full input/source
and package metadata drift checks.

`artifact_pair_binding` remains `caller-supplied-unverified`. Matching hashes and
architecture alone cannot establish that a supplied map belongs to that binary.
An independent owned build receipt must bind both outputs after successful
compiler completion. Object-table membership does not mean every member payload
byte survives dead stripping. Object names are not Rust crate version, source,
license or notice identities. This observation does not approve distribution or
change the public audit gate. See the [Mac native-link checkpoint](../docs/reports/macos-link-closure/README.md).

Example for private investigation after dependency preparation:

```sh
go build -ldflags="-linkmode=external -extldflags=-Wl,-map,/tmp/vellum-workbench.map" -o /tmp/vellum-workbench ./cmd/workbench
# Run the collector from the repository root, with the same controlled Go cache.
python3 scripts/distribution-inventory.py --binary /tmp/vellum-workbench --link-map /tmp/vellum-workbench.map --output /tmp/vellum-link-inputs.json
```

Use fresh output paths and preserve the exact build receipt; the example alone
is not a source/build provenance certification. No application is launched.

## Owned private Mac build

`scripts/build-macos.py --output <new-directory>` owns one successful
`go build` invocation of `drivers/cogent/cmd/workbench` on a Mac host, with
external linking and a linker map. The resolved Go executable, exact argument
vector, cwd, selected Go configuration, controlled environment digest/key names
and sampled source/SDK inputs are recorded. Generated staging paths avoid Go
linker quoting ambiguity; the chosen final output directory may contain spaces.
The produced app is never executed. No signing identity, notarization or
publication is requested. The native dependency preparation remains a prerequisite.

Before/after inventories, Go executable identity, producer identity and controlled
environment must agree. Initially absent binary/map paths are hashed after exit
zero, checked through the collector and checked before publication. The collector
continues to label its observation `caller-supplied-unverified`; the separate
build receipt labels the verified pair `producer-observed-same-build`. The stronger
label belongs to this observed invocation and its artifact hashes, not arbitrary
future collector input. Linker `# Path` need not equal the Go output path because
Go may use an intermediate native linker output.

The output directory is exclusively created outside the checkout with private
permissions. Descriptor-relative, no-follow, exclusive writes refuse replaced
parents/directories/files. Artifact reads stream in 64 KiB chunks within the
collector's 1 GiB file bound. Compiler stdout/stderr drain concurrently with a
4 MiB bound per channel and a 600-second deadline; overflow, timeout or sink failure
terminates the owned child group using the bounded cleanup path. The streamed
log counts/hashes and serialized inventory hashes must match published files.
Normal subprocess completion retains its existing semantics; it is not a claim
that all possible descendants have retired.

`build-receipt.json` is the last published file and binds six artifacts: executable,
map, before/after inventories and both logs. Its own published bytes are checked.
Successful command return and the receipt are required; directory creation alone
is not completion. Failure retains diagnostics when the owned destination remains
valid, removes only an owned partial success receipt and refuses to overwrite
existing/replacement data. If receipt retirement or diagnostic publication fails,
the original exception is preserved with a cleanup diagnostic; CLI failure JSON
reports the phase and unconfirmed output completion, including unconfirmed receipt
retirement when applicable. Directory publication is not atomic, and no success is
claimed for partial diagnostics. Consumers must validate the receipt and hashes
when consuming artifacts, since files can change after the producer returns.

This is sampled local evidence, not a hermetic build, cryptographic attestation,
independent source provenance, complete Rust/C dependency closure, redistribution
rights approval or Mac runtime qualification. Transient input changes reverted
between observations cannot be ruled out. The public audit gate stays pending.
See the [owned-build checkpoint](../docs/reports/macos-owned-build/README.md).

## Private source-built GPU workbench compatibility

The [private workbench consumer](../docs/reports/macos-source-gpu-framework/README.md)
uses alternate module manifests and an unchanged static-only replacement outside
the checkout. Both ARM64 Go 1.27.2 consumers declare macOS 13.0; the workbench links
141 entries from the exact source-built archive and loads no WGPU dylib. Its
independent public automation client links no WGPU objects. Two owned native
workbench cases pass 80 source-reviewed semantic requests and normal closure.
Compiler warnings and both earlier validator failures remain recorded. This is
private compatibility evidence, with no installed library, SDK or runtime source
adoption. Source/component/standard-library rights closure, broader Mac runtime,
physical input/presentation, performance/memory/endurance and distribution remain
separate release gates.

### Explicit private GPU SDK overlays

Private `release-sdk.py` builds may carry a digest-pinned ARM64 static WGPU
candidate and separately hashed caller-declared provenance. All-or-none input
validation and a hard public-release refusal precede inventory and output work.
The bounded archive/header checks, four-file capsule, derived module hashes,
absolute generated-application replacement and strict relocated Go origin follow
[sdk.md](sdk.md#private-arm64-mac-gpu-candidate). Default SDK packaging does not
select an ignored capsule. The private manifest retains pending audit status;
matching library bytes, module origin and copied Go notices do not establish
source attestation, binary ABI/runtime compatibility or complete component/rights
closure. Observed build/linkage and admitted native qualification remain required.
