# Checkpoint 2: independent existing-delivery verifier

Internal review, 30 October 2026. Developer tooling only; no framework/app behavior
or packaged payload changed. Desktop checks remain deferred.

Reviewed trust boundaries, bounded reads, exact archive names and complete app set,
owned archive snapshots, SDK all-file/source verification, receipt/source/hash
linkage, installed tree equality, preflight source/hash, app plist/manifest identity
and strict codesign on both extracted and installed apps. The verifier does not
call the producer, installer, compilers, app binaries, LaunchServices or notebook
readers. Existing bounded archive/signer helpers are reused. No metadata is trusted
as permission to follow an arbitrary installation path: the caller selects it.

Findings corrected in review: compare bool fields by type as well as value; reject
non-object plist/receipt values; compare the pre-signing source-binary hash with the
app resource manifest, not the post-signing executable bytes. Codesign checks the
final executable/resources. A pinned bundle-manifest digest is optional and never
claimed to prove publisher authenticity. Check is read-only for selected inputs;
a local concurrent writer is outside the snapshot guarantee for installed files.

Checks: verifier unit suite (4 tests with corruption/metadata/manifest subcases),
installer (7), producer (9), stage dispatcher (2), diff whitespace. All passed.
The actual final build23 packet and installed apps were verified offscreen using
manifest f0d3a2ac660fd68e6bbace8adc19096dc07d056af8753ded8fc9d4da794da6d4:
SDK a9857bff74590959d688e2827af2918d87468b1667773b6a278123f8fb4cbc75,
source e0e011c605f8ee6ea84827414a1636bcca74c97e, five archive and five installed
strict signatures plus all-file/hash/receipt/app-metadata checks passed.
No app launch, physical input, notebook access, reinstallation or new native
qualification is implied. New unit suite joins the existing tooling CI stage.
