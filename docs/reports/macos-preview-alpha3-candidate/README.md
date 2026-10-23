# Alpha.3 private candidate packaging checkpoint

Candidate source `412e5b8266fb8655935bcdad203b73887923c9db` was clean. This is a
private candidate, not a completed or published developer-preview release. Runtime
features and the approved preview policy remain unchanged. Final native acceptance,
resource observations and distribution-rights decisions remain pending.

The corrected SDK is `v0.1.0-alpha.3`, SHA-256
`ad48d9c9c6112c48db4e4e7d1fb32eb19f31d734891f15761b864d5c2e61b391`.
It has 8,468 verified files and 89 Go graph records. Declared Go 1.25.6 standalone
consumer race/vet/build and public API contract race/vet checks pass after
extraction into a fresh path containing spaces; all replacements remain inside the
SDK. Native starter launch was not repeated during the lifecycle fixture; the R2
alpha.2 native check remains separate scoped evidence. Alpha.3's initial check
failed on a stale alpha.2 contract literal, fixed and reviewed at `412e5b8`.

The matching Placement binary was built from this extracted SDK. Its link map
contains the exact SDK-native static archive path. The native archive SHA-256 is
`786aaad8bd66fa5f7c0dd395c57ff7454838387c8aae29f33f3c6652ecec4883`.
The Placement binary before app signing has SHA-256
`0d7539594ad742844f629ea1ab6080edb0a86ddb8ffbffe8192d3b53d636a16a`;
the signed, packaged executable has SHA-256
`557761aa1368ac312605ee540adb8ada77a8ccd36024a8b81421f8b7783cd7c3`.
The map has non-UTF-8 symbol bytes, so the path check used raw bytes. This check
establishes the observed archive path, not complete component applicability or
absence of every possible linked dependency. The package tool's native dependency
inspection also passed. Numeric bundle version is 0.1.0, build 27, with SDK release
alpha.3 in the app manifest. Only ad-hoc signing was used; no notarization.

The local delivery zip SHA-256 is
`963cde49cc4b9d934ebc9d2ae4a0c14c5d4110e76b8ff3fab393462123b883b5`.
It includes the app, a short README and path-free build identity. After fresh
extraction outside the checkout, strict code-sign verification passes, signed
executable hash matches, and SDK release/binary/notice bindings agree. This closes
the candidate's earlier Documents-location signing concern for these delivered
bytes; it does not fix or qualify packaging in arbitrary metadata environments.
No launch/editing acceptance is claimed yet for this newly extracted alpha.3 app.

Original notices include 416 unique files and 174 conservative component entries.
The 89 effective Go records and native component inventory match the earlier source
notice collection exactly; the bundle was rebound to this binary and validated.
The scope includes Go/native/standard-library and adjacent font declarations; it
is a conservative superset, not a completed linked-content applicability or rights
audit. Both real release and notice audit records remain pending. The public native
route now exists with both guards retained, but cannot create an approved release
from current metadata. Rights triage retains the pinned upstream objc2 caveat and
identifies the concrete unresolved compiled-component family. No external contact,
terms acceptance, public push, tag or release publication occurred.

Retained JSON records normalize user-home paths. Original build logs, archive,
link map, notices and packages remain local. No latency or cadence measurements
were rerun or reclassified.
