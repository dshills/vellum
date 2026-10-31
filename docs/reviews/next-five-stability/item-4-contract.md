# Item 4 preparatory correction: SDK consumer version identity

The first build25 attempt exported and relocated its SDK but failed the external
consumer scenario because that test still required the literal v1.0.0. Its output
at /private/tmp/vellum-stability-build25 is diagnostic, not a completed delivery;
no installation receipt was published from that attempt.

The scenario now compares vellum.Version with the selected checkout/SDK release
metadata and independently requires the stable-v1 major line. Both layouts place
the package at examples/sdk-contract beneath that metadata root. This preserves
version consistency without pinning every compatible future SDK to v1.0.0.

The full consumer race suite passed with temporary Unix-socket access, as did vet.
A focused identity/control regression passed after diagnostic refinements.
Default Prism completed with full coverage and two low diagnostic findings; the
layout requirement is now explicit and version-mismatch/major-line errors are
separate. No public API, runtime behavior, native dependency or old release identity
changed. Candidate construction is retried in a new output directory after commit.
