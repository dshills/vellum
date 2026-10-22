# SDK alpha.3 contract correction

The first alpha.3 private SDK external-consumer check failed because the exported
SDK contract example still expected alpha.2. Corrected that one release literal.
Focused declared-Go-1.25.6 module race tests and vet pass. Prism Anthropic
`claude-sonnet-5-5` reviewed the changed file with complete coverage and no findings,
discarded findings, exclusions or truncation. Committed review copies normalize
user-home paths. Fresh alpha.3 SDK verification follows this commit; the earlier
failed export/check remains preserved. No runtime behavior or release gate changed.
