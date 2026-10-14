# Bounded span construction scratch

Owned patch 0062 reuses temporary row-span arrays inside one scene packet builder.
Flush copies every completed operation into exact detached backing before pooling
its construction rows. A row may borrow only after the unchanged logical
header/span charge succeeds. Row band growth, spare-header refunds, operation
charges, span order and packet encodings remain unchanged.

A builder retains at most 32 buffers, 64 KiB of backing, and 256 spans per buffer.
Full pools drop buffers and use the normal allocation path; they never refuse a
packet. The fixed slot headers are separate transient metadata. Successful and
refused Seal retire every buffer. There is no cross-packet or global pool.

The pinned prior builder checks state and bytes after each span and flush, both
encodings and seeded replay across repeated streams. Reuse, held operations,
slot/backing/cap limits, warm admission boundaries and spare-header refunds have
separate tests. New tests are MIT. Preparation checks all input/output hashes
and the patch SHA256. See [the report](../../../../../docs/reports/macos-span-scratch/README.md).
