# macOS v1 private release closeout

Approved25 October2026: close CI, adopt the stable API/documentation contract and produce final macOS v1 SDK/apps, then relaunch Notes. Baseline is `b633872`, following the completed RC. Features, backend, rendering/pacing options and existing release responsiveness policy remain frozen.

1. CI checkpoint: resolve observed required-workflow failures with focused tests; preserve test selection and runtime assertions/deadlines. Review with Prism/Anthropic Claude Sonnet5.5 and commit. Final workflow success must refer to the final release source, not an earlier passing revision.
2. API/documentation checkpoint: adopt [stable API scope](stable-api.md), align current metadata/version/docs, retain named experimental APIs and hardware/human limits, run portable and affected consumer/tooling checks, review and commit. No API removals or new controls.
3. Packaging checkpoint: from clean committed source export a private SDK with the unchanged hash-pinned native capsule and notice inputs, verify independent consumer/starter origins and builds, package matching Notes/Mac Widgets/shell apps, verify exact ZIP bytes/signatures/checksums, record final CI, review and commit evidence. Save/close the existing Notes window normally, relaunch the verified final package and record its actual startup/saved-notebook result. Report-only commits do not relabel built bytes.

The deliverable is `v1.0.0` for private personal use on the recorded Apple Silicon macOS 26.6.2 configuration. Human IME and VoiceOver remain unqualified by prior owner choice; additional OS/hardware/displays and public distribution are outside this closeout. The changed-payload distribution audit stays pending. This private closeout did not authorize public tags/releases/assets. The later root source release is separately authorized below; native SDK/app assets remain private. Do not use a signing identity or notarize. Ad-hoc signing remains the private-package policy. No historical exact-payload audit approval is extended.

Reuse unchanged RC physical/input/history/soak evidence. Desktop actions are limited to the requested final app relaunch and app-local smoke; preserve notebook data, clipboard and global preferences. Actual new correctness defects block delivery; optional polish stays in the backlog. This contract governs current closeout; older candidate/preview phase documents retain their original scope and results.

## Source release follow-up — 26 October 2026

The owner requested a tagged release and README update after merging the showcase
apps. This separately authorizes a public root `v1.0.0` source tag and GitHub
release. It does not extend the private native payload's audit or authorize
SDK/app downloads, signing identities or notarization. The source release includes
the committed public checkout and the new demos; it does not relabel the older
private build9 packages as builds of that tag. Compatibility, experimental
exclusions and runtime qualification limits remain unchanged. Native-module
release coordination remains separate; native consumers use the checkout and
its explicit replacements.

## Frozen-feature correctness maintenance: window exposure

A visible window restored after a sibling closes must present its own retained
scene without requiring a document command or value edit. Coalesce exposure
requests, retain one while hidden/in flight, and return to unchanged idle after
presentation. Retire closed windows through the existing ownership guards. This
is a correctness repair within the freeze, not a new feature or pacing option.
Separate deterministic scene/composition regressions from actual packaged Mac
window observations; preserve historical binaries/audits and record any new
source, prepared-Go patch, executable and notice bindings explicitly.

## Coherent final public candidates — 27 October 2026

The owner requested a fresh SDK and all six apps from the same green source.
[Build13](../docs/reports/macos-v1-build13/README.md) pins clean 64bf9cf: the SDK
is freshly exported and every executable is compiled from its verified extraction.
Stage0108 uses ordinary preparation, with no manual overlay. Reuse green CI;
run independent SDK and affected exposure checks, exact package/notice guards
and bounded native render/close checks. Disclose new prepared-Go/binary identities
and retain original notice bodies and historical audit decisions.

The final packet proposes seven archives plus SHA256SUMS on a separate
macos-v1.0.0-build13 packaging tag/release at64bf9cf, matching the binary source.
The existing v1.0.0 source tag/release stays untouched at4b42fb8. This authorizes
local preparation/review/commit, not publication, new tags, audit rewrites,
signing identities or notarization. Features, stable API, backend, performance
gate and qualification limits remain frozen. Report-only commits never relabel
the build-source CI or artifacts.

## Approved publication — 27 October 2026

The owner approved the exact final packet and requested merge/push/publication.
The SDK and six apps are published under macos-v1.0.0-build13 at64bf9cf, using
the reviewed text and ad-hoc/unnotarized route. [Publication record](../release/macos-v1-build13-publication.json)
binds the eight public assets and actual download checks. GitHub normalized
three filenames' spaces to dots; checksum filename entries alone were adapted.
The seven SDK/app hashes and reviewed release text are identical to the packet.
Existing source tags/releases, audit records and historical preauthorization
reports remain unchanged. Only this payload is authorized; future releases,
identity signing/notarization and broader qualification remain separate work.
