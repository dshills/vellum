# Minimum Go source compatibility on Mac

Clean `340faf0` passes the scoped source compatibility check with the actual
**Go 1.25.6 darwin/arm64** compiler declared by the modules. This covers root
framework and public-contract race/vet tests, the complete nested driver's
offscreen race/vet suite, and a real native loaded-grid application build.
The resulting binary's embedded metadata reports `go1.25.6`, `GOOS=darwin`,
`GOARCH=arm64` and `CGO_ENABLED=1`. It is **not launched**.

Recent Go 1.27.2 and legacy-JSON-experiment checks did not establish this
minimum-compiler result. The new result is source/build compatibility for these
cases; it does not qualify all example modules, a rebuilt/relocated SDK, native
runtime, physical interaction, performance, memory/endurance or redistribution.
No full M6 or other Mac completion gate closes.

## Method and results

The installed Go 1.27.2 launcher obtains the exact `go1.25.6` toolchain through
Go's checksum-verified download mechanism, with `GOSUMDB=sum.golang.org`.
The recorded Go toolchain zip checksum (`ziphash`) is
`h1:6MNqR27QsE1copdkh66OXB/euDBMiS+ifCFva/AuCJ0=`; the executed Go binary SHA-256 is
`9c35eaeb87fd3df884c3540b25f3556902c164d88afe8a88e62ceb819d0da211`.
Subsequent checks invoke that binary directly, with automatic upgrades disabled
(`GOTOOLCHAIN=local`) and network dependency fetching disabled (`GOPROXY=off`).
The accepted 92-module graph declares no Go version above 1.25.6.

| Check | Result | Seconds |
| --- | --- | ---: |
| `root-race` | PASS | 19.613 |
| `root-vet` | PASS | 4.000 |
| `contract-race` | PASS | 1.817 |
| `contract-vet` | PASS | 0.167 |
| `driver-race` | PASS | 344.742 |
| `driver-vet` | PASS | 30.465 |
| `native-build` | PASS | 37.106 |

Root `./...` excludes nested modules, so the driver and public contract are
checked explicitly. Native-driver checks use the existing private module file,
stage0079 prepared source and static ARM64 WGPU capsule; no row-fragment overlay
or diagnostic binary is involved. Exact command vectors, cwd, source/compiler/
manifest pins, output logs and terminal environment annotations are retained.
Effective test/build settings include `GOMAXPROCS=2`, `CGO_ENABLED=1`, `GOENV=off`,
`GOWORK=off` and read-only module resolution. These are compatibility checks,
not ordinary latency or allocation measurements.
Existing Cocoa deprecation and duplicate-library linker warnings are retained
in the driver logs; they do not make either command fail.

Both checkouts were clean at the recorded terminal snapshots of `340faf0`. Their 6,260 prepared files still have
aggregate `1d9861ac10a0d83f607a36fd70dc79f00ce51d427c792b30b6a1331c0c3451ba`.
The private 37,401,760-byte static archive remains
`786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.
The compiled consumer is 59,873,442 bytes, SHA-256
`46f9517f4928f6f9599234142376f185b9ee328645a5fac7809a5901cdcea92c`. No source or prepared dependency changes are needed.
No executable, SDK, toolchain or module cache is duplicated into this checkpoint.

## Evidence and remaining work

[records.json](records.json) hashes every member of [records.tar.gz](records.tar.gz).
`assessment.json`, seven command receipts/log pairs, terminal annotations,
toolchain download/cache identity, the module graph and embedded binary metadata
record the actual outcomes. Internal reviews cover scope, commands, terminal
identity and packaging. This evidence-only checkpoint changes no implementation;
it reuses the existing implementation reviews without another Prism call.

M6 still needs the final clean SDK with current cancellation and stage0079
changes, relocated consumption/all-demo checks, required native consumers, exact
runtime/font/asset rights closure and owner decisions for any public distribution.
Signing/notarization/publication is not performed. M3 still needs a separate
architecture decision against unchanged frame targets. Native M2, M3–M5 and M7
remain open. The failed placement operator and rejected row-cache experiment
stay sealed; neither is retried here.
