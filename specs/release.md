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

The `go_standard_library` section records the Go-selected standard-library
package source, assembly, cgo, embedded and `.syso` object files from the installed
GOROOT, with normalized relative paths, sizes and SHA256 identities. It records
named ancestor LICENSE/COPYING/NOTICE/PATENTS candidates and package-to-file
references. Standard package metadata now participates in the final Go-list
resampling; file bytes and newly added/changed declarations are rechecked. The
existing owned build producer compares this section before and after compilation,
so standard-library source drift refuses a successful build receipt. External
GOROOT aliases are supported; symlinks inside the selected installation, including
links back to its root, are refused. Source/declaration and package lists are
bounded before surplus hashing.

These are the installed Go tool's selected inputs, not authenticated upstream
source, compiler-generated data, transitive includes, external SDK headers,
compiler/linker component closure or code surviving linking. Nearby declarations
remain candidates; notice applicability and redistribution approval remain open.
Older schema-1 inventories without this additive section do not establish this
coverage. The inventory does not update distribution gates or execute the app.

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

The [pinned Go-module declaration audit](../docs/reports/macos-module-declarations/README.md)
confirms that all five WebGPU linker-module notice gaps exist in their exact
upstream root/module-root trees, with matching selected Go source bytes. These
are absent declarations, not collector omissions. The82-component source-built
native archive bundle does not close those Go-module gaps. Applicable declarations
or a reviewed dependency replacement remain required; the distribution audit
stays pending. No runtime, SDK or Mac milestone changes at this checkpoint.

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

## Current private SDK minimum-Go consumption

The [stage0079 private SDK checkpoint](../docs/reports/macos-sdk-stage0079/README.md)
packages clean `abe67c6`, including M1 cancellation and the source-built GPU notice
bundle. Its 7,879 payload files pass complete integrity, actual Go 1.25.6 relocated
consumer/public-contract checks, all 26 default-main builds and the explicitly
tagged window-cycle helper. Twenty-eight distinct shipped fonts are mapped to
name-table metadata and declaration candidates, with twelve selected Darwin
embeds. Initial cache/inventory failures are preserved. No native application is
launched; final runtime, complete component/font/non-font/generated asset rights,
owner distribution decision and other release gates remain open. This is a
private consumption checkpoint, not final release qualification.

The subsequent [current SDK native checks](../docs/reports/macos-sdk-native-stage0079/README.md)
execute starter, drawing, 3D and imaging fixtures successfully using those same
Go 1.25.6 non-race binaries, with GPU readback/source assertions, geometry isolation
and owned closure. No rebuild or repeated Prism call is needed for unchanged
implementation evidence. Full native/human, performance/memory/endurance, older-OS,
rights and release gates remain open.

## Private source-built GPU consumer

The [pinned-source GPU consumer checkpoint](../docs/reports/macos-source-built-gpu/README.md)
builds the exact WGPU source/header commits with recorded Rust/compiler/features,
then passes private static-only Go consumers and nine headless GPU roots with
23 subtests. Two Rust stripping warnings and the failed initial consumer attempt
remain preserved. The candidate is not adopted into installed libraries or the
SDK; complete component/rights closure, broader native UI/presentation, older
macOS runtime, performance/memory/endurance and full Mac release remain open.

## Source-built GPU component notices

`release/native-components/wgpu-darwin-arm64/INVENTORY.json` binds collected
notices to the exact source-built ARM64 archive SHA256. The current bundle covers
all 80 compiled Cargo packages (103 compiler units), pinned webgpu headers and
the executed Rust 1.99 toolchain's bundled standard-library declarations. Its
82 component entries and 408 notice/declaration files are a conservative compiled
superset, including host/build dependencies; they are not an exact linked-crate
inventory, license selection or redistribution approval. Original declarations,
copyright texts and alternative license files are retained without invented
copyright holders. Missing crate-package notices use exact VCS-commit repository
declarations; hexf-parse's CC0 declaration has the official standard-text supplement.

The existing SDK producer automatically attaches a `component_notices` record
only when the private GPU candidate's archive hash matches. The known source-built
archive refuses a missing or mismatched inventory. Its checked-in
inventory and notice files travel through the existing release source path,
with their exact file hashes and component count referenced by the SDK manifest
and `THIRD_PARTY_INVENTORY.json`. Unknown candidates receive no matching claim.
No additional native capsule member or module overlay is introduced.

Producer and extracted SDK consumer validate inventory schema, artifact binding,
unique component identities, notice references and exact SDK member/hash coverage.
They refuse traversal/symlink paths and apply bounds of 512 files, 256 components,
4 MiB per file and 16 MiB aggregate. Validation after source enumeration detects
an omitted or changed notice before publication. Extracted validation rechecks
inventory semantics, not only individual file hashes. Older private capsules
without the optional record retain their prior pending-audit behavior.

The bundle's source/metadata receipts and open issues are recorded in its inventory
and [notice guide](../release/native-components/README.md). The Apple-derived
binding declarations, exact runtime-subcomponent attribution, generated code,
fonts/assets and owner distribution decision remain open review items. No public
release gate changes. A final SDK refresh and relocated native acceptance still
follow runtime stabilization; fixture packaging is not that final qualification.

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
SDK producer/checker and app packaging child commands run in owned POSIX sessions.
Timeout and Python interruption retire that group before bounded pipe/reaping
cleanup. On POSIX, synchronous main-thread calls temporarily handle default
SIGTERM and SIGINT dispositions: cancellation during `Popen` creation is deferred
until the returned child handle is owned, then the group is retired. When
cancellation initiates failure, SIGINT propagates `KeyboardInterrupt`; SIGTERM
propagates `ChildCancelled`, a `SystemExit` with status 143. These initiating
cancellations are not swallowed by ordinary `Exception` handlers. Signals received
during an already-failing cleanup preserve its original failure instead. Further
SIGTERM/SIGINT delivery cannot interrupt retirement. Original handlers are restored
on success, child failure, spawn failure and cancellation.

Custom and ignored signal dispositions remain caller-owned. Worker-thread calls
do not install process-wide handlers; they retain deadline/Python-exception
cleanup but require caller-managed process-signal cancellation. This does not
bound an OS process-creation call before `Popen` returns, intercept SIGKILL or
handle every termination signal. Normal/nonzero completion retains ordinary
subprocess semantics, including deliberately launched background work. Windows
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

Mac applications can optionally supply `--notice-bundle <directory>`. The
packager admits the bundle against the **staged executable's pre-sign SHA256**,
then copies its original inventory and declaration bytes into
`Contents/Resources/THIRD_PARTY_NOTICES`. `APP_MANIFEST.component_notices` records
the inventory hash, executable binding, file/component/byte counts and pending
review status. Legacy packages without this option make no notice-coverage claim.
The option is currently Mac-only; it neither discovers nor selects applicable
licenses automatically.

The input directory contains `INVENTORY.json` and exactly the files that inventory
lists. Schema 1 requires these fields and no others:

- `schema`: integer `1`; `target`: `darwin`.
- `artifact_sha256`: lowercase SHA256 of the executable before signing.
- `distribution_audit`: `pending`; `applicability`: `review-required`.
- `scope`: a nonempty description of the caller-selected coverage, up to 4,096 characters.
- `files`: a map of `NOTICES/...` paths to exact `{bytes, sha256}` records.
- `components`: objects with unique nonempty `id` values (up to 512 characters)
  and nonempty `notice_files` lists referring to those paths. Every file must be
  referenced; several components may share one declaration.

The validator refuses duplicate JSON keys, nonfinite values, unrecognized schema
fields, invalid identities, duplicate component/references, case-colliding file
or directory spellings, traversal, nonregular inputs and symlinks at the root or within it. Parent
installation aliases are resolved before descendant checks. Notice paths use
ASCII letter/digit-leading components followed by letters, digits, dots, dashes
or underscores, with no trailing dots; each starts with `NOTICES/`, has at most
16 components and at most 1,024 characters. Extra files/directories refuse.
Limits are 1 MiB inventory, 1,024 declaration files, 512 components, 4 MiB per
file and 16 MiB including inventory. Counts and declared byte totals are checked
before declaration contents are read. These limits bound admitted memory and
copy work; source paths/bytes are sampled rather than isolated against concurrent
adversarial directory replacement.

The copied bytes come from a validated bounded snapshot. The packager checks
staged notice contents after copying. It checks executable bytes, exact app
manifest bytes and optional notices before signing/publication and again at
the final output. Explicit signing may transform the executable; the command's
JSON result records `packaged_binary_sha256` after that transformation, and final
copy verification uses that identity. The app manifest and notice inventory
retain the pre-sign source identity. The unchanged expected app manifest is
also checked after signing. These are sampled byte/identity checks, not filesystem
isolation. This is caller-selected material,
not attested component provenance, complete linked membership, applicable-license
selection or rights approval. Both inventory and app manifest remain pending;
source-built GPU notices must not be applied to a different native archive.

macOS bundle publication reserves the output exclusively and fills that directory.
It is non-atomic: successful command return and its JSON report mark completion.
Normal exceptions remove the partial directory only while its sampled device/
inode identity still matches the directory this call created. A detected
replacement is preserved and packaging fails. A failure immediately after output
creation, before its identity can be captured, preserves that unproven empty
path too: no payload is moved and no completion result is returned. Inspect and
remove that empty partial output manually before retrying; the tool never guesses
that an unproven path is safe to delete. This identity check is not isolation
against adversarial concurrent path changes. Forced termination or power loss can
leave partial output requiring inspection/removal before retry.
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
