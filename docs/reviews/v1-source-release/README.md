# v1.0.0 tagged source release

The owner requested a tagged release and README update on 26 October 2026 after
merging the three showcase apps. The current source already identifies v1.0.0;
this checkpoint creates that root tag and a GitHub source release. No native
SDK/app assets, native-module tags, signing identities, notarization or audit
approval changes are included. The native public packaging guards remain intact.

README and release notes distinguish source archives from prepared SDKs, pin
checkout instructions to the tag, link the original demos/screenshots and retain
Mac/human/experimental qualification limits. Historical private build9 evidence
keeps its own identities. The initial focused version/automation metadata test
passed; 81 local documentation links resolved. Calling the actual public SDK
release guard with the updated metadata still refused the pending native audit.

Prism / Anthropic Claude Sonnet 5.5 initially identified ambiguous publication
flags. Addressed by explicit nonbreaking publication_scope fields: historical
personal-v0.1 artifacts and native SDK/app assets. The existing false guard flags
are retained. The closeout paragraph now explicitly scopes its original public
tag prohibition and points to the later source-only authorization. No guard is
renamed or weakened. A final review covers this clarification before commit.

Remote workflow results must remain tied to their actual source SHA. No pending
workflow is described as passing and old package verification is not relabeled.

The follow-up Prism review asked for a distinct source authorization key.
Addressed by naming the new field `source_publication_authorized`, leaving both
pre-existing native/historical `public_publication_authorized` flags false.
This is a metadata-name clarification; the SDK guard still reads the pending
third-party audit and refuses public SDK packaging. Both reviews have complete
coverage and their single findings are addressed. Version-test evidence is reused
for the unchanged version/automation fields; final JSON and guard assertions were
checked again after the source-field rename. No runtime source changed.
