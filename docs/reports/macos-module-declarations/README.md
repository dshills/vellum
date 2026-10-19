# Missing declarations in pinned WebGPU linker modules

The five Go graph notice gaps are confirmed in the exact pinned upstream
repositories. Each repository root has only `README.md` and its platform module.
The module-root listings have no LICENSE/COPYING/NOTICE file, the repository-root
LICENSE endpoint returns404, and the README only identifies prebuilt WGPU
libraries. The fetched go.mod and Go linker shim match the selected module-cache
bytes; versions, sums, full source commits and observation hashes are retained.
This distinguishes an absent upstream declaration from a collector omission.

| Selected module suffix | Full repository commit |
| --- | --- |
| libs-android | 6b27e30a172e428c2e8aeedb4f69fed102d7a1ee |
| libs-darwin | 66a5dfa57f8dd22992aa1b90bbb67ffa9ec2b46c |
| libs-ios | fe2537e7ddac1d0dfb6f2069750eca43074e8650 |
| libs-linux | 421b8a341d0881131d657a01477d1499e9b8f1f1 |
| libs-windows | f47d1b682eb818318a2e590970dfe2ae4bcf248b |

All names above are under `github.com/oliverbestmann/webgpu/`. The
[evidence](evidence.json) links every exact primary-source URL, status and hash.
For example, the [Darwin root tree](https://api.github.com/repos/oliverbestmann/webgpu/contents?ref=66a5dfa57f8dd22992aa1b90bbb67ffa9ec2b46c)
and [module tree](https://api.github.com/repos/oliverbestmann/webgpu/contents/libs-darwin?ref=66a5dfa57f8dd22992aa1b90bbb67ffa9ec2b46c)
show the inspected scope. This checks root/module-root text declarations, not
all embedded binary content or exact final-link membership. Cached origin/sum
metadata and byte equality are observations, not independent source attestation.

The first batch inspection assumed every shim was named `libs.go`. Android uses
`lib.go`, so its exact-byte assertion failed. Four completed records and Android's
raw root/module/mod/README observations were retained. The missing Android source
and explicit404 response were fetched separately; the final five-record audit
checks every retained hash and cached module sum. Original programs and partial
observations remain in the archive; the first attempt is not called successful.

## Effect on completion

No license terms are inferred from a newer parent module or another platform.
The current source-built native archive's82-component bundle also cannot supply
these Go modules' missing declarations. An applicable upstream declaration or a
deliberately reviewed dependency replacement is needed before this distribution
gap can close. The existing five inventory records remain review-required and
the public release gate remains pending.

This documentation-only checkpoint builds and launches no application, edits no
module cache, changes no framework/runtime/prepared source, rebuilds no SDK and
closes no Mac milestone. It does not contact an upstream maintainer or approve
rights. The [original-byte archive](records.tar.gz) and [member manifest](records.json)
retain the bounded source observations; verification rejects changed hashes,
missing members or unexpected declarations. Internal review covers the exact
checkpoint. No additional provider call is used for this source-inventory result;
Prism remains required for the next significant implementation checkpoint.
