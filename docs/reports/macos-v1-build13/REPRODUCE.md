# Focused reproduction

Use a separate clean checkout at 64bf9cfbb9cfb57d96e1aecd0adad01239ca3c7a,
Go 1.25.6, Python 3, patch and Xcode Command Line Tools on ARM64 macOS. Preserve
release/version.json and historical audit records. This later report commit
is not the product's source identity. Indexed evidence contains the actual
controlled-environment controllers/arguments and all attempt logs.

1. Export scripts/release-sdk.py --public with explicit --macos-gpu-library,
   --macos-gpu-sha256 and --macos-gpu-provenance. Pinned native archive:
   786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883.
   The old verified SDK supplies only this unchanged capsule; current Go source
   and prepared dependencies come from the clean source through normal preparation.
2. Run scripts/check-sdk.py with --archive and independently recorded --sha256;
   extract with checksum verification. In drivers/cogent/build/upstream run
   `go test -race ./core -run '^TestVellumExposure' -count=1`. Reuse green CI and
   unchanged input/history evidence rather than repeat broad qualification.
3. Build from the extracted SDK, -mod=readonly: examples/macapps ./cmd/notes,
   ./cmd/macshell; drivers/cogent ./cmd/gallery; examples/showcase ./cmd/forma,
   ./cmd/luma, ./cmd/pulse. Verify replacements resolve into that same SDK and
   actual link maps select its pinned native archive. Check unchanged module
   versions/sums, selected embeds, standard declarations, original icon/font bytes.
4. Preserve each corresponding original notice bundle. Rebind only artifact_sha256
   to the new pre-sign binary, after those comparisons. Keep the showcase Go-font
   notice. Mismatches must fail. Collection pending/review-required labels stay
   separate from the approved release/native audit guard.
5. Use unchanged scripts/package-app.py --public with explicit --binary,
   --notice-bundle, --version 1.0.0, --build-version 13, --sign-identity -, and
   the original generated --icon PNG for each showcase. Archive with
   `ditto -c -k --norsrc --noextattr --keepParent`; extract and run
   `codesign --verify --deep --strict`. Copy with ditto --norsrc --noextattr to
   a nonsynchronized temporary launch directory for exact-byte native checks.
6. Check source/pre-sign/signed/manifests/notice/archive hashes and Info.plist.
   Keep sealed target_runtime=unverified; record actual observations separately.
   Notes requires a disposable --data-dir and refusal of an existing Notes
   process. Never use the personal notebook for the smoke. Native Raise/render,
   normal close and exact process absence are bounded checks; reverify signatures
   and every SDK file afterward.

The local retained pre-sign inputs are pinned in private-inputs.json. They are
not download assets. Captured paths/timestamps identify this run. Fresh builds
in a different path/toolchain are not asserted bit-identical; generate a new
hash packet for different outputs. Only the present SHA256SUMS identifies these
proposed bytes. Nothing here uploads files, moves tags or changes release gates.
