# Local SDK consumption and performance measurements

Current private macOS v1 delivery follows the [release closeout](macos-v1-release.md)
and [stable API contract](stable-api.md). The older developer-preview and personal
SDK records retain their exact source identities. Existing inventory, dependency,
consumer-origin and distribution guards below remain in force. Broader numeric
performance/platform qualification remains separate from this bounded personal
release; this pass neither changes its existing policy nor marks M3–M7 complete.

The root framework module stays standard-library-only. The native adapter is a
separate module, `github.com/dshills/vellum/drivers/cogent`, matching its repository
directory. The framework namespace is `github.com/dshills/vellum` and its own code
is MIT-licensed. Public distribution remains gated on the dependency/native audit
and target qualification.

Current source, public prepared downloads and private candidates are separately
identified in [project status](../docs/status.md). The published build13 SDK is
older than v1.1.3 source. Use a selected SDK's bundled docs for its actual APIs;
do not assume a version bump rebuilds the SDK. The
[Work Log recipe](../examples/work-log/README.md) demonstrates generation from an
identified exported SDK.

## Documentation in the SDK archive

Consumer guides and contracts retain their links to other packaged files.
Repository-only reports, reviews, research, artwork, prototypes and repository diagnostic tools remain excluded.
For first-party inline Markdown links into those known existing source trees,
the exporter retains the label and exact repository path/fragment as an explicit
nonlinked full-repository reference. It adds a link to `SDK_DOCUMENTATION.md`,
which explains that inspecting this evidence requires the matching full checkout.
No remote publication or additional qualification is implied. Unknown missing or
escaping inline targets refuse packaging rather than becoming evidence references.
Code examples, third-party prepared documentation and native-component
`NOTICES/` trees remain unchanged. First-party component READMEs still follow the
normal documentation policy.

`SDK_MANIFEST.json` hashes the actual archived document bytes and generated scope
note. Its additive `documentation_overlays` record contains each original source
hash, derived archived hash and exact rewritten target/reference mappings.
These derived document bytes must be associated through this record when comparing
an SDK against its producing source; they are not claimed to be unchanged source.

## Public document editing

The SDK includes the standard-library-only `editor` package and its managed
`ui.MultilineText` descriptor. The independent `examples/editor` module uses
exact document bytes, one history journal and the native input settlement API
before persistence. SDK relocation checks remove the producing source tree and
compile a consumer that reads the exact CRLF/Unicode document from the exported
package. The public API contract consumer also exercises composition, stale
content epochs and `document.Session` installation.

## Native dependency preparation

`drivers/cogent/prepare.sh` and `env.sh` own preparation and Go cache setup.
Preparation reads only driver-owned inputs. Its manifest pins Cogent and go-text
module versions/checksums, patch digests, original input hashes and patch order.
It downloads immutable module source, retains upstream license files, copies into
writable owned staging directories, verifies hashes and applies patches there. The driver multiwindow safety stage
also verifies expected final source hashes and runs its ownership regression tests
through the repository check script.
Both dependencies must pass before replacing existing prepared directories. Failure
rolls back installed directories. If filesystem errors prevent full rollback,
preparation retains original backups in the reported recovery staging directory
and reports their exact paths. Inspect and restore those backups before removing
the recovery directory. A filesystem lock rejects concurrent preparation.
A killed process can leave the lock; inspect for an active process before removing it.

Driver replacements resolve to `drivers/cogent/build/upstream` and
`drivers/cogent/build/typesetting`. Prototype modules and evidence remain historical
and are not required to prepare, test, build or consume the framework modules.

`check-text-backend.py` copies prepared upstream module metadata to a temporary
location and explicitly resolves its typesetting dependency to the prepared
owned tree. It runs exact-text, composition and visual geometry regressions with
an initialized offscreen display and a driver-owned installed-font cache. Direct
upstream tests do not inherit the driver module's replacements. The rendered
fixtures need offscreen display initialization; those results are separate from
native AppKit consumer runs. The helper does not mutate prepared source or OS
preferences.

## Independent application module

`examples/hello` has an independent `go.mod`, imports public framework packages and
the native driver, and renders a working counter and GPU rectangle. Its explicit
replacements cover both framework modules and both patched native dependencies.
Dependency-module replacements are not transitive in Go.

`scripts/new-app.py` creates the same application in a new directory, with a caller
selected application module path and absolute local framework/dependency replacements.
It retains the starter template's MIT LICENSE notice in the generated application.
It refuses existing destinations and requires preparation first. The generated
application can be tested and built outside the repository with `GOWORK=off`; it
still requires this checkout for its local SDK dependencies. Publishing versions,
installing a remotely fetched SDK, and relocating those dependencies are separate
release tasks.

## Repeatable measurements

`scripts/benchmark.sh` runs existing portable data benchmarks plus the standalone
application's UI construction/validation benchmark, with allocation reporting and
three repetitions. `--native` additionally builds and launches three visible native
processes via `benchmark-native.py` on a desktop with a supported GPU.

The standalone app's `-benchmark` mode disables native window-geometry loading
and saving before constructing the app and restores its previous process-local
switch on normal return after teardown. Ordinary interactive behavior is unchanged.
Its report includes the actual `geometry_persistence_disabled_in_fixture` flag;
`scripts/check-sdk.py --runtime` requires exactly one GPU report with that flag
strictly true. The generator copies this same starter, so relocated native checks
exercise the geometry-isolated benchmark rather than saved application geometry.

Every native run reports:

- Startup from application entry to a pixel-verified GPU rectangle; external process
  launch-to-report timing is also recorded. Both include initialization; first frame
  readiness is polled at 25 ms, so this is not compositor presentation latency.
- Ten synchronous GPU render/readback durations. Readback forces rendering; these
  observations include transfer/synchronization and are not ordinary frame FPS.
- Go heap allocation, native image dimensions, runtime/platform/machine metadata,
  and resident memory samples before/after the quiet interval.
- Process CPU-time difference during 60% of the quiet interval after a brief render-settling delay, expressed as seconds
  and percentage of one core. Linux uses `/proc` CPU ticks; other supported
  hosts use `ps` accounting. Reports record the sampling source and time quantum.
  Accounting precision limits small deltas; a zero observed delta does not prove
  zero idle CPU. Linux accounting parsing has fixture tests; native Linux runtime
  measurements remain unverified.

The harness fails if no verified GPU report arrives, the application exits with an
error, or sampling fails. Measurements provide host-specific baselines without
invented pass thresholds. Cold OS/font caches, other desktop processes, GPU/power
state and hardware affect results; comparisons should record matching conditions.
Windows/Linux runtime performance and accessibility/input correctness remain gates.

The public `document` package ships in private SDK archives alongside app/ui/history.
A relocated archive consumer exercises bounded file persistence after deleting
the fixture source checkout. Session/close policies follow [document.md](document.md).

The portable SDK exports `execution` alongside `tasks`, `document`, `editor` and
`app`. The archive relocation test deletes its source fixture and consumes these
packages through an independent module, including a manually stepped task whose
final result is admitted while the ordinary owner queue is full.

## Private ARM64 Mac GPU candidate

The SDK builder accepts an explicit private static-library overlay:

```sh
python3 scripts/release-sdk.py --output /private/tmp/candidate-sdk.tar.gz \
  --macos-gpu-library /absolute/path/libwgpu_native.a \
  --macos-gpu-sha256 <64-lowercase-hex-digest> \
  --macos-gpu-provenance /absolute/path/provenance.json
```

All three options are required together. Private candidates retain pending audit
status. The approved public route below additionally requires the existing release
approval gates and notices bound to the selected archive before output creation.
Ordinary packaging does not discover or reuse an ignored candidate directory.
The root framework module and portable example modules remain unchanged.

The library must be a regular, bounded ar archive with the expected SHA-256 and
at least one object. Every non-index member must have a thin ARM64 Mach-O
`MH_OBJECT` header; fat, foreign-architecture and unsupported payloads refuse.
This observes headers and exact bytes, not complete Mach-O semantics or ABI
compatibility. The same captured bytes are packaged. The archive limit is 64 MiB;
the provenance JSON limit is 1 MiB, with duplicate and nonfinite JSON values refused.
Provenance requires integer `schema: 1`, `target: "darwin-arm64"` and
`artifact_sha256` matching the selected library. Additional fields are retained
as caller-supplied declarations. The builder never interprets them as commands,
source attestation, executed-build evidence, component closure or rights approval.

A new private candidate supplies five Vellum-authored compatibility modules at
`drivers/cogent/build/libs-{android,darwin,ios,linux,windows}`. They retain the
upstream module/import identities because the WebGPU binding blank-imports all
five, but select freshly authored source with its own MIT declaration. The
original upstream linker-module distributions are not selected. This does not
change the separately selected WebGPU binding, headers or native archive.

Every compatibility module has `go.mod`, `package.go`, `unsupported.go` and
`LICENSE`. Darwin additionally has `link.go`, `arm64/libwgpu_native.a` and
`PROVENANCE.json`. Only Darwin ARM64 with CGO selects `link.go`, whose flags name
the exact archive path rather than searching for an alternative WGPU library.
Unsupported Darwin architectures/no-CGO and the four other native targets fail
compilation with `VellumSDKRequiresDarwinARM64WithCGO`. Non-target packages are
inert so the five unconditional imports remain consumable on the supported Mac.
The private candidate does not qualify another platform.

`SDK_MANIFEST.native_gpu.owned_linkers` binds all five module directories,
authored source/license hashes, role and target policy. The Darwin `files` record
also binds the exact archive/provenance; object count, unverified trust label and
pending distribution audit remain explicit. Module rewrite records retain both
source and archived hashes. `source.commit` names the framework baseline and
`source.clean` describes the checkout; generated sources and derived module files
are disclosed overlays. Earlier one-Darwin-capsule records remain readable as
legacy records and acquire no new ownership or licensing claim.

The notice collector inventories the same effective replacement graph used by
consumers. Each owned entry has `status: owned-source-declaration`, null effective
`version`/`sum`, original resolution metadata in `selected_requirement`, and an
`effective_source` with the SDK-relative directory and authored file hashes.
The owned license points to that module's SDK file. It is not copied into the
upstream notice collection or presented as an upstream declaration. An unrelated
missing declaration remains unresolved. Go source declarations remain separate
from native/Rust/standard-library/font/asset applicability and distribution audit.

Driver and native example main-module manifests receive all five relative local
overrides, with conflicting replacements refused. Generated standalone apps
receive all five absolute overrides because dependency replacements are not
transitive. The generator rechecks all capsules and module-overlay bytes,
refusing symlinks, partial/changed source or notices, and extra files such as a
competing dylib. It retains destination and starter-license behavior. The archive
verifier couples the optional record to exact membership; orphan capsules and
mismatched records refuse. Relocation requires every actual replacement origin
to equal its SDK-local directory and runs candidate builds only on an ARM64 Mac,
explicitly targeting `GOOS=darwin GOARCH=arm64`. Direct source-checkout builds and
tests from the prepared upstream module remain on their existing graph unless
the caller explicitly supplies the candidate replacements.

Digest and origin validation do not prove static linkage or runtime behavior.
A candidate qualification must separately inspect the compiled artifact/linkmap
for the exact archive and absence of WGPU dylib fallback, then run admitted native
checks. Other hosts/architectures, older OS runtimes, presentation, physical input,
performance/memory/endurance, signing and public distribution remain separate.

## Clean private Mac candidate verification

The [clean SDK consumer evidence](../docs/reports/macos-source-gpu-sdk-native/README.md)
binds one `3da753e` private SDK to an independently generated native starter after
relocation with spaces. All 7,436 payload hashes match the prior non-GUI-tested
precommit archive; source/untracked manifest metadata is separately identified.
The retained ARM64 Go 1.27.2 build declares minimum macOS 13.0, maps 141 entries
to exact packaged static-archive members and has no WGPU dylib dependency.
One separately admitted native run passes the permissive initial Snapshot center
predicate and three further successful Snapshot calls, then retires its owned
group. Image-wide equality, compositor presentation, physical input, ordinary
latency/memory/endurance, other SDK candidates and public distribution are not
qualified by this check. The raw warnings and internal reviews remain retained.

The [current private SDK GPU checkpoint](../docs/reports/macos-gpu-sdk-stage0077/README.md)
packages clean `5eef841` with 7,441 hashed payload files, stage 0077 and all three
corrected GPU examples. Relocated generated-consumer race/vet/build and public
API race/vet checks pass. Drawing, 3D and imaging each pass one serial race-enabled native
source-assertion/readback case with exact packaged ARM64 static linkage and owned
shutdown. The initial relative-path verifier failure remains failed; seventeen
completed component outcomes are reused and sixteen remaining commands pass in
a fresh continuation. Three actual grouped cancellation controls and eight path
controls qualify the bounded test harness. This is a private consumption checkpoint,
with internal reviews and no unchanged implementation Prism rerun. Physical/human
workflows, presentation, ordinary/sustained latency, native/GPU memory, endurance,
component/rights closure, public release and full Mac acceptance remain open.

## Approved public native-library route

The same three native-library options may be used with `--public` only after the
existing release gates pass, including `third_party_distribution_audit=approved`,
a clean committed source tree, selected namespace/license and index checks. Public
prepared dependencies still regenerate and verify through the existing path.
The selected archive must additionally have a component inventory and notice files
bound to its exact SHA-256; absent or other-archive notices refuse packaging.

The public manifest and generated third-party inventory record the release audit
as approved. The copied component declarations retain their historical pending
audit status: collecting notices does not grant distribution rights or approve a
release. Relocated consumers and archive verification require both the public
release approval record and the native notice record for consistency. These
manifest fields are declarations, not signatures or independent rights approval;
a consumer must obtain the expected archive SHA-256 through a trusted channel to
bind checks to the intended producer artifact. The producer alone enforces the
existing source release-audit gate. Consumers also check the
notice inventory against the selected archive bytes. Private candidates retain
pending status and may carry an unknown archive without matched declarations.

This route changes tooling only. The repository's current release audit remains
pending, so public packaging is still blocked. No public artifact is authorized
or produced by this checkpoint; caller-declared provenance remains unattested.
