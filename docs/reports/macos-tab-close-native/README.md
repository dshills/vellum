# Native document-tab acceptance

The bounded `tab-close-smoke` command targets actual Cocoa action selectors for
managed document closing. It creates one disposable window with geometry
persistence disabled. Owner FIFO acknowledgements distinguish proposal admission
from accepted callbacks and native publication. Each stage has a five-second
deadline; the process has a sixty-second deadline and checks final native cleanup.

The required sequence edits a draft natively, rejects selected and nonselected
closes while retaining draft identity, accepts nonselected removal, refuses the
now-absent target and a disabled target, accepts final removal, reopens documents
with new native object identities, then accepts selected and final removal.
It checks exact callback counts and rejects unintended tab selection. Object
identity checks do not claim that an old native object was invoked after remount;
the public selector API resolves IDs against the current snapshot.

Run `go build -tags=vellum_native_checks -o <new-path> ./cmd/tab-close-smoke`
from the prepared Mac driver module, then launch that executable in an authorized
quiet-machine test window. Other builds explicitly refuse native qualification.

The eight-stage sequence passes in a real Cocoa window on the current `4c0df93`
framework base, including progress controls and checked RGBA row uploads, with
6,214 exact prepared dependency leaves. The model race test,
vet and native build pass. Six close proposals, one native draft edit, one native
checkbox decision change and zero unintended selection callbacks are observed.
Actual main-thread native cleanup reports stopped callbacks and zero pending,
native, accessibility and window resources.

The same executable fixture source also passes on the earlier stage-58 base with
6,209 prepared leaves. Two faulty native builds on that earlier base fail at
their intended stages: omitted checkbox owner change cannot enable acceptance;
an injected tab-selection callback is detected during rejected closing. Both
faults clean up native resources. These fault controls were not rerun against the
current framework. V1/V2 source-review repairs and original evidence are retained.

Earlier positive runs on `68c3750` and `3a8bd26` retain their own source bindings.
The final source/evidence review and one Prism Anthropic Claude Sonnet 5.5 review
precede commit. This qualifies bounded native action-selector tab acceptance,
with separate current-renderer positive and earlier-base fault evidence. Physical
input, human VoiceOver, performance, retention and the full Mac milestone remain
open.
