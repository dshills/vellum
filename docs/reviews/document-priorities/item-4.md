# Item 4: normal restoration and interrupted-session recovery

Native acceptance used the unchanged installed build24-b9212489afc9 on macOS 27.2, exclusively with the disposable fixture recorded in item-4-result.json. No runtime source change was needed.

After normal Quit of the two saved eight-note notebooks, the exact test process was absent. Relaunch restored last-active A first, selected n11, exact title/body and “Opened notebook”. Closing clean A exposed B: eight notes, selected n11, exact saved title/body and “Opened notebook”. Both external origins were retained.

Separately, B received unsaved n12 with the literal title and Unicode body recorded in the result. Its completed recovery checkpoint contained nine notes while its external primary remained byte-identical at eight. An exact executable-plus-data-dir process match identified the sole disposable process (PID 6245); only that process received SIGKILL. Relaunch offered recovery. Restore Recovery showed nine notes, selected n12, exact title/body and “Recovered notebook · Save to keep it”. Save changed the status to “Saved notebook” and wrote nine notes to the original external B file. A’s external bytes were unchanged. A final normal Quit completed.

Validation: the race-enabled offscreen TestWindowSession family and TestFilesRecoveryRestoreDirtyWithoutPersistedUndo cover external origin, recovery precedence, session-local history, invalid metadata, and coordinated failure handling. Native observations establish the visible workflow, not geometry or physical-pointer qualification. This is a controlled interruption test, not a power-loss durability guarantee. Default Prism receipt is retained alongside this report.

Prism reported one low-severity retention concern: the temporary fixture can expire. Final synthetic A/B notebooks are retained in item-4-fixture with hashes recorded in the result; the original recovery checkpoint and UI observations remain historical acceptance evidence, not replayable UI traces.
