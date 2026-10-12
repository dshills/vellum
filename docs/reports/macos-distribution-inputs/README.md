# Mac distribution input checkpoint

The new collector records actual SDK source files and selected workbench build
inputs without launching an application. It provides reproducible evidence for
the still-pending distribution audit; it approves no redistribution rights.

The final capture used macOS arm64, cgo enabled and Go 1.27.2. Two corrected-source
runs produce byte-identical JSON SHA256 `a22aec2c66aa03fd71b84f0caff43321bcc5d35df04b238c672c895c7a0c8358`.
Capture source is base 777afeb with private precommit changes and explicit
untracked inclusion. It records 6793 SDK files, 232 nonstandard selected packages,
4438 resolved embedded assets (including 12 fonts) and 89 module notice records.
The complete 6172 prepared files remain unchanged from the prior checkpoint.

The one native linker-search file candidate is Darwin arm64
`libwgpu_native.a`, 34356312 bytes, SHA256
`7d8b688a6b55f0728c847e1c3b028d07ccb1009d8cea4833a6186a5775384248`.
This is a conventional file in a declared module-owned search directory, not
proof of final linker selection or the archive's embedded Rust/C component
closure. The captured build has no unresolved native options in the collector's
recorded syntax. Framework references retain unverified resolution.

Five downloaded library modules lack their own conventional root/license-directory
declaration: WebGPU libs-android, libs-darwin, libs-ios, libs-linux and libs-windows.
All appear in the Go notice graph; only the Darwin archive is a native file
candidate for this target. Nearby notices are candidates, with all asset/native
rights remaining review-required. This inventory does not interpret those terms
or resolve native provenance. Namespace/MIT choices and release gates are unchanged.

Nine inventory fixtures pass, covering relocation/path aliases, exact identities,
pending status, missing/external/direct/unsupported link inputs, source-byte/set
and build-metadata drift, pre-hash entry admission, streaming notice refusal and
exclusive output. Existing SDK 28, SDK checker 20 and app packaging 16 tests pass.
Real collection, exact repeat and current source-file identity checks pass.
These are tooling checks, not application/native GPU or performance acceptance.

Internal review found and corrected two admission issues: selected source counts
were checked after hashing, and inherited notice collection loaded unused full
contents. Admission now precedes surplus hashing, and metadata-only notices
stream 64 KiB chunks with a 1 GiB file bound and retain no content copies. The SDK
producer's default notice-copy semantics stay fixed. Initial tests also exposed
path-alias/test-stream issues; the initial real run exposed Homebrew's Go LICENSE
placement in GOROOT's parent. Corrected identities and declaration candidates
pass; failed receipts remain archived.

One Prism review used Anthropic Claude Sonnet 5.5 with complete seven-file,
three-chunk coverage and reported one medium finding 4120b84c2f4db58d. Its raw exit 1
is preserved. The correction adds direct native file candidates and explicit
unresolved records for unknown/-Wl/-F, external searches and missing candidates;
framework names no longer imply system ownership. Independent V3 source review
approved that exact correction and finding closure. No unchanged provider rerun
was performed. V1 NEEDS_FIX and V2/V3 approvals remain distinct records.

Live windows/input/performance tests stay deferred until the user says the
computer is free. No app executed, OS preference changed or public artifact
published. Complete Mac acceptance, target performance, signing/notarization,
final native dependency closure and distribution approval remain open.

[Checkpoint](records/checkpoint.json), [raw evidence](records/evidence.tar.gz)
and [archive index](records/archive-index.json) bind the static observations.
