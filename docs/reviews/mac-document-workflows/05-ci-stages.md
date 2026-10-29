# Visible CI stages and verified cache reuse — internal review

Split the existing wrapper into ordered, individually runnable stages and named
Actions steps. Preserved all original test/vet/build commands, explicit pinned
selectors, timeouts, SDK consumer job and offscreen benchmarks. An automated
comparison of original/new invocation lines found no removals; only the new
stage dispatcher regression was added. The default wrapper still runs all six.

Reviewed working-directory/env boundaries, failure propagation, invalid argument
refusal and cache trust. Each job has read-only contents permission, no persisted
checkout credentials, a disjoint cache namespace and exact dependency/Go/host
restoration prefix. Only downloads/compiler caches are reused. Native sources
are reconstructed with original checksum and patch validation; root/native
`go mod verify` remains mandatory after any cache restoration. No prepared trees,
notice/audit state, private files or release output enter the cache.

Focused dispatcher tests pass for early listing/refusal, failure-free ordered
stage composition and retained critical commands using inert tools. Shell and
workflow YAML parse checks passed. Existing cached root/native modules verified.
The test's first selector spelling was corrected to match the actual existing
regression name. Full hosted execution is pending a later push; unchanged broad
campaigns were not rerun. No external review service was used.
