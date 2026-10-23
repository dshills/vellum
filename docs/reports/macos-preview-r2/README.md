# Preview R2 — independently consumable candidate

A private SDK from clean `d8de4fc` exports successfully after the focused
repository-only tools-reference repair. SHA256
`d842f20676cbb2fb68412b659c91b4954b67a067c2a1ce407b149a81a92df405`
binds all 8,451 manifest members. Repository tools remain excluded. Native archive
`786aaad8` and production rendering are unchanged; this is not the older bd3cbc9 SDK.

The existing SDK verifier generated a new external Go module under a fresh
space-containing temporary path. Integrity, module-origin containment and module
verification pass, then starter race tests/vet/build, public API contract race/vet,
and bounded native launch/GPU readback/normal exit pass. The declared Go 1.25.6
renewal passes, alongside the first installed Go 1.27.3 run. These native smoke
checks do not measure frame latency, physical input, or native deallocation.

The matching Placement app builds from a separately verified extraction of the
same SDK in a filtered Go 1.25.6 Darwin-arm64/cgo environment. It packages at
`/private/tmp/vellum-preview-r2/Vellum Placement Preview.app`, ad-hoc signed with
strict verification. Application interactions and resource acceptance are R3.
Candidate metadata remains alpha.2; coordinated alpha.3 metadata and final
repackaging are R4/R5. The package is private, not notarized, with rights pending.

Failures retained: the initial SDK export refused an excluded tools link; the
reviewed exporter repair passes 43 tests. A direct Go 1.25 build inherited a Go
1.27 GOROOT and failed; filtering the environment fixes that without source edits.
One Documents-location packaging attempt failed strict signature verification
with resource-fork/Finder metadata and was removed by the tool. Temporary staging
passes; final archive extraction/signature acceptance remains mandatory in R5.
Compiler AppKit deprecation and duplicate-linker warnings are reported, not a
claim of a warning-free build. Original stderr and source/toolchain records are
retained. No public distribution audit or full Mac qualification is implied.

Prism using Anthropic Claude Sonnet 5.5 reviewed eleven changed result/ledger
files with complete coverage. Its one finding concerned home-directory paths
in committed logs. These reports and review records now normalize that prefix
to `<USER_HOME>`; exact originals remain local. `review-completion.json` binds
original/normalized hashes. Internal review confirms no numerical/runtime result
changed, no path prefix remains and the finding is resolved. No second provider
review was needed for this formatting-only correction.
