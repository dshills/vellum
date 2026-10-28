# Record native dependency ownership and guard patch inventory drift

Prism: Anthropic Claude Sonnet 5.5, complete selected diff coverage.

Addressed patch-header counting by delegating full unified patch parsing to read-only git apply --numstat -z. Deleted targets, prefixes and hunk contents are handled by Git. Exact patch hashes and manifest identity remain checked.

Checks are offscreen/source checks; no new native qualification is claimed.
