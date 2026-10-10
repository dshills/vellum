# Historical baseline Prism review

Prism 0.7.1 reviewed 83 source/patch/script files, 634,744 bytes in 20 chunks. Coverage is complete. It reported three high and 23 medium findings. This commit archives prior experiments and their original evidence; it does not approve them as production code.

The active follow-up replaces the native client and event transport. It addresses protocol conformance, Intel structure-return dispatch, range arithmetic, semantic scalar bounds/geometry/indexed traversal, action parity, and timer ownership. Historic experiment files remain unchanged so their recorded hashes and measurements retain meaning. Predictable temporary build/cache defaults are unsafe on a shared machine; active follow-up runners use private mktemp defaults. The grapheme mutation/performance findings and Gio projection/reporting findings are retained as open historical defects until independently fixed.

- `b05152b85629b8e3` high: Predictable Cargo home permits local code execution (prototypes/foundation/run.sh).
- `5a33f7ea1264c820` high: Use the structure-return dispatcher for NSRect on Intel (prototypes/nativeinput/inputbridge/input_darwin.m).
- `43ff19295bef6d0e` high: Register the input-client protocol on the dynamic subclass (prototypes/nativeinput/inputbridge/input_darwin.m).
- `4b51aed3bf75c1ab` medium: Use a private directory for build caches (prototypes/foundation/env.sh).
- `195fce38bb1ab1bf` medium: Accessibility text editing implicitly commits the value (prototypes/foundation/main.go).
- `1fbfd49c667f30c2` medium: Tree leaves omit their selection state (prototypes/foundation/main.go).
- `710e7b166a2e9809` medium: Bounds validation accepts non-finite coordinates (prototypes/foundation/semantics/tree.go).
- `92e71192187eb30a` medium: Tree traversal repeatedly scans all nodes (prototypes/foundation/semantics/tree.go).
- `406ce773b8504e53` medium: Keep build caches out of predictable shared temporary paths (prototypes/nativeinput/env.sh).
- `c63d704d7f9cd310` medium: Validate offsets before signed range arithmetic (prototypes/nativeinput/inputbridge/input_darwin.m).
- `9a051e12667b9be8` medium: Idle timers access window state outside its lifecycle (prototypes/nativeinput/main.go).
- `def64443cda0e5f7` medium: Boundary lookups repeatedly segment and allocate the entire field (prototypes/nativeinput/patches/0001-grapheme-and-selection.patch).
- `3338eb304215f530` medium: Reject character lengths that split UTF-8 scalars (prototypes/nativeinput/semantics/tree.go).
- `10a3a24fc164190a` medium: Use the existing node map during tree traversal (prototypes/nativeinput/semantics/tree.go).
- `cd84edeffe858901` medium: Handle failures when writing the requested report (prototypes/p0/cogent/main.go).
- `43c87dbb36b60ef1` medium: Use private directories for build and module caches (prototypes/p0/env.sh).
- `f26115f287a376d8` medium: Tree clicks can select the wrong component after projection changes (prototypes/p0/gio/main.go).
- `d56bb738da252542` medium: Report output failures are silently discarded (prototypes/p0/gio/main.go).
- `6d77ee073839b8e6` medium: Move Go caches out of a predictable shared temporary directory (prototypes/remediation/env.sh).
- `a2f02d154aff46ae` medium: Idle timers access the window after teardown (prototypes/remediation/main.go).
- `b657019ec9ddfc9c` medium: Expose Rotate's disabled state to assistive technology (prototypes/remediation/main.go).
- `b3fa2436bebb2c10` medium: Revalidate grapheme boundaries after deletion (prototypes/remediation/patches/0001-grapheme-and-selection.patch).
- `c953b3f18149eab3` medium: Avoid rebuilding every grapheme boundary for each cursor operation (prototypes/remediation/patches/0001-grapheme-and-selection.patch).
- `c2e3a7cd9c9d1942` medium: Reject character lengths that split UTF-8 sequences (prototypes/remediation/semantics/tree.go).
- `ff8d507f99f13d29` medium: Reject non-finite accessibility bounds (prototypes/remediation/semantics/tree.go).
- `de4b039336164ef6` medium: Use indexed node lookup during tree validation (prototypes/remediation/semantics/tree.go).
