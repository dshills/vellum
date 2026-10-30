# Historical fictional document fixtures

These files were emitted by the historical application encoders identified in
provenance.json, with original source-file and output hashes. The isolated
historical package links against current compatible framework types; it is not
a launch or historical-binary qualification. Fixture text, identities and paths
are fictional. Never open session paths during parser tests.

The tests verify version-one reads, clean history installation, next identities,
Unicode/CRLF, recovery and current-encoder round trips. Separate recorded
historical-decoder checks establish rollback readability for these samples.
No schema migration or general promise of downgrade compatibility is added.

The generate-*.go.txt files retain the exact model/encoder test drivers. To
regenerate, extract non-test application Go files from each recorded commit,
copy a matching generator to compat_fixture_test.go, supply Go 1.25.6 and the
prepared current framework replacements, set VELLUM_COMPAT_OUTPUT to a private
temporary directory and run only TestHistoricalFixture. VELLUM_COMPAT_READ
selects a current-encoder output for the historical-decoder rollback check.
Keep regenerated files separate until their provenance and byte changes are
reviewed. Do not execute the native application entry point.
