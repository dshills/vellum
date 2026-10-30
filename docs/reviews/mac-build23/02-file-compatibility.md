# Historical saved-file compatibility — internal review

Added fictional codec-produced fixtures with source-file/generator/output hashes:
Notes and Recovery from published build13 source 64bf9cfbb9cfb57d96e1aecd0adad01239ca3c7a;
Notes, Recovery, WindowSession and Forge study from build22 source
4f447e85aa97dbcfa3b36d2707a679c173791777. Historical application source was
extracted to isolated temporary modules and exercised only through model/codecs,
linked to current compatible framework types. No historical GUI binary or user's
storage was launched/read. Exact generation drivers are retained as testdata text.

Current-reader regressions preserve Unicode/CRLF, IDs/next IDs, selected/pinned
notes, complete recovery contents, explicit dirty recovery restoration and clean
post-load history. Forge preserves bodies, round shaft, multi-selection, camera,
features and subsequent Add/Undo/Redo identities. Session metadata is decoded
without opening its fictional recorded path. Fixtures remain unchanged.
Current encoders round-tripped the historical snapshots and separate historical
build13/build22 Notes and build22 Forge decoders accepted those written outputs.
This establishes sample-level rollback readability, not arbitrary downgrade
support or historical-binary/native acceptance. Schemas and production code are
unchanged. Initial test setup needed its private directory created before seeding
Recovery; that fixture correction required no application change.

Internal review checked authentic encoder generation, detached model installation,
full recovery comparison (not only count), fresh journals, monotonic identities,
the missing-vs-explicit recovery decision, representative geometry and immutable
baseline bytes/provenance. No real notebook/study was migrated or overwritten.

Validation: Go 1.25.6 focused historical Notes race tests and offscreen Forge
race tests passed. Both packages passed vet (Forge with offscreen tag). The three
historical-decoder rollback runs passed under race checking. Fixture digest
checks and git diff --check passed. Generation/replay used temporary directories
and no native application run. Public APIs, formats, backend, feature freeze and
desktop deferral are unchanged. Existing build23 application runtime bytes remain
applicable; only regression fixtures/tests were added in this phase.
