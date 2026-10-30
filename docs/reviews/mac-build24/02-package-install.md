# Checkpoint 2: matched private build24 packaging and installation

Internal review, 30 October 2026. Package source is exactly green main
b9212489afc921171fc5abbd5275d8919ae930dd. Framework 41091357397, Stable root API
41091357457 and Platform adapters 41091357554 passed their first attempts before
packaging began. Existing manual qualification remains skipped by policy.

Rebuilt pinned native preparation, selected actual Go 1.25.6 on PATH and ran the
existing private producer with build-version 24 and nonsynced versioned install.
Reviewed clean-source identity before/after, bound PREFLIGHT/manifest/delivery,
SDK relocation/origins and generated minimal/document race/vet/build checks,
public API scenarios, all five app builds, extracted signatures and installation
receipts. Producer completed successfully; no fix or change to its policy needed.

Copied only completed SDK/app archives and metadata into durable private delivery
storage. Independently compared six archive digests and metadata bytes with the
producer output. The normalized summary is in reports/mac-build24/packaging.json.
Original manifests/receipts remain unchanged. Build23 remains present; no prior
installation was overwritten and no notebook directory was opened.

Signing remains ad-hoc, unnotarized and private. No identity, audit, release tag,
public asset or native dependency update. Known AppKit deprecation/duplicate-link
warnings did not fail the build. The Forge icon-only executable path ran solely
to export its icon and exits before UI/data setup. No application window, helper,
LaunchServices action, desktop capture or physical input occurred.

This commit records packaging; it does not relabel built source as this later
report commit. Independent read-only delivery verification and static linkage
inspection are the next checkpoints. No physical/native acceptance is claimed.
