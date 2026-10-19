# Original Vellum cursor artwork

These 39 geometric SVG drawings were authored for Vellum without copying the
previous cursor SVG paths. The artwork is copyright 2026 Davin Hills and is
provided under the adjacent MIT LICENSE. Cogent Core's cursor Go implementation
retains its original BSD notice and license.

`python3 -B generate.py --output /path/to/new-directory` reproducibly writes the
SVGs and manifest. It refuses an existing output directory. Hotspots are on a
256-unit square. The manifest records exact hashes and all public cursor names.
The pinned dependency recipe replaces every built-in cursor and updates its
hotspot; no subset of cursor capabilities is omitted. The existing color tokens
and renderer remain in use. Native cursor allocation/cache ownership is a
separate lifecycle qualification; this asset replacement does not change it.
