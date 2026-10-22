# Guarded public macOS app producer review

Prism used Anthropic `claude-sonnet-5-5` for all three code/spec reviews. Each
covered all three staged files completely, without exclusions or truncation.
Original provider results and receipts remain local; committed copies normalize
user-home paths. Exit-one results and all findings remain retained, not relabeled.

First review's three medium findings:

- Import-time policy execution: changed to public-only lazy loading, checking
  script/loader availability. Private builds do not load that policy.
- Possible private-mode gate regression: the SDK gate was already a no-op when
  public=false; the call is now public-only. Added a private non-Git copy test
  with policy loading forbidden.
- Authorization drift during output population: added an after-final-validation
  recheck and a regression revoking authorization during the output move. Failure
  removes only the owned partial app and preserves input/notice files.

Second review's two medium findings:

- Local policy code integrity: treated as a documented trust boundary. The local
  producer, interpreter and imported repository helpers are trusted execution
  inputs, just as the existing SDK producer is. This is not a sandbox against a
  maliciously edited packager or compromised host; a self-verifier would supply
  no independent trust anchor. No binary-source attestation is claimed. Owner
  source review and trusted final artifact hashes remain required.
- Initial gate/source snapshot drift: compare source identity across the initial
  gate calls; added a focused commit-change regression that refuses output.

Third review's remaining medium finding asks for an atomic local publication
transaction against a concurrent directory reader. Accepted as a documented
limitation of the existing output-construction contract, not a feature added to
this frozen release. Checks run before reservation and after verification; a
failure retires only owned partial output. Directory existence is not success,
and consumers must wait for successful return then validate artifact hashes.
No upload/public publication occurs here. An atomic rename would still not make
Git/source revocation atomic with the filesystem action. No such stronger
transaction or adversarial-host guarantee is claimed. The final wording-only
clarification was internally checked; no fourth review campaign was warranted.

Focused validation: 33 package tests and nine notice tests pass, including pending
audit, missing/wrong-binary notices, dirty source, index flags, wrong checkout,
initial authorization drift, staging drift and output-move revocation. A stale
alpha.2 test literal now follows recorded release metadata. Real release audit
and notice declarations remain pending; no rights approval, signing identity,
notarization, public push or upload occurs. Framework APIs, runtime features,
rendering/pacing and the preview policy are unchanged.
