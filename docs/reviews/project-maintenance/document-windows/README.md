# Route independent document windows through owner-delivered activation

Prism: Anthropic Claude Sonnet 5.5, complete selected diff coverage.

Workspace activation and menu callbacks execute on Application owner, so active/document map state is serialized; documented that invariant. Removed redundant Close defer. AttachFiles is the final fallible initialization step and no error follows its successful scheduler attachment; failed partial registration closes its window and repository. Race, exact-ref activation and menu/history isolation checks pass.

Checks are offscreen/source checks; no new native qualification is claimed.
