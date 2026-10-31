# Item 5: frozen Documents daily-use kit

Added scripts/daily-documents.py and a double-click command wrapper. The launcher pins the qualified build24 candidate, runs the existing local archive/installation verifier, checks minimum macOS and existing candidate processes, and uses a separate persistent evaluation directory. Check-only performs no data creation or launch. Unsafe existing directory permissions and symlink ancestors fail without repair. Existing app repository locking remains responsible for concurrent data access; the process precheck is advisory, not an atomic lock.

The daily-use guide freezes the workflow scope, explains launch and data location, and provides a short session checklist. The issue log begins without invented daily-use observations. Multi-day use remains future user evaluation, not a completed acceptance claim. No runtime or installed app bytes changed.

Validation: two Python unit tests exercise nonmutating check mode directory preparation, private creation, symlink ancestry rejection, and refusal to repair broad existing permissions. The actual check-only launcher passed installed-candidate and host verification without creating data. A subsequent normal launcher run created the separate evaluation workspace and opened Documents; native accessibility showed B with six sample notes, enabled controls, and “Sample notebook · Command-S to keep locally”. It remains open for evaluation. No multi-day use is claimed. Earlier native acceptance records qualify the document workflows on this installed candidate. Default Prism review receipt is retained alongside this report.

The first combined default Prism request failed before producing a receipt: “response cut off even at 16384 output tokens”. Review was retried in separate code/documentation batches without changing provider or model defaults.

The code review completed with two low findings and no medium/high findings. The wrapper now pins its working directory and forwards arguments, including --check-only and --help. Python already resolves all data and verifier paths absolutely. Wrapper --help and zsh syntax checks passed after the adjustment.

The separate documentation review completed with two low findings and no medium/high findings. Initial status is now dated in the issue log, and the prior focus/Undo report is linked. Both successful default review receipts are retained; coverage was complete for each batch.
