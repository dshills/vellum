# Final local macOS v1 public packaging and launch checks

The owner's 27 October 2026 exact-payload audit approval is recorded in
[the release-specific decision](../../../release/macos-v1-distribution-audit.json).
Final public-mode production succeeded for the SDK and all six audited apps.
**The SDK and five apps are recommended for a separate publication decision.
Mac Shell is held for a dark-appearance contrast defect. Nothing was uploaded.**
Features, APIs, backend, rendering/pacing defaults and the release gate are unchanged.
No signing identity, notarization, global setting or user notebook change was made.

## Artifact identities and disposition

Frozen product source/tag: `4b42fb89948b8abf0f2f65970fb5980987c1af0c` / `v1.0.0`.
Packaging/approval source: `1fc594b6952d3b072a4c72b619a6f259da389474`, clean.
That later commit changes approval/docs only. The source tag is unchanged.
Final SDK SHA256: `0519cc8645616ac14489ee69d94efb07ab8dcccc1e56ed1f4be14976daa33ab3`.

| App | Original notice files | Render/close outcome | Final ZIP SHA256 |
| --- | ---: | --- | --- |
| Vellum Notes | 418 | Pass | `3d60955b9336bcdfe931ca68222057cb4d36022f44b37ec74c4515cd8693fc25` |
| Vellum Mac Widgets | 418 | Pass | `c9a3b636faa8cb24498a6f7f09f180d4811e407dbe3b89489eb11611f6999ca7` |
| Vellum Mac Shell | 418 | Held: dark contrast | `7d3ad7aaefe1d19cab5e1b2dcf9b20c7c1e690b08d1fcfd99c418ee26af1db7f` |
| Forma | 419 | Pass | `cc9d994a3322a04412b891b5762c6e2ef6230500faeec5b7f812764a7a5285ec` |
| Luma | 419 | Pass | `291753ff5c401cc079a9d96cf837f76611e6b9914c99a7c0e24bf206b19391f1` |
| Pulse | 419 | Pass | `1a03c26bcea3ff8d81a15377d548aa627462ea4d7b1d0ad7b65866bac005a1e6` |

[All produced archive hashes](SHA256SUMS) include the held fixture for traceability.
[Recommended archive hashes](RECOMMENDED_SHA256SUMS) exclude Mac Shell.
[Production record](package-result.json) retains pre-launch status; the separate
[post-launch record](postlaunch-verification.json) records final observed outcomes.
The delivery manifest outside Git also has those outcomes. Signed executable
hashes differ from the audit's pre-sign hashes; both are retained and checked.

SDK/app archives and a short delivery README are retained outside Git at the
research workspace's `deliveries/macos-v1-public-candidates-20261027/`.
ZIP/tar archives are the durable handoff; launch copies under `/private/tmp/`
are disposable and must not be substituted for the recorded archive bytes.

## Focused automated evidence

The unchanged public SDK/app guards accepted the approved audit, clean committed
source, verified dependency/native payload and exact executable-bound notices.
The final production record contains 38 successful command steps: SDK export,
independent SDK consumer, and six app production/ZIP/extraction/signature/copy
sequences. All seven archive hashes were rechecked after native checks.
The local [packet checker](verify-packet.py) also compares the production,
post-launch, README-table and checksum identities, verifies the held-app exclusion
and hashes every evidence member. Run `python3 verify-packet.py --assets /path/to/candidates`
from this report directory to include the actual seven candidate archive bytes.

The public SDK verifies all 8,856 listed files. Its independent relocated
consumer passes race/vet/build and public API scenarios with replacements inside
the extracted SDK. Native runtime is deliberately not executed by that SDK check.
[Product equivalence](public-product-equivalence.json) compares the public SDK to
the approved private audit candidate: only approval/release/source documentation,
public markers and the inventory's native-audit field differ. Code, dependencies,
prepared native files and assets match. The native archive remains `786aaad8…`.

All six ZIP extractions and launch copies pass strict deep ad-hoc signature
validation. After closing, all six signatures, signed executable hashes,
manifest hashes and original notice inventories pass again; all six exact launch
process paths are absent. Exact notice hashes still match the approved packet,
including the added Go-font family notice in the three showcases. Public manifests
carry `private_build=false` and `native_distribution_audit=approved`.

Notice-discovery labels `pending`/`review-required` and sealed producer
`target_runtime=unverified` remain unchanged. They are collection-time declarations,
not rewritten to assert human review or runtime results. The separate owner audit
and this observed runtime record supply those distinct decisions/evidence.
Existing [exact-product-source CI](../macos-v1-distribution/reused-ci.json) is reused;
no new broad campaign or claim of CI on the later local packaging commit is made.

## Native observations

These are bounded local launch/render/native-close observations on this Apple
Silicon macOS host, using the exact ZIP-extracted signed bytes. They do not assert
full physical-input, performance, accessibility or download/Gatekeeper acceptance.

- Notes: native toolbar, sample list and proportional prose rendered; native close
  retired the process. A fixed, digest-checked, app-restricted launcher used a new
  disposable data directory. The personal notebook was not opened or changed.
- Mac Widgets: catalogue/theme controls and native toolbar rendered readably in
  the observed dark appearance; native close retired the process. No appearance
  switch or fresh full catalogue interaction pass is claimed.
- Forma: poster GPU surface, three layers and inspector rendered; native close
  retired the process.
- Luma: landscape, six-image filmstrip and histogram rendered; native close
  retired the process.
- Pulse: metrics/chart, selected service row and inspector rendered readably in
  dark appearance; native close retired the process.
- Mac Shell: both independent native windows and toolbar labels appeared, and
  closing both retired the process. Body labels were nearly black against the
  dark native background. **Visual acceptance failed; hold this app ZIP.**
  [Reproduction and smallest next action](SHELL_CONTRAST_ISSUE.md).

Raw AX trees and unedited screenshots are inside [evidence.tar.gz](evidence.tar.gz),
with [member hashes](evidence-index.json). They contain only generated fixtures,
local paths and checkpoint metadata. Native close success is observed process
retirement; no uncollected process exit code is reported as zero.

One computer-use inspection call for Widgets stalled for about 73 minutes before
returning. That is a tool inspection delay, not a measured app launch time or
framework responsiveness result. After it returned, the remaining checks completed.

## Controller corrections and remaining limits

Two local reporting-controller attempts stopped while parsing producer stdout:
dependency-patch progress before SDK JSON, and icon conversion output before app
JSON. Both actual producers had exited successfully. Already-produced archives
were verified/reused; parsing was corrected only in the disposable controller.
Original attempts/logs are preserved. A final reporting assertion initially compared
an annotated tag object to a commit; dereferencing `v1.0.0^{commit}` fixed the check.
No producer, signature, audit or dependency verifier was weakened.

The apps are ad-hoc signed and unnotarized, ARM64 only. No downloaded-quarantine
launch or other host/display qualification is claimed. Human IME and VoiceOver
remain explicitly unqualified. The owner's decision approves the identified
payload, not a new upstream license grant; the prior pinned objc2 caveat remains
in the audit packet. [The final publication packet](PUBLICATION_PACKET.md)
identifies exactly what can be proposed next. No tags, assets or release text were
published, and no Developer ID credential or Apple submission was used.
