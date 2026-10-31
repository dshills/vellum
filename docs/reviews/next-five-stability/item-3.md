# Item 3: repository-owned version-aware demo suite

Migrated the existing external demo hub into examples/demo-suite as an independent
public-API consumer module. Relative module replacements and a checked build catalog
replace machine-specific paths. The Go hub uses vellum.Version; Python requires
matching release metadata/Version and the root go.mod toolchain. Data repositories
for Documents/Notes remain isolated by source version and feature.

Build output is exclusively reserved outside the checkout. Failed builds do not
promote a new selection. A clean source commit is checked before compilation,
after compilation and after archive verification. All catalog entries are built;
signed metadata and executable hashes are verified in the working suite and a
fresh extraction. DELIVERY binds the selected archive/manifest hashes. Launch
checks that identity and restores only the same hash-checked archive if temporary
app files expire. Successful restoration is remembered. This is local integrity,
not independent authenticity, complete dependency attestation or notarization.

The source version is not a tag claim: a later development commit may still carry
the same API version. Old private suites and their historical evidence remain
unchanged. Nothing publishes native assets, enables automation or requests native
service permission automatically. The existing opt-in services remain explicit.

Validation before commit: hub offscreen race tests and vet passed. Six Python
delivery checks cover metadata mismatch, existing-output preservation/alias refusal,
changed archive refusal, changed executable refusal, restoration persistence and
manifest path-component rejection. Catalog agreement is checked
by the Go tests. Shell syntax passed. The stage-dispatch fixture was extended for
the new module; both dispatch tests pass. Normal example and offscreen check stages
now include the hub. Full source-bound packaging requires the reviewed clean commit
and remains item 3's closing check before item 4 starts.


## Review dispositions

The initial full review and a combined tooling retry exceeded Prism's default
16384-token response limit; neither is counted as completed coverage. File batches
use the same default Wafer / GLM-5.3 reviewer. Successful receipts and the hub's
incomplete coverage receipt are retained; skipped files are separately reviewed.

The documentation-only batch's high missing-implementation finding is rejected:
all referenced Go/module/tool files are staged in this same checkpoint, verified
with git ls-files --stage. The filtered batch omitted their diffs; it did not
represent a separate commit. The module reads the repository's release/version.json;
it does not need a second version metadata file inside the example. The license
year finding is rejected: the file is byte-identical to the root 2026 MIT license,
and the current environment date is in 2026. The untagged coverage concern is
resolved by an actual passing untagged hub race run; these tests have no offscreen
build constraint, and offscreen tests also pass.

The launcher path finding is addressed with absolute paths and a directory check
before restoration. Selection publication uses a unique temporary file and atomic
rename; source launch reads that same JSON selection rather than a second pointer.
The helper path-component finding is addressed before constructing bundle paths,
with a rejection test. The view-validation test now reports every failing view.
The suggested Close-error handling is rejected: app.Application.Close returns no
value (app/app.go), so the proposed change would not compile. Generic service
negative validators belong to their app package tests (services/notifications/tray);
the consumer test checks its actual valid payloads rather than duplicating those
implementations or adopting the review's invalid examples of legal payloads.


The long suite launcher was ultimately split into ordinary Go compilation units
for identity, catalog, feature arguments/launch, automation process lifetime,
replay results and UI composition. Final coverage uses staged-file reviews of
those units, not standalone excerpts. The superseded excerpt review's unused-import
finding is not a production compile failure; both complete-package build modes
passed, and the final units have only their actual imports.

Later low findings were addressed by a source-run working-directory hint, explicit
launch policies for every catalog entry (with a coverage/isolation test), truthful
error/success handling, a documented helper-only cancellation scope, reset modal
state, complete view diagnostics and bounded Unicode replay failure detail.
The worker review's two medium findings are fixed: macOS automation children get
an independent process group and their exit failures are retained. Each launch
uses a distinct mode-0600 log; an actual exit-7 helper test confirms separate logs
and recorded status. The final worker review completed with zero findings.
Untagged and offscreen hub race tests, offscreen vet, six Python delivery tests,
shell checks and stage-dispatch tests pass. No actionable medium/high finding is
left unresolved; raw scope-related false positives remain with their dispositions.
