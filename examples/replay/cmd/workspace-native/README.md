# Full workspace native integration

This diagnostic consumes all 216 operation kinds/order and 27 domain checkpoints
from the unchanged 10k/1M professional workspace records. It uses the public
`cogent.DesktopSession` API, actual originating FIFO, hit testing/capture,
keyboard, command/history, worker delivery, composition and window teardown.

Build from the replay module on Mac with the native-check tag, for example:

```sh
cd examples/replay
CGO_LDFLAGS_ALLOW='^-Wl,-rpath,@executable_path/native$' \
  go build -tags=vellum_native_checks -o /tmp/vellum-workspace-native ./cmd/workspace-native
/tmp/vellum-workspace-native -rows=10000 -timeout=3m
```

Prepare the pinned native dependencies first using the repository setup
instructions. Each invocation is a fresh process; the native driver runs once.
The checked private campaign uses a source overlay over its separately identified
retained SDK/native archive and GOMAXPROCS=2. That result does not qualify the
retained SDK as a refreshed release or production-default scheduling/performance.

The native scheduler and wall time remain autonomous. Extra native composition
barriers precede input/inspection; only authored Worker operations select manual
worker jobs. Explicit target clicks use current native bounds. Untargeted points
use independently authored first-frozen-cell/root-disclosure/tab-header/canvas-
world anchors from authorCase. Unknown or clipped anchors refuse; whole-window
scaling is not a fallback. Only authored Select All operations17/122 adapt the
software Control+A intent to Mac Command+A, check the live binding, and report
the actual chord. No keymap or OS setting changes.

Domain data, objects, history, selected keys, notes, tab order, zoom and horizontal
scroll match immutable goldens. Vertical grid scroll follows an independent
reference built from generated stable rows, stable sort/filter, the authored
wheel, source anchor preservation and actual viewport clamping. The reference
first reproduces every software golden and never reads candidate scroll offsets
to calculate expectations. Native viewport dimensions are independently
validated; camera centers permit the frozen one-logical-pixel tolerance.
Completed canvas objects and accepted placement preview are checked directly.
Manual incarnation integers and software image hashes are not native evidence.
The normal result joins scenario/workers and checks tracked fixture/service/
session retirement. A fatal unjoined scenario exits before unsafe deferred cleanup;
an outer owned-process watchdog must retire such failures.

Exactly one JSON result is written on stdout. The inherited editor may emit its
known Monospace diagnostic on stderr; the campaign accepts only that exact pattern
under a combined64-line limit, rejecting unknown stderr and extra JSON frames.
Its synthetic parser controls are separate from actual native evidence.

Automated native Go routing does not verify physical pointer input, IME,
VoiceOver, display presentation, cross-window palette drag, general performance,
full catalogue, resource endurance or full Mac completion. The observed inactive
untitled TUINS Cocoa window keeps native ownership/deallocation classification
open even when tracked fixture counters are zero and the process exits.

The [recorded checkpoint](../../../../docs/reports/macos-native-workspace/README.md)
passes both complete row-count cases; it lists the exact tested source/library
and scheduling configuration, earlier failures and remaining limits.
