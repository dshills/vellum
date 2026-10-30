# Producer toolchain and preparation preflight — internal review

Before reserving output or exporting/building, the private producer now checks
an exact declared Go version, darwin/arm64/cgo configuration, resolved Go driver
and compile/link identities, no active prepare lock, expected prepared module
names and five nonempty regular sentinel files. It refuses existing or
in-checkout output and known synced producer/install locations after resolving
aliases. This directly addresses the phase-one Documents signature failure.
It does not strip attributes, alter security settings, select a signing identity,
update dependency pins or prepare dependencies automatically.

PREFLIGHT.json captures source and sampled compiler/native identities. Go's
resolved directory is prepended only to child PATH, avoiding later ambient alias
changes. Tool/sentinel/manifest hashes and the owned preflight receipt are
rechecked before packaging completion. The bundle/delivery manifests bind that
receipt hash; install also rejects receipt tampering. Older complete bundles
without the optional field remain installable. Preflight is a diagnostic and
sampled record, not a replacement for full pinned preparation, module integrity,
relocation, signature checking or authenticated/compiler closure qualification.

Internal review checked ordering before output creation, exact compiler rather
than only go.mod metadata, no downloads/global modifications in the probe, native
preparation concurrency refusal, alias-resolved synced paths, preserved existing
outputs, backward-compatible receipt admission and failure-before-success-record
behavior. The fixture/native hash set is explicitly sampled; no full prepared
source integrity is inferred from it. No production app/runtime, stable Go API
or file format changed.

Validation: 5 preflight tests (table cases for compiler/target/cgo, missing or
active preparation, wrong module, drift, synced alias, source containment and
existing output), 9 producer/delivery tests, 5 installer and 2 stage-dispatch
tests passed. Real positive admission selected Go 1.25.6 and all three binaries,
checked five native sentinel hashes and created no output. The ambient Go 1.27.3
was refused with an actionable version error and no output. CLI help/diff checks
passed. The new test suite joins the existing CI tooling stage. The committed
producer will be exercised for the final build23 payload, retaining prior
versions and opening no app windows. Desktop checks remain deferred.
