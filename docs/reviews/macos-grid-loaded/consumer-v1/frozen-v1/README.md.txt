# Loaded stored-data grid

This independent module starts with two usable empty DataGrid windows. **Load records** generates an immutable array of signed numeric records and builds its initial projections on a bounded task worker. The large window uses one million stored records by default; the small window uses the first 10,000 records from the same initial backing array. Each record contains twenty actual `int64` values, generated with seed 7. Subsequent edits produce independent immutable domain revisions for the affected window.

From a prepared checkout, run:

```sh
export VELLUM_CACHE_DIR=/tmp/vellum-p0-cache
. drivers/cogent/env.sh
cd examples/loadedgrid
go run .
# A smaller data set or indexed variable heights:
go run . -rows 10000 -variable-rows
```

See [native setup](../../drivers/cogent/native/README.md). The module uses repository-relative replacements; its prepared native dependencies belong to the SDK or checkout, rather than a machine-specific development overlay.

**Sort** changes numeric column 1 order. The controlled filter field applies a case-sensitive Contains filter to column 2. Apply first settles that field through its originating window FIFO, so admitted native text and composition reach the controlled value before the worker query captures it. **Clear query** restores the complete domain. **Cancel work** cancels an admitted query, load, or edit without pretending its worker has already terminated. Each window has two task credits; a credit returns only after both worker termination and owner completion. A refused new task leaves an already accepted query intact.

**Edit active cell** (or F2/Return in the grid) opens the real managed cell field. Column 0 is read-only; other columns accept canonical signed `int64` text. Return or **Save cell** settles that field, then copies the stored array and builds the new projection off-owner. The owner checks the captured domain pointer, query token, edit session, cell version and request before accepting the candidate; it publishes its domain pointer only after `CompleteEdit` succeeds. An edit that stops matching a filter can close successfully and remove its row from the view. **External +1** constructs a new domain revision on a worker; an older draft remains recoverable as a conflict. Cancel discards a draft explicitly. Changing sort/filter preserves stable identities, cursor/viewport policy and the captured selection universe; the selected count describes the captured identities, rather than just currently visible filtered rows.

The immutable full-domain source uses dense stable keys and arithmetic reverse lookup. Worker projections store indexed order/inverse arrays. Domain revision and accepted view revision are separate. Variable heights depend on stable domain row identity and stay attached to that identity after sorting/filtering. Source snapshots and retained results are concurrent-read safe; diagnostic counters are atomic and selection captures have no mutable counter state.

The example admits at most 1,000,000 rows, four sort/filter clauses, 4,096 query bytes, 64 bytes per numeric cell, 4,096 edited-cell version entries and 4,096 materialized visible cells. Builder index and working budgets are 64 MiB and 256 MiB respectively, **excluding borrowed stored-domain payloads**. One million records hold 160,000,000 logical value bytes; an edited large-domain revision makes another full array copy on its worker. Current domains, the shared initial base and bounded in-flight immutable revisions can coexist. These bounds are not a total process/RSS or peak-memory estimate.

Portable tests check physical storage, dense indices, immutable edits, cancellation, refusal, domain/query/draft CAS and dual-terminal task credits. A separately reported Go `HeapAlloc` delta observes one retained stored million-row array; it is not an RSS, native/GPU-memory or latency measurement. Offscreen tests declare the actual controlled fields, buttons and DataGrid, then route primitive keys/pointers through Cogent's real managed scene/hit/focus path with supplied fonts/theme, explicit workers and owner batches 1/7/64. They include queued filter input followed immediately by Apply, out-of-order canceled jobs, window closure, settled async editing, conflict recovery and an edit removing the sole filtered row.

The bounded desktop fixture is:

```sh
go run -race . -smoke
go run -race . -smoke -variable-rows
```

It starts empty, loads the stored million records, checks completed current-stamp/frame geometry, runs worker sort/filter/clear, preserves captured selection, saves through the actual originating cell-field settlement and tests an external conflict. Its commands are owner operations, rather than physical desktop typing/pointer automation. Its terminal check closes domain windows; it does not independently measure private decoder cleanup or compositor teardown. Human IME, VoiceOver, broad accessibility, peak memory and end-to-end professional workload performance remain separate acceptance work. Source-bound qualification and preserved exploratory failures are handed to the integration owner for [the durable report](../../docs/reports/macos-grid-loaded/README.md).
