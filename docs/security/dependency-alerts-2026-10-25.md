# Dependency alert applicability — 25 October 2026

GitHub reported four open alerts on main at 39250c6. This checkpoint changes
only the disposable Gio prototype dependency graph; the active driver, native
archive, original notice bytes and historical build/evidence hashes stay intact.

| Alert | Manifest | Decision |
| --- | --- | --- |
| [3: x/image TIFF PackBits](https://github.com/dshills/vellum/security/dependabot/3) | prototypes/p0/gio/go.mod | Upgrade x/image from 0.26.0 to 0.44.0 (patched threshold 0.41.0) |
| [1: x/image TIFF allocation](https://github.com/dshills/vellum/security/dependabot/1) | prototypes/p0/gio/go.mod | Same upgrade exceeds patched threshold 0.38.0 |
| [2: x/net HTML parser](https://github.com/dshills/vellum/security/dependabot/2) | prototypes/p0/gio/go.mod | Upgrade x/net 0.48.0 to 0.57.0 (patched threshold 0.55.0) |
| [4: rand logger soundness](https://github.com/dshills/vellum/security/dependabot/4) | release/native-components/wgpu-darwin-arm64/NOTICES/itertools-0.13.0/Cargo.toml | Copied original upstream dev-dependency declaration, not a maintained Cargo build manifest |

The first three alerts are applicable to running the old prototype, which is
excluded by the SDK exporter. The maintained driver already uses x/image 0.44.0
and x/net 0.57.0. The prototype update raises its module floor to Go 1.25 and
updates x/sys/x/text as required; it adds no driver or framework dependency.
Old P0 evidence remains a record of its original pins, not this patched build.

For alert 4, the original itertools manifest declares rand 0.7 under
`dev-dependencies`. The hash-bound native inventory lists 82 components and no
rand component. That supports classifying this detected manifest reference as
notice material, rather than proof that the SDK builds that dev dependency.
It does not establish exhaustive binary reachability or a blanket native
security audit. Do not rewrite upstream notices or dismiss the alert solely to
make the repository display green. The alert remains open with this disposition.

Verify patched Go pins with module verification, prototype tests/vet/build.
Preserve original notice hashes against INVENTORY.json. Remote alerts update
only after the relevant fix reaches the default branch; a successful local
check does not assert that GitHub has closed an alert.
