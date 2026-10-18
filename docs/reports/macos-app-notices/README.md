# Binary-bound Mac app notice payloads

The Mac app packager now accepts `--notice-bundle` and carries the explicitly
selected original declarations and inventory into `Contents/Resources/THIRD_PARTY_NOTICES`.
Previously it copied no third-party notice material. The input bundle must name
the staged executable's exact pre-sign SHA256. The app manifest records that
binding, inventory hash and bounded file/component counts, retaining pending
rights/applicability status. This closes the notice payload transport gap; it
does not select applicable licenses or complete M6/Mac acceptance.

## Implemented behavior

The reusable `scripts/app_notices.py` helper admits a strict schema-1 inventory,
checks component/file references, exact file/directory coverage, safe paths and
case spellings across every directory prefix, regular files and symlink refusal.
It reuses the existing bounded regular-file reader and strict JSON parser.
Counts/declared totals precede payload reads; budgets are 1 MiB inventory, 1,024
files, 512 components, 4 MiB per declaration and 16 MiB including inventory.
The copied bytes come from the admitted bounded snapshot. Input root symlinks
and descendant symlinks refuse; external parent aliases are resolved. Sampling
is not filesystem isolation against adversarial concurrent directory changes.

The Mac packager checks executable bytes, exact app manifest bytes and optional
notice payloads before signing/publication and at final output. Explicit signing
may transform the executable: the returned JSON captures the post-sign identity,
while the manifest and notice inventory remain bound to the pre-sign source.
The unchanged expected manifest is checked after signing. Output directory
identity is checked around publication; cleanup preserves a detected replacement
rather than deleting it. These are sampled checks, not an adversarially isolated
publisher. If output creation succeeds but the initial ownership observation
fails, no payload moves or successful JSON occur; the unproven empty output is
preserved for inspection/manual removal. The new regression verifies that behavior
and existing-output refusal on retry. Existing non-atomic publication/completion
semantics still apply.

No UI runtime, Go API, prepared dependency, installed library, SDK or public
release gate changes. Notice support is Mac-only and optional; an older package
or an omitted option makes no notice-coverage claim. Both inventory and app
manifest keep pending audit status. No input declaration is treated as executable
instructions, upstream attestation or automatic rights approval.

## Regression evidence

The final focused suites pass **36 test methods**: nine notice-validator methods
and 27 app-packager methods, including the prior cross-platform/default-signing
regressions. Tests cover exact original-byte copying and relocation, executable
mismatch, corruption, malformed schema/duplicate JSON/nonfinite values, file and
aggregate bounds before payload reads, duplicate/missing component references,
extra/missing/symlinked inputs, file and directory-prefix case aliases,
validated snapshot behavior, pre-sign and final executable/manifest/notice
tampering, allowed simulated signing transformation, output replacement and
failure cleanup. Signing and native dependencies are mocked in these unit tests;
they do not execute an app or perform actual signing.

The first 30-method endpoint also passes, but internal review identifies missing
directory-prefix collision handling, incomplete executable/manifest publication
checks, and cleanup that could delete a detected replacement output. The
implementation and discriminating regressions are corrected before the
35-method pre-provider endpoint and real packaging exercise. A further ownership-
acquisition regression after Prism brings the accepted focused total to 36. Both source versions and raw test receipts are
preserved; no earlier result is substituted for the corrected source.

## Real private packaging exercise

The reviewed headless command packages a retained real Darwin ARM64 executable
in **0.276 seconds**, without launching it or requesting signing. The finished
private `VellumNotices.app` has **113 regular files / 60,486,868 payload bytes**,
normalized file/directory modes, valid bundle metadata and pending/unverified
manifest status. Original notice bytes and exact executable/manifest identities
match; the CLI result's packaged checksum equals the pre-sign source checksum.
The 125 controller input pins remain stable, logs are complete, the owned group
97693 is absent and its cleanup/observation error list is empty.

The executable is reused from the prior qualifying Go 1.25.6 owned build,
base `d3d4eb8` plus that checkpoint's collector changes. It is 60,138,530 bytes,
SHA256 `846ab24b0fdf9f7437fe29e9e684ab1545fb29a8db1f0d940973d0a5c2c660dd`.
This is no new compilation, final SDK refresh or native runtime qualification.
The packaging controller reuses the reviewed owned executor, a 180-second outer
watchdog, fresh output/stop scope, resolved Python, exact source/helper/release
pins, controlled environment and an IP-denied/Unix-IPC-allowed sandbox. Native
metadata inspection is allowed; no app window, desktop input or OS setting is
changed. Unit signing is simulated; no actual signing operation is performed.

The explicitly selected input has **109 original declaration files / 86
components / 346,810 bytes including its inventory**. Ninety-nine declarations
from 84 of the build inventory's 89 Go modules match retained private SDK notice
bytes, module versions/sums and file hashes. Five `webgpu/libs-*` graph modules
have no recorded declarations and remain manual-review gaps. Nine installed
Go 1.25.6 named ancestor candidates and Vellum's own declaration complete this
partial input. Both raw graph inventories, omissions and original paths/hashes
are retained. The full Go graph includes dependencies that need not survive
linking; this is conservative declaration transport, not exact linked membership.

The actual executable's selected native archive hash begins `7d8b688a`, whereas
the current private SDK's archive-specific source-built notice bundle binds
`786aaad8`. That mismatched bundle is deliberately excluded. This exercise does
not approve the old prebuilt archive or transfer the source-built archive's
notice claim to it. Actual native/font/generated/embedded/non-font attribution
and applicable notice completeness remain open.

The independently inspected `.app` stays at the local path in
`real-package-verification.json`. Its complete member/hash/mode inventory and
actual app manifest/plist are preserved in the evidence; the compiled binary is
not duplicated into the repository.

## Review and remaining work

Internal review covers source, corrected regressions and the headless packaging
controller. One Prism call with Anthropic Claude Sonnet 5.5 covers all eight
changed source/documentation files (49,920 bytes), with complete coverage. Its
raw CLI exit is 1, retaining two medium findings and discarding/suppressing none.
The early output-identity acquisition failure is accepted as an explicit
fail-closed limitation, with the additional passing regression and documented
recovery. The branch-local checksum concern is refuted: both assignment and
return read use the same unchanged Darwin condition; failures never return and
other platforms never read it. Internal closure checks those dispositions.
There is no production change after provider review, no second provider call and
no unnecessary repeat of the real packaging exercise. [Raw evidence](records.tar.gz) and its [file manifest](records.json)
retain source versions, regression logs, input origins, admitted original notice
bytes, packaging receipts and reviews. The compiled executable and duplicate
`.app` payload are excluded; identities and retained local paths are recorded.

The shipped-source/notice applicability audit remains open for actual native,
font, embedded, generated and non-font material. The prepared partial bundle is
an explicitly selected declaration collection, not the final application notice
inventory, independently attested provenance or redistribution approval. Final
SDK/app runtime acceptance, performance, resource/endurance, human/native/display
workflows and the final completion audit remain open.
