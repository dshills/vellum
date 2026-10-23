# Guarded public native SDK tooling review

User approval permits this tooling repair only, with both existing release audit
approval and notices bound to the exact archive retained. Real release metadata
and copied component inventories remain pending. No artifact is published.

Prism used Anthropic `claude-sonnet-5-5`; both reviews covered all five staged
source/spec files, complete, without exclusions or truncation. Original receipts
remain local; committed copies normalize user-home paths.

- First medium finding: archive verification did not read the inventory's semantic
  archive identity. Fixed by reading the bounded, hash-verified inventory and
  comparing archive identity, inventory checksum, notice map and component count.
  A regression rewrites that identity and all corresponding hashes; verification
  refuses it. Existing relocated-consumer checks also validate actual notices.
- Follow-up medium finding: manifest declarations cannot independently authenticate
  a release approval. Accepted as a trust limitation and clarified in the spec,
  rather than adding a signing/approval service to the frozen scope. Producer
  approval comes from the existing checked-in release gate, never from notice
  presence. Consumer checks are consistency checks; a trusted expected archive
  SHA-256 binds them to a particular producer artifact. Neither review nor tooling
  grants rights or changes real audit approval. No unresolved behavior defect was
  identified under that documented contract. The final wording-only clarification
  was internally checked; no third model review was warranted.

Focused checks: native GPU 13 tests; SDK archive verifier 24 tests; SDK exporter
43 tests, all pass. The exporter run used declared Go 1.25.6 and writable session
caches after a sandboxed default cache access failed. The full repository check
already passed earlier in this session; no runtime/framework code changed here.
