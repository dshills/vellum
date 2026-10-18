# Current private SDK: minimum-Go relocated consumption

A fresh private SDK from clean `abe67c6` includes M1 child cancellation, stage0079,
the adopted grid/text/accessibility fixes and the source-built ARM64 GPU notice
bundle. Relocated generated-consumer race/vet/build and public-contract race/vet
pass using actual Go 1.25.6. All 26 discovered default-build main packages across
13 example/driver modules compile; the explicitly `vellum_native_checks` tagged
window-cycle helper also compiles. All 27 executable metadata records identify
Go 1.25.6, Darwin and ARM64. No native framework application is launched.

This closes current private SDK incorporation and these scoped compilation checks.
Final native SDK acceptance, other M6 requirements, public distribution and full
Mac completion remain open. Earlier SDK native checks are separate historical
evidence from source `5eef841`; they are not relabeled as this SDK's execution.

## Artifact and independently consumed inputs

The archive is 29,816,117 bytes, SHA256
`0bce4f09f8a51a7b196b2bba57c92d1c3fc4eaa999093fdef5c50edbea7bb7f9`.
It contains 7,879 hashed payload files plus its manifest. A standalone copy is
preserved locally at
`artifacts/macos-sdk-stage0079-abe67c6-cache-v2/deliverables/vellum-v0.1.0-alpha.2-macos-arm64-go1.25.6-sdk.tar.gz`
in the research workspace. This regular-file hardlink preserves the produced
archive outside temporary storage without inserting another SDK into the evidence
archive. It is private, caller-produced and digest-verified; no independent
artifact authenticity or redistribution approval is claimed.

All payload hashes are rechecked after consumption. At the recorded terminal
snapshots, both checkouts remain clean at `abe67c6`, with 5,288 tracked files and
the same 6,260-file stage0079 vector. The SDK has 7,763 equal source/prepared
counterparts and eleven declared, hash-verified Go module overlays. Its M1
`child_process.py` exactly matches current committed source. Compared with the
7,441-file stage0077 SDK, 438 files are added, none removed and 34 changed;
this includes the native component bundle, current fixes and acceptance contracts.
The separate `sdk-delta.json` preserves every changed member.

The exact source-built ARM64 static archive remains 37,401,760 bytes, SHA256
`786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.
Its 82 component entries and 409 inventory/notice files are now present in the
SDK, in addition to the 89-module Go notice inventory. This conservative compiled
component superset does not determine all linked runtime components or select
licenses. Apple-derived declarations, generated code and complete asset rights
remain review items. The installed native libraries are unchanged.

## Checks and coverage

The existing producer and SDK checker run against a digest-pinned archive. The
checker extracts to a path with spaces, generates a consumer, verifies its module
cache and confirms all five framework/native replacements reside in that SDK.
Consumer race tests, vet and native compilation pass, followed by the independent
public API contract's race tests and vet. Package production takes 5.768 seconds;
the relocation command takes 65.957 seconds. Durations describe bounded checks,
not startup/frame latency.

A retained extraction is then checked module by module. Public-only automation
and SDK-contract modules require the exact SDK public root; modules with native
dependencies require the complete exact Cogent/typesetting/GPU replacement set.
Every selected replacement is inside the extraction. The fixture lists active
main packages, builds their executable set and inspects each file's compiler/path
metadata. Coverage is all discovered default-build mains plus the one named
build-tagged helper, not every possible build-tag configuration or non-main package.

| Group | Modules | Default mains | Additional tagged mains |
| --- | ---: | ---: | ---: |
| Driver | 1 | 15 | 1 |
| Runnable examples | 11 | 11 | 0 |
| SDK-contract (test-only; race/vet already pass) | 1 | 0 | 0 |
| Total | 13 | 26 | 1 |

The thirteen modules are 1 driver + 11 runnable examples + 1 SDK-contract.
Default mains are 15 + 11 = 26; the helper is inside the existing driver module,
so it adds one executable, yielding 27, without adding a fourteenth module.

All 71 completed owned commands have zero exit, complete bounded logs, stable
pins and observed process-group absence. This count includes three reused
successful driver graph/list/build commands and the two font inspection commands.
The private SDK functional checks use the reviewed shared-group host for nested
Python CLI commands, with unconditional group retirement by the outer executor.
They do not independently requalify standalone SDK cancellation behavior; M1's
source acceptance remains its separate evidence. Go compile/metadata commands use
the same owned executor directly.

Production SDK child environments restore public proxy/SumDB configuration, while
the enclosing sandbox denies IP access and permits Unix IPC. Checks are backed
by the explicitly bound populated Go cache. No network warming occurs. This is
not a claim the unmodified SDK scripts expose an offline configuration mode.

## Shipped font mapping

The SDK contains 28 font files with 28 distinct hashes: 24 Noto Sans/Roboto Mono
files across desktop/JS variants and four typesetting test assets. Actual Go 1.25.6
selected-package metadata identifies twelve default Darwin `EmbedFiles`. The
other sixteen files still ship in the SDK and require attribution review.

A pure headless metadata utility uses the exact SDK OpenType parser, verifies
each font hash and reads name-table IDs 0/1/2/3/4/5/6/13/14/16/17. All 28 fonts
parse without error. Original selected name/copyright/version/license fields and
nearby declaration/README hashes are retained in `font-result.json`. These are
selected decoded records, not an exhaustive enumeration of localized name entries.
Thirteen faces have empty license-description and license-URL fields; absence
from these fields is an attribution review item, not proof of missing rights.
This includes the Ubuntu Mono test font. Do not assume the module's own code
license covers every bundled font.

This mapping establishes bytes, embedded declarations and selected Darwin embeds;
it does not approve redistribution, prove historical subsetting provenance,
identify the actual Face selected for arbitrary application text/system fallback,
or inventory all non-font/generated assets. The utility performs no system-font
scan, cache update, native application launch or preference change.

## Failures and review boundaries retained

The initial plan is rejected before execution for nested process-group ownership:
an outer immediate kill could bypass the SDK CLI's cleanup of separate nested
sessions. The corrected plan reuses existing reviewed shared-group hosting.
The first actual package attempt then fails before archive creation: its controlled
environment omitted configured GOPATH/GOMODCACHE, so Go used an empty default
cache and the sandbox refused proxy access. Its group 42659 is recorded absent.
A fresh continuation explicitly binds the existing cache and succeeds; it does
not remove the network restriction or overwrite the failure.

Two main-inventory corrections precede execution: public-only modules need no
native graph, and zero-main modules need no executable-output build. The first
actual inventory attempt still fails after three successful driver commands:
`go list` names a test-only `window-cycle-smoke` package `main`, but it has no
active Go/Cgo source in the default configuration. The Go build correctly emits
fifteen executables. A fresh continuation filters actual source-bearing mains,
reuses those exact fifteen binaries and three completed commands, then performs
remaining metadata/example builds and the separately tagged helper build.
The original controller failure remains failed. Neither failure is a product
compile failure. Existing Cocoa deprecation and duplicate objc linker warnings
remain in raw logs; these checks are not warning-free.

Internal reviews cover source/admission, corrections, completed results and the
final archive. One fresh Anthropic Claude Sonnet 5.5 Prism review covers all four changed
documentation files in one complete chunk (14,898 reviewed bytes). The raw CLI
exit is 1: one retained medium finding, zero discarded and zero suppressed.
Finding `c73d91c9d96bc163` requests a clearer module/main reconciliation; the table
now states module counts and 26 default + 1 tagged = 27 executable coverage
explicitly. Independent internal review closes this documentation finding; no
implementation change, repeated compilation or provider rerun is required. Evidence records original failures,
plans, exact pins, all compiler/module/log records, SDK delta, font mapping and
post-check identities in `records.tar.gz` with per-member `records.json` hashes.
SDK and executable binaries are separately fingerprinted and excluded from that
archive. Source `abe67c6` remains the artifact's identity after this documentation
commit. Final SDK native execution, human/input/display/service workflows,
performance/memory/endurance, rights closure and M7 audit remain open.
