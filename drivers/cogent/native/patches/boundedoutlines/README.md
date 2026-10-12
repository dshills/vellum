# Bounded static outline retrieval

Stage 3 follows deferred resource stages 1/2. The additive bounded API charges
composite point growth, traversal and segment construction before allocation.
It preserves normal static TrueType outlines; encoded resources, CFF and variable
fonts refuse this candidate route. Ordinary GlyphDataOutline retains its existing
behavior. The scene compiler checks the complete source and uses one aggregate
outline budget across its captured text. Three owned regressions cover exact
ordinary parity, composite exhaustion/failure latch and encoded-resource refusal.
