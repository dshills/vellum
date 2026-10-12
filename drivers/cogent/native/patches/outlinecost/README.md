# Exact bounded outline work costs

Stage 4 adds an opaque cost token to successful stage-3 bounded static retrieval.
It records actual aggregate point/operation deltas, including composite traversal,
phantom points, temporary copies and segment bounds. Charge uses the same checked
budget/refusal latch. Failed retrieval and the zero token never authorize work.
Successful replay costs exactly match fresh retrieval; partial consumption after
terminal refusal need not match. The caller still owns each new segment slice.
Ordinary public outline methods do not share renderer cache storage.

Tests cover every glyph in the real static fixture, admission boundaries, shared
batch exhaustion, unsupported encoded/variable/CFF routes and expanded composites.
Run these tests from the driver module, using its pinned production replacements
and dependency versions. Standalone dependency-module test resolution differs.
