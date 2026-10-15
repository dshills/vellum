# One-pass glyph fingerprint

Owned patch 0068 combines fixed-path grammar/coordinate/bounds validation and
internal FNV candidate indexing in one walk. Fingerprints normalize coordinates
to the first validated point pixel origin, preserving integer translations and
fractional phases. Signed relative words may be negative. The returned minimum
origin, guarded hull and full minimum-normalized path comparison stay unchanged.
No new cache, retained metadata, storage bound, packet format or work exemption
is introduced. Hash values are private bucket indices, not persisted identities.

The original geometry-admission/origin and full-equality oracles remain. The old
minimum-origin hash assertion is deliberately superseded by an independent
first-point callback-walker oracle, 4,096 fractional phases, integer translations,
permissive private grammar, malformed/refused inputs and forced collisions.
Actual software grid encoded/accounting/pixel parity, failure controls, pinned
preparation and SDK checks are required. Local software elapsed times are
diagnostic while the human is online; native frame targets remain unqualified.
