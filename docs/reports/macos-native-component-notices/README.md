# Mac native component notice packaging

The private SDK producer now binds a checked-in component-notice inventory to the
exact source-built GPU archive and verifies it after extraction. This closes a
notice packaging omission; the distribution audit and final Mac SDK remain open.

## Implemented behavior

The [notice guide](../../../release/native-components/README.md) and
[release contract](../../../specs/release.md#source-built-gpu-component-notices)
describe the source bundle and validation. All 80 actually compiled Cargo
packages/103 units are represented, along with pinned webgpu headers and the
executed Rust 1.99 toolchain's bundled standard-library copyright/terms. The
82 component entries conservatively include host/build dependencies. They do
not identify every crate surviving the final link or select/approve licenses.

The 408 notice/declaration files preserve their original bytes. Of those, 383
copied crate files match exact inputs recorded by the executed source build;
missing package notice declarations use their precise VCS repository commits.
The CC0 standard-text supplement is explicitly identified. Rust's bundled
copyright document and license directory retain their own stated scope.
No invented copyright holders or grants are inserted.

Packaging requires the bundle for the known source-built archive hash. Unknown
private candidates gain no component coverage claim from it. The existing four
GPU capsule members and module overlays remain unchanged. The SDK manifest and
THIRD_PARTY_INVENTORY reference the inventory, hashes and component count. The
extracted consumer revalidates schema, selected archive identity, references,
counts, file/hash coverage, symlink/path rules and size limits. Notice reads
admit at most the remaining aggregate budget before loading bytes. Older private
capsules without this optional record retain their prior pending-audit behavior.

## Verification and review

77 affected tests pass across native_gpu_test, release-sdk_test, check-sdk_test
and distribution-inventory_test. Tests package and relocate an actual fixture SDK,
accept a notice larger than the old 1 MiB provenance limit and refuse notice
omission/tamper, traversal/symlink paths, known-archive missing/mismatched inventory,
hash-consistent wrong archive/count metadata and exhausted aggregate limits.
The exhausted-budget case verifies that no notice is read. The first test command
used the wrong working directory and failed import; the corrected/final suite
outputs are the verification evidence. Existing prerequisite and intermediate
76-test results are retained separately from the final 77-test result.

The actual 37,401,760-byte source-built archive also passes the existing private
candidate validator and new notice binding: 541 ARM64 object members, 82 components,
409 inventory/notice members. Existing producer source enumeration includes all
409 members. Their staged Git blobs equal original worktree bytes; attributes
preserve upstream CRLF and spacing. No new full SDK or binary copy was produced,
and this checkpoint executes no native application or changes OS preferences.

Internal review found and closed the extracted-consumer semantic-validation gap.
One complete Prism Anthropic Claude Sonnet 5.5 call covered six owned source/test/
contract files, 22,750 bytes, one chunk. Third-party text and generated inventory
were checked mechanically and internally rather than sent as a large provider
payload. Both raw retained medium findings/CLI exit 1 are preserved: required
known-archive bundle omission and pre-read aggregate admission. Both were fixed,
regression tested and approved by final internal review. There was no external
re-review after fixes and no second provider call.

The [record index](records.json) binds the [small evidence archive](records.tar.gz).
Verbatim notice contents are kept once under release/native-components, rather
than duplicated here.

## Remaining gates

Apple-derived binding declarations retain the upstream rights question. Exact
runtime/subcomponent and generated-code attribution, Go/native/font/asset rights
closure and the owner distribution decision remain pending. The existing SDK
must be refreshed and accepted after runtime stabilization. Frame performance,
current endurance and the remaining native/human/display workflows are not
qualified by this packaging checkpoint. The Mac milestone is not complete.
