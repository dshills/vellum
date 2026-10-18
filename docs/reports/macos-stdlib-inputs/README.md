# Go standard-library build input coverage

The distribution collector now records Go-selected runtime/standard-library
source, assembly, cgo, embedded and object files. Full standard-package metadata is resampled within each collection. The stored
selected-file identities, package/source lists and named ancestor declarations
participate in the existing owned compiler's before/after comparison. Source or declaration
drift refuses a successful build receipt. This closes a specific M6 input-coverage
gap; it does not approve redistribution or complete the Mac framework.

Previously both the package inventory and metadata comparison skipped every
package marked `Standard`. The recorded Go executable and root declaration
candidates did not bind those selected source bytes. `scripts/distribution-inventory.py`
now adds `go_standard_library`, keeping schema 1 additive compatibility and the
existing pending distribution status.

## Behavior and limits

The new record contains GOROOT-relative paths, byte counts and SHA256 identities,
package-to-source-field references, and named ancestor LICENSE/COPYING/NOTICE/
PATENTS candidates. Source metadata includes `.syso` objects and embeds as well as
Go, assembly and cgo files. It excludes ignored files. The selected source/entry
and declaration counts are bounded before surplus hashing; the existing 1 GiB
streamed per-file bound remains. Duplicate packages/source entries, incomplete
metadata, escaping paths and symlinks within GOROOT refuse. External installation
aliases, including macOS `/var` versus `/private/var`, remain supported; the
outermost matching installation prefix prevents an internal link back to GOROOT
from masquerading as a new root.

The collector rehashes these files and re-enumerates named declarations after a
second Go package query, so inserted declarations and changed source metadata
are detected. The owned build producer already compares full inventories before
and after compilation; its existing failure path now covers this section too.
No new build command, native runtime path, public Go API or approval bypass is
introduced.

These are the installed Go tool's selected package files, not authentication of
upstream Go, a complete compiler/native dependency audit, compiler-generated
source, transitive includes, external SDK headers or proof that bytes survive
linking. Nearby declarations remain candidates; embedded third-party source may
need additional notice attribution. This metadata neither copies applicable
notices into application bundles nor establishes rights approval. The root
public distribution gate remains pending. Older inventories without the new
section do not supply this coverage.

## Verification

The final focused inventory and owned-build suites pass 16 and 11 test methods.
Related SDK producer, extracted checker and app packager suites pass another
29, 24 and 16 respectively: 96 methods total. The build failure table includes
an actual `stdlib-drift` subcase, which must withhold a completion receipt.
Regressions cover selected source/assembly/embedded/object identities, root and
ancestor declarations, relocation, source and package-metadata drift, changed or
inserted notices, malformed/duplicate/over-budget inputs, and internal symlinks
including a loop back to GOROOT.

An initial unit run fails on lexical macOS installation aliases and is preserved.
Internal review subsequently finds an internal root-link bypass in the first
alias correction and a defined-but-unexecuted build failure variant. Both are
fixed before the qualifying compile; affected tests and the related packaging
suites are rerun on the corrected source. Review also tightens the headless
controller's SIGTERM cleanup, resolved Python executable and imported-helper
pins before the final compile. A preliminary successful compile is retained
separately and is not substituted for the final-source result.

The qualifying private precommit build uses Go 1.25.6 on Darwin ARM64 with cgo,
from base `d3d4eb8` plus the three tracked Python changes. It completes in
15.107 seconds. It records **236 standard packages,
1348 distinct selected files (13,972,226 bytes)** and **nine
named ancestor declaration candidates (20,738 bytes)**.
These include GOROOT LICENSE/PATENTS, the selected BoringSSL-directory declaration
and vendored x/crypto, x/net and x/text LICENSE/PATENTS files. They are declaration
candidates, not selected licenses or proof those components survive linking.

Before/after input inventories match exactly apart from the added linker
observation. The completion receipt binds the compiler's emitted binary/map,
serialized inventories and streamed logs. The final executable is
60,138,530 bytes, SHA256
`846ab24b0fdf9f7437fe29e9e684ab1545fb29a8db1f0d940973d0a5c2c660dd`.
All seventeen controller pins remain stable; the owned process group is observed
absent after terminal cleanup. The preliminary 33.665-second build used earlier
source and remains historical. The qualifying result makes no claim about a
later final SDK artifact or an executed Mac application.

No produced application is executed. The headless controller uses the existing
owned executor with an IP-denied/Unix-IPC-allowed sandbox, minimum Go 1.25.6,
controlled child environment, fresh output and stop marker, bounded logs and a
1,200-second outer watchdog. The producer retains exact compiler argv, Go
executable/environment identity, raw compiler logs, binary/linker-map hashes and
input inventories. These are observed same-build relationships and sampled drift
checks, not a hermetic build or independent attestation. Compiler diagnostics
remain in the raw logs.

## Review and remaining work

Internal review covers the implementation, corrected compiler-only controller,
regressions, actual outputs and final scope. One Prism review with Anthropic
Claude Sonnet 5.5 covers all seven changed source/documentation files (40,323
bytes), with complete coverage, one provider call and CLI exit 0. It retains no
findings, suppresses none and discards one declaration-symlink concern because
its quoted evidence does not match. That semantic concern is also checked:
`file_identity` calls the existing component-wise symlink-refusing helper;
three supplemental root/package/dangling declaration probes all refuse. The raw
candidate and disposition remain in the evidence. No provider rerun is used.
[Evidence](records.tar.gz) and the [per-file manifest](records.json) preserve
source versions, the initial failure, final tests, preliminary/final build
receipts and inventories, compiler logs and reviews. Compiled binary and linker-map payloads, the Go/Python executables and the SDK
itself are not duplicated in this archive; their identities remain in the
receipts and payloads stay in the local owned build/toolchain/artifact paths.

This advances M6's input accounting only. Applicable notices, exact linked
Rust/native/generated/font/non-font rights and the owner distribution decision
remain open, along with final SDK/app-bundle acceptance. The unchanged M3
performance failures, native M2, complete M4 resource/endurance gates, M5 human/
display/input workflows and M7 audit still prevent full Mac completion.
