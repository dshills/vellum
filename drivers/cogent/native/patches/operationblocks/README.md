# Bounded scene operation staging

Stage 0044 applies after stage 0043 to the pinned Cogent source. Input, patch
and output SHA256 manifests are checked during transactional preparation.

Private 64-record blocks replace construction-time slice growth. Admission
checks accepted count and the unchanged logical operation charge before block
allocation. At most 1024 blocks exist per builder, with up to 63 unused slots in
the last block. Seal flattens once into exact-capacity retained storage and
retires construction state on success and refusal. Refused span streams skip
compaction; the operation charge is checked after spare row headers are refunded.

The encoder reuses tile counts as fill positions after emitting each original
count. Encoded output remains detached and byte-identical to the prior format.
Logical packet bytes exclude block metadata, spare slots, construction overlap
and encoder scratch; no total-memory or frame-latency guarantee follows.

Tests capture hashes from unchanged fdb67bc before this patch, check a separate
tile-outer painter-order oracle, exercise operation/block/payload boundaries,
source mutation and both retirement paths. Existing randomized image/draw and
exact scanner coverage tests remain. Live GPU and performance qualification is
deferred until the user says the computer is free.
