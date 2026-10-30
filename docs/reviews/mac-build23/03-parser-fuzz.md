# Bounded codec fuzzing — internal review

Added FuzzDecodeNotebook, FuzzDecodeRecovery, FuzzDecodeWindowSession,
FuzzDecodeStudy and FuzzDecodeWorkspace. Historical valid fixtures, clean recovery,
empty collections, invalid versions/encoding, duplicate-key examples, numeric
extremes and nested JSON seed the corpus. Each mutation input is capped at 64 KiB;
production byte/schema limits are unchanged. Notebook/study accepted values must
validate, encode/decode without data drift and install on fresh clean model
history. Metadata/recovery codecs must have stable reload/encoding behavior.
No native window, file chooser, recovery worker or app entry point is started.

Internal review keeps absent/empty optional collections semantically equivalent
where codecs omit them, avoids expecting rejected partial JSON structures to be
installable, and compares all preserved notebook/study state after valid install.
Separate size regressions pad otherwise valid documents beyond the actual
16 MiB notebook/recovery, 1 MiB study and 16 KiB metadata limits. This avoids
mistaking the smaller fuzz mutation cap for full-size boundary coverage.

Validation: all five targets passed bounded Go 1.25.6 race-enabled runs with
GOMAXPROCS=2, -parallel=2 and -fuzztime=8s. Reported executions: notebook 1457,
recovery 2841, window session 43389, study 1244, workspace 5365. Ordinary seed
runs and actual-byte-limit tests passed, as did focused Notes and offscreen Forge
vet. No crash, unstable round trip or unintended accepted-model/history change
was found, so no production/parser format fix was required. Fuzz discovery is
bounded evidence, not exhaustive qualification or a responsiveness benchmark.

Reproduce from the module with go test -race [Forge: -tags offscreen] PACKAGE
-run '^$' -fuzz '^TARGET$' -fuzztime=8s -parallel=2, using Go 1.25.6 and the
prepared driver environment. Normal CI runs the committed seed corpus as tests;
additional mutation campaigns remain explicit developer actions. No features,
stable API baseline, native dependency pins or release gates changed.
