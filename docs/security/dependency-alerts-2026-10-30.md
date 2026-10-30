# Dependency applicability recheck — 30 October 2026

This is an offscreen maintenance recheck of GitHub alert 4, not a native security
certification or a change to the approved distribution payload. The earlier
[25 October disposition](dependency-alerts-2026-10-25.md) remains historical.

## Finding and action

GitHub's API still reports [alert 4](https://github.com/dshills/vellum/security/dependabot/4)
as open, low severity, development scope, for `rand` in the original copied
`release/native-components/wgpu-darwin-arm64/NOTICES/itertools-0.13.0/Cargo.toml`.
Its `rand = 0.7` declaration is under `dev-dependencies`, not Vellum's active Go
module graph or a Vellum-maintained Rust build recipe. These notices are copied
original declarations; packaging does not compile their development dependencies.

[The upstream advisory](https://rustsec.org/advisories/RUSTSEC-2026-0097.html)
describes reentrant random-number use from a custom logger with the relevant
logging/thread-RNG features enabled. [The upstream fix](https://github.com/rust-random/rand/pull/1763)
removes the problematic logging dependency. Merely editing the copied declaration
would neither repair a binary nor preserve the original notice inventory.

No dependency upgrade or native archive replacement is justified by this detected
notice reference. Keep the alert open with its applicability recorded; do not
rewrite notices, suppress scanning, dismiss the alert, or change audit approval.
The new `scripts/dependency-disposition_test.py` guard rechecks the exact current
inventory, original declaration and absence of rand in that recorded component
list. A changed archive/inventory/declaration requires a fresh review; the guard
never auto-approves new evidence. It runs in the existing tooling CI stage.

## Reused evidence and its boundary

- Inventory SHA256: `443e1760808a72d4fee88330c58271e18eb87eea633d30ef180d819c6254f55e`.
- Original itertools Cargo declaration SHA256:
  `49abb2101a0dd9cb137df206454b6620d04929a4975921fab6682ba834435620`.
- Recorded source-built GPU archive:
  `786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.
- All 408 original notice/declaration files still match the inventory. Its 82
  component records contain no rand component (80 compiled packages plus the
  retained header/toolchain declarations).
- Re-read the existing source-built GPU evidence archive without extraction or
  execution. Build receipt, analysis and Cargo.lock hashes match their recorded
  identities. The lock contains no rand package; the actual 80-package compilation
  list contains no rand, and its static archive digest matches the inventory.

This conclusion is specific to that recorded archive. Private build23 has no
source-built `native_gpu` SDK overlay; it uses the existing separately pinned
`github.com/oliverbestmann/webgpu/libs-darwin` module
`v0.0.0-20260628152755-66a5dfa57f8d`. Its current cached ARM64 archive hashes to
`7d8b688a6b55f0728c847e1c3b028d07ccb1009d8cea4833a6186a5775384248`, which is a
different payload. Do not transfer the source-built archive's component-absence
finding to that opaque prebuilt library. This checkpoint does not establish its
complete Rust graph or exploit reachability. If a concrete vulnerability is
reported against that payload, investigate its producer provenance and component
graph separately; the notice alert is not evidence that the prebuilt library is
vulnerable or safe. No new runtime-security claim is made for installed build23.

## Verification

The focused disposition suite passed three tests, including changed component,
artifact and declaration cases that require reassessment. The existing notice
verifier checked every original file. The stage-dispatch suite passed two tests.
Historical lock/compilation evidence was checked separately; it is deliberately
not a new SDK runtime/test dependency because historical reports are excluded
from exported SDKs. No Rust build, Go pin change, binary replacement, notice edit,
remote alert mutation, application launch or desktop check occurred.
