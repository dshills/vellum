# Window lifetime references and alpha.2 SDK

The application previously retained every historical window name to prevent
reuse. It now stores live registrations only. `OpenWindow` returns an opaque,
comparable `WindowRef` for one opening; closed names may be reused, while an old
reference stays retired. Close, scoped delivery, dirty consumption and completion
reservations cannot affect a same-name replacement. Task/document controllers and
managed native/manual adapters retain their originating references.

The SDK is `v0.1.0-alpha.2`. This intentionally changes the alpha.1 string-scoped
APIs. The [migration contract](../../../specs/window-lifetimes.md), tutorial,
independent consumers and release metadata describe the new API. Logical UI,
persisted workspace and replay names remain strings.

Verification covers all 15 root packages and the full adapter offscreen race
suite, root/adapter vet, all 11 independent consumers' race/vet checks, application
builds and 65 packaging/SDK Python regressions. The test-only SDK contract module
has no production executable. A private 7,342-file SDK passed checksum verification,
relocation through paths containing spaces, generated consumer race/vet/build and
public API tests. Native commands were compiled without execution.

Regressions cover 1,000 unique-name and 1,000 same-name retirement cycles,
reentrant policy replacement, stale/foreign/zero references, exact opening
receipts, dirty state, retained callbacks, editor history, virtual-view remounts,
owner admission, workbench callbacks and native/manual source retirement. External
consumer tests keep a retired document's close command and worker manager apart
from a new controller/model. The portable retention negative control and six
adapter controls retain their intended failing assertions; adapter production
bytes are unchanged from that reviewed stage.

Initial checks remain recorded as failed: one unused consumer local, generated
build outputs entering the source inventory, the test-only module's inappropriate
executable output flag, and a two-minute workspace replay suite deadline. Current
checks use stable source and outside-source outputs. The unchanged replay suite
completed with a ten-minute Go suite deadline and an eleven-minute process
watchdog; this is functional evidence and does not certify latency targets. See [receipts and bindings](evidence.json).

No live Cocoa window, native GPU workload or physical input was exercised for
this transition. Fresh native same-name lifecycle, full Go/native/GPU memory,
physical input/accessibility/display behavior, sustained performance and endurance
acceptance remain open. Historical native campaigns retain their own source
identities. This checkpoint does not qualify the full Mac milestone or authorize
public distribution, signing, notarization or a tagged release.
