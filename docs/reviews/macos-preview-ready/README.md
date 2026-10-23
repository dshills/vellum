# Prism: macOS preview final publication packet

Anthropic claude-sonnet-5-5 reviewed all six initial staged textual files
(complete coverage, no skipped/truncated/discarded/suppressed findings). Raw exit
was 1: one medium finding, 05a44939e565239b, asked for script-verifiable SDK
prototype exclusion instead of an unsupported assertion. It is addressed by
verify-sdk-alert-scope.py and the actual output on the pinned SDK archive.

A focused follow-up reviewed the new proof and README (three files, complete
coverage). Raw exit was 1 with two medium findings; do not call this a zero-finding
review. Dispositions after internal review:

- f7b9c62c604100d3: addressed. Missing inventory members now have explicit errors;
  a missing selected module produces the explicit version-mismatch failure.
  Re-running on the actual archive returns exactly the retained output.
- ef185a54dc95ee0f: not applicable to the stated claim. PurePosixPath parts matches
  the exact prototypes directory at any depth. It already rejects every path
  containing prototypes/ as a directory, including prototypes/p0/gio/go.mod.
  A renamed unrelated path is outside this literal directory-exclusion claim.
  The pinned archive hash prevents an arbitrary modified archive being accepted.
  Checking arbitrary Go text for the word gio or asserting every top-level
  directory is not needed to verify this exact artifact's excluded directory.
  No general source-origin or dependency-security guarantee is claimed.

Positive archive inspection and wrong-payload hash rejection pass. Both proposed
download hashes, JSON, local report links, exact release-note copy and identical
non-doc source-tree checks pass. No dependency/notice/runtime change was made,
no alert dismissed, and the historical prototype remains outside preview support.
No unresolved candidate runtime finding results from these documentation reviews.

Raw local receipts are preserved; presentation copies replace the home prefix.
Final publication, merge, tags and downloadable assets remain separately
owner-controlled under the approved R5 plan.
