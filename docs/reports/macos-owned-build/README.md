# Owned private Mac build checkpoint

The new [build producer](../../../scripts/build-macos.py) compiles the private
workbench and binds its executable and linker map to the same observed successful
Go invocation. The existing collector continues to label supplied pairs
`caller-supplied-unverified`; the producer's separate receipt records
`producer-observed-same-build`. The [contract](../../../specs/release.md#owned-private-mac-build)
defines these distinct scopes.

The producer checks sampled source/SDK/configuration, resolved Go executable and
controlled environment stability. It compares emitted, collected and published
artifact hashes; serialized inventories and streamed compiler logs also have
expected identities. Descriptor-relative output ownership, no-follow file opens,
exclusive filenames and receipt-last publication refuse substitutions and preserve
existing data. Compiler logs have a 4 MiB limit per channel and a 600-second
deadline. Overflow, timeout and sink failures use bounded owned-group cleanup.
Cleanup/diagnostic failures remain visible alongside the original error, with
unconfirmed receipt retirement reported explicitly.

Regression checks cover compiler failures, missing/wrong-architecture output,
input/tool/output drift, metadata substitution and in-place edits, copied-binary
drift, altered receipts, parent redirection, directory replacement and partial
publication. Real child/grandchild fixtures test bounded log cleanup and reject
failed process observations. Checkpoint acceptance requires a real Mac compile
with the actual native map, private packaging and SDK relocation. Exact source
hashes, results and internal/provider review dispositions must be retained in
`records/qualification.json` and the indexed evidence archive. Superseded drafts
and failed fixtures remain diagnostic evidence, excluded from final acceptance.

No produced application is executed. No signing identity, notarization or public
release is requested. The build is not hermetic or an independent attestation;
sampled checks cannot exclude transient changes reverted between observations.
The receipt does not establish complete embedded Rust/C source/component closure,
rights approval, native latency, retained memory or full Mac acceptance. All those
gates remain open. Live desktop tests stay deferred while the human is online.
