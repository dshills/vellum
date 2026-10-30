# One-command private delivery — internal review

Extend the existing producer, rather than introducing a second build pipeline.
--build-version validates one to six decimal digits and is threaded into every
app and the bundle manifest (default 22 retained). --install-to enables the
existing versioned private installer after successful packaging. No app launches,
window focus, notebook IO, settings changes or public publication are added.

Before installation, require the complete five-app set, all six regular bounded
archives, matching checksums/list and a verified SDK with the same clean source.
Delegate app snapshots, ZIP admission, strict signatures and version preservation
to the existing installer. Check the returned installation against the original
manifest identity. Write DELIVERY.json exclusively only after success, binding
source/build/SDK, manifest and installation receipt. Keep packaging completion
(BUNDLE_MANIFEST) distinct from optional installation completion (DELIVERY).
Partial output is retained for diagnosis; no runtime acceptance is recorded.

Internal review corrected a contradictory generated README statement: Notes uses
its established storage; only generated starters avoid defaulting to Notes data.
Reviewed dirty-source refusal, fixed ad-hoc signing route, bounded child commands,
receipt ordering, preservation of prior versions/data, complete app admission,
matching checksums, SDK origin and installation drift refusal. No release audit,
notarization, compiler or backend policy changed. This is a reproducible procedure,
not a promise of identical compiler bytes across hosts.

Validation: 8 producer/delivery tests, 5 installer tests, 35 app packaging tests
and 24 SDK checker tests passed; CLI help and diff checks passed. Failure cases
cover SDK/app corruption, checksum disagreement, incomplete app set, different SDK
source, strict signature failure and manifest drift without a success receipt.
The tests use synthetic ZIPs plus the real installer; SDK checking/signatures are
substituted only in those unit fixtures.

A separate real offscreen reuse exercise copied the unchanged source-4f447e8
build22 SDK and five signed ZIPs to a fresh temporary directory. All archive
hashes and SDK source matched; the actual installer extracted all five apps and
strict codesign verification passed, producing a source-bound DELIVERY receipt.
No app was launched. This proves compatibility with that existing payload, not
newly built binaries containing phase-three fixes. The receipt is retained in
the private mac-offscreen-closeout-20261030 artifact area. A new-source full build
and hosted CI are not claimed by this tooling checkpoint. Desktop checks remain
deferred.
