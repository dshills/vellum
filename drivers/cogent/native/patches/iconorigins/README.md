# Stage 0099: icon provenance

This stage preserves all Core Go exports and replaces 13 unresolved SVGs through hash-bound asset overlays. Five existing Apache Material derivatives also gain prominent per-file XML modification notices without artwork changes; 18 SVG outputs are overlaid in total. Exact preimage hashes and real parser/raster equivalence controls cover those five prefixes. It adds a compact complete per-file NOTICE.origins ledger, human change notice, full MIT declaration for new artwork only, and actual parser/raster/integrity tests. Existing Core BSD and full Apache declarations are input/output pinned unchanged. See ../../assets/iconorigins/README.md for generation and semantics.

The integrated recipe follows stage98 and checks the contiguous order1..99.
Stage98 modifies independent frame tracing; stage99's icon input identities are
unchanged. This stage never contains the13 replaced SVG bodies. Their previous
hashes appear only as provenance identifiers. Full SDK/application packaging and
remaining M6 checks are separate from this icon source checkpoint.
