# SDK versioning, compatibility and release artifacts

`release/version.json` and the public `vellum.Version` constant identify
`v0.1.0-alpha.1`. This is an alpha packaging/API version; no Git tag or remote
release is created by these tools. The owner selected `github.com/dshills/vellum`
and the MIT license for Vellum's own code. The native module is
`github.com/dshills/vellum/drivers/cogent`; example module paths also match their
repository directories. Root and module-local LICENSE files record this choice.
The third-party/native distribution audit remains pending. The Go module graph notice inventory does not establish
complete embedded Rust/native binary/font redistribution rights.

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
