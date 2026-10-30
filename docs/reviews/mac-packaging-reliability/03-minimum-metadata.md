# Checkpoint 3: bind minimum OS across executable, plist and resource manifest

Internal review, 30 October 2026. Packaging metadata correction only; no application,
framework API, deployment target, backend, support or release policy change.

A shared static inspector reads each Mach-O slice, requires unique macOS deployment
metadata, and records the highest slice minimum for the whole bundle. New packages
write it to LSMinimumSystemVersion and APP_MANIFEST's per-slice deployment record.
Verification before/after signing/publication and independent delivery/linkage
checks compare those fields with actual executable observations. Legacy packages
with neither field remain explicitly unrecorded; partial/contradictory fields fail.

Review findings corrected: Apple's true binary includes arm64e.x1, so architecture
admission accepts bounded dotted subtype names; icon plist-tamper regression now
changes only the icon field rather than deleting unrelated minimum metadata.
Universal floor selection preserves every included requirement, never clamps it.
Public audit/source/notices authorization checks remain untouched.

Checks passed: deployment tests (4), package-app (36), icons (7), producer (12),
independent verifier (4), private linkage (6), dispatcher (2), whitespace. Native
packaging tools also produced and strictly verified a disposable ad-hoc signed
Metadata Fixture copy of an existing Forge binary. Plist/manifest/executable all
report 27.0. The copy was not installed or executed; original build24 was read-only.
No desktop, notebook, signing identity, notarization or public asset change.

Existing build24 remains installed with its original hashes/metadata. New behavior
is for future packages; a full five-app rebuild is unnecessary to validate this
correction and is not performed here. Temporary fixture/evidence is private.
