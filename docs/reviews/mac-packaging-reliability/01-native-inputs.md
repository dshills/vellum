# Checkpoint 1: selected native build inputs

Internal review, 30 October 2026. Developer packaging provenance only. Features,
backend, release/performance policy and desktop deferral are unchanged.

Preflight now captures effective Go cgo compiler/flag settings, PATH-resolved
compiler drivers, xcrun-selected C/C++/linker binaries with hashes/version output,
SDK path/version/SDKSettings hash and controlled deployment overrides. Compiler
-### probes record default ARM64 deployment and planned commands without compiling,
linking or writing an application. Re-capture before completion detects selection
or sampled identity drift; existing Go/prepared-source checks remain intact.

Review correction: a compile-and-link dry run emits randomized temporary object
names. Changed the link-only probe to treat /dev/null as linker input, then checked
two actual captures are identical. Synthetic tests cover flags, single-driver
admission, SDK hash changes and drift refusal. No wrapper is silently attributed
as an underlying compiler. No environment/global Xcode selection was changed.

Scope is selected tools and SDK settings, not all headers/runtime libraries,
hermeticity, source/binary authenticity or a claim that default deployment proves
actual executable deployment. Subsequent artifact inspection supplies that fact.

Checks passed: native_toolchain_test.py (2), private_bundle_preflight_test.py (6),
private-macos-bundle_test.py (9), stage dispatcher (2), whitespace and two actual
local dry-run captures (SDK 27.0/default 27.0.0). No app executed or rebuilt;
installed build24 remains unchanged. Actual capture retained in private artifacts.
