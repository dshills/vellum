# Repository layout review — 10 October 2026

Documentation now lives under docs/, specifications under specs/, and the macOS
fixture packaging tool under scripts/. Public packages and native executable
locations retain their existing Go import paths and module boundaries. Existing
prototype sources, dependency preparation paths and capture-time JSON evidence
remain intact.

Prism used Anthropic claude-sonnet-5-5 with a medium finding threshold. The initial
review covered the relocated documentation, framework contract, packaging tool
and manifest-generator changes. Its two findings were resolved: only scratch
reviews are ignored, and the Architecture C caption now describes the option
without implying an architecture selection. The final review covered those fixes,
the final report-path corrections and prototype README links. Both reports declare
complete coverage of their selected inputs, with no truncation or skipped chunks.
The final review has no medium or high findings.

The full scripts/check.sh workflow passed: portable race tests and vet, native
module tests and vet, and workbench executable build. The linker emitted the
existing duplicate -lobjc warning. The layout check resolved 254 local Markdown
destinations across the reorganized documents and indexes, parsed changed Python
tooling, checked its document inputs, and verified that relocated non-Markdown
files retained their exact bytes. All 114 existing Go sources and module files
also retained their exact bytes. Native GUI checks were not repeated for this
file-layout change.

prism-initial.json and prism-final.json record the reviews; checks.txt and
verification.json record local validation. The source changes remain limited to
manifest-generator document paths; recorded evidence was not regenerated.
