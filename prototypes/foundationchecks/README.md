# Remaining foundation checks

Separate, disposable macOS checks layered on the unchanged nativeinput prototype. See [results](../../docs/reports/FOUNDATION_CHECKS_RESULTS.md). G1 remains open. These are diagnostics and two local remediation patches, not an accepted framework or expanded workbench.

## Reproduce

Use the pinned Go/module environment in `env.sh`; build the sibling `nativeinput` prototype first to obtain its pinned AccessKit dylib. `run.sh` creates an owned Cogent module copy, verifies patch input hashes, applies nativeinput patches 0001–0004 and local patches 0005–0006, vets the new commands and builds them. `test.sh` runs focused foundation/lazy-spelling race tests and the four portable bridge/queue packages; it does not claim the entire upstream suite passes. Expanded package/unchanged-baseline failure logs are retained. It never changes the Go module cache. On macOS arm64:

```sh
./run.sh
./test.sh
build/textcheck -nogui                     # original rune primitive; expected failures
build/textcheck -nogui -grapheme           # fixed movement; newline failures remain
build/queuecheck -nogui -events 100000     # lossless FIFO, unbounded admission
build/fontcheck -nogui -font-cache "$PWD/build/font-cache" -repeat 3
build/fontcheck -nogui -font-cache "$PWD/build/font-cache" -repeat 3 -system-only
build/lifecyclecheck -cycles 1000
build/lifecyclecheck -cycles 5000 -fresh-windows
python3 packagecheck.py
. ./env.sh
python3 windowcheck.py
build/windowcheck -minimized -idle-sample 60s -exit-after 66s -report "$PWD/build/minimized.json"
```

Desktop commands must be permitted to initialize Cocoa/GLFW. `-nogui` is required for diagnostic CLIs: the pinned toolkit's `offscreen` build tag alone does not suppress desktop initialization in a standalone executable. `font-cache` must be an owned writable directory; discovery failure is fatal rather than mistaken for a low-memory success. `system-only` changes fallback policy and is a diagnostic control, not an accepted memory solution.

The lifecycle fixture attaches both adapters to hidden raw GLFW views, executes nine deterministic protocol queries per cycle, verifies the resulting text and ten callbacks, detaches in reverse order, drains Cocoa autorelease pools and optionally recreates the view. `-pooled=false` reproduces the misleading unpooled harness control. No live IME, VoiceOver, GPU lifecycle or supported Cogent before-show hook is established by that test. No macOS preference is changed by these runners.

`windowcheck.py` derives an owned idle fixture from a hash-locked, unchanged nativeinput source. It only minimizes its own window through the toolkit API and records the actual GLFW iconified state. Idle CPU/canvas paints/scheduler wakes are proxies; no compositor, GPU allocation or system wakeup measurement is claimed.

`packagecheck.py` copies the original nativeinput bundle, signs the included library and seals the bundle with an ad-hoc signature, then records strict verification, stapler and Gatekeeper assessment. It does not use a signing identity, submit notarization or publish a release. Desktop assessment may require unrestricted local execution to obtain meaningful LaunchServices results.

## Local patches

- 0005 adds logical grapheme movement to the multiline buffer and uses it for the editor's forward/backward/delete/backspace path. Existing rune movement API is preserved. Pointer, vertical, word, arbitrary selection and visual bidi navigation remain separate checks.
- 0006 finds glyphs by cluster interval regardless of ascending/descending order, including ligature interiors. Correct cluster lookup does not establish visual bidi caret or selection behavior.

Patch input hashes and the Cogent BSD license accompany the patches. Existing prototype sources/evidence are preserved. No upstream submission is made.

## Evidence boundaries

`text-baseline`, `font-default` and `font-system-only` were captured with the four nativeinput patches. `text-patched`, `font-patched` and targeted tests use all six patches. The original `adapter-1000`/`window-1000` are unpooled harness controls; only `*-pooled-*` runs support pooled lifetime conclusions. `diagnostic-font-discovery-failed*` are failed setup artifacts, excluded from coverage/memory conclusions. `diagnostic-minimized-dispatch-failed*` records a discarded harness run with nested main-thread dispatch; it has no idle sample and is excluded. `package-validation` records the original package; `package-development` records the sealed copy. `relocated` is a same-host smoke test with developer path environment variables unset, not a clean-machine test. The manifest records hashes and limitations.
