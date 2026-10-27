# macOS distribution-packet review

Owner request: prepare a delta audit/publication packet using existing evidence,
keep features frozen, review with Prism / Anthropic Claude Sonnet 5.5 and commit.
No audit approval, public binary output, source-tag movement, identity signing,
notarization or target-app desktop work is included.

The initial Prism pass covers all 15 textual files (181,852 bytes), complete,
and retains one medium finding: duplicate initial/corrected check labels made
final-candidate evidence ambiguous. Addressed by unique phase-qualified check_id,
phase and final_candidate_evidence fields. The array actually has 67 entries;
its count is programmatically asserted. Fifty-four initial steps plus thirteen
font-notice-correction steps are retained, with twelve superseded initial package
checks explicitly marked. Fifty-five checks directly support final candidates.
Historical controller attempts are referenced separately by archive path, rather
than duplicating a second ambiguous structured list. No failed attempt is called
successful. The focused packet validator checks unique IDs and both counts.

Retained private pre-sign inputs additionally include the three original icon
PNGs, with their hashes, so later public production does not need an asset rebuild.
The packet's guarded-production recipe demonstrates exact notice reconstruction
and existing public producer calls only after owner audit approval. Current
producer guards still refuse pending approval; release metadata is unchanged.

The added direct font proof is scoped to actual pre-sign binary hashes and finds
the exact original Regular/Bold TTF sequences in all three demos. Full original
font notice text is included in corrected app bundles. The committed evidence
archive has 454 integrity-checked records and contains no app/SDK binaries.
Private candidates and binary-bearing input archives remain outside Git.

The final Prism pass covers 19 textual files (199,478 bytes), complete, with one
medium metadata-clarity finding: Notes/Widgets omitted generated_font_source
while other apps used a list. Addressed with an explicit empty list for those
two entries and a schema-1 packet marker. Their actual selected graphs, Mac Shell
and the Placement reference contain no gofont packages; each showcase selects
Go Regular/Bold. The packet validator now checks the uniform field and expected
font-source counts. JSON member ordering has no semantic effect. No candidate
bytes, implementation, approval flag or producer changed; no broad test/review
campaign was repeated for this serialization clarification.

Both Prism findings are addressed. The final packet integrity/notice tests are
rerun after these metadata changes. Private binary/archive inputs remain outside
Git; the report commit adds no runtime/framework API code and reuses unchanged
exact-source CI. Publication and audit completion remain owner decisions.
