# Bounded shared glyph-mask cache

Patch 0046 bounds strong face keys to 32 and retained mask pixel data to 8 MiB.
Whole face groups use FIFO admission order; hits do not refresh that order.
Eviction clears the backing-slice prefix and deletes all masks in that group;
current-face self eviction relooks up its map before storing the new mask.
Caller-held masks keep their pixels and lifetime through ordinary Go references.
Cacheable dimensions, key quantization, placement and rendering are unchanged.

The three test files cover original-compatible retention failures, exact byte
boundaries, self eviction, FIFO/hits, backing-prefix clearing, concurrent use,
caller ownership and 64 embedded real-font cold/warm/evicted pixel/placement
comparisons against original cache methods. The reference is generated from the
pinned pre-0046 file with type renaming and duplicate/global declarations removed;
its source hash and transformation are in glyphcache-original-oracle.json.
Original Cogent BSD-style notices remain; new Vellum tests carry MIT notices.

These bounds exclude parsed font size, map/allocator overhead, other caches,
temporary rendering and caller-held masks. The earlier native overlay experiment
reduced Go retention but still showed major IOSurface/physical-footprint growth.
This patch is not full P11 or native/GPU memory acceptance; live retesting is
deferred after the human quiet-host window. No module-cache mutation is used.
