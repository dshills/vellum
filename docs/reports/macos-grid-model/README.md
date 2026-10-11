# Indexed portable grid checkpoint

`virtual.Grid` adds fixed-height two-axis viewport projection over a borrowed,
owner-confined indexed source. Headers and leading columns remain frozen; detached
records include natural rectangles, region clips and opaque cell press identities.
Navigation and revision refresh preserve stable active keys and best-effort
scroll anchors without an owner scan or full-row copy. Existing Table and Tree
APIs remain compatible. See the [contract](../../../specs/grid.md).

The procedural 10,000-row and million-row ×20-column fixtures both expose 400
visible cells and 520 materialized cells, with 52 RowKey, 52 IndexOf and 520 Cell
calls. A far jump uses two RowKey/two IndexOf calls and no cell materialization.
These are callback/geometry counts; they do not measure stored-dataset memory,
native rendering or frame performance.

Independent review exposed short strings retaining large backing allocations.
The corrected model clones accepted identities/definitions/active/anchor keys,
and each snapshot detaches row keys and cell text. A private identity token
prevents an opaque ref from keeping the Grid/source alive. Retired or foreign
refs refuse activation, including when a pooled position displays another row.
Historical passed checks and the subsequent red retention regression remain
separate from the final corrected qualification.

[portable-qualification.json](portable-qualification.json) binds the final seven
draft inputs and five checks: repeated race tests ×20, vet, unchanged legacy
Table/Tree compatibility, independent original retention regressions green ×20,
and equal 10k/million-row workload counts. [qualification.json](qualification.json)
binds 100 integrated root/SDK/Python inputs unchanged before/after five passing
checks: full portable race/vet, external SDK API race/vet, and real archive
consumption after deleting the synthetic source checkout. This qualification
uses the exported Grid API for reorder, selection/anchor identity, stale press
and atomic capacity refusal.

The source and specification review uses Prism with Anthropic Claude Sonnet 5.5;
see [review evidence](../../reviews/macos-grid-model/README.md). Raw proof/source/
log documents are preserved as exact round-trippable UTF-8 in
[evidence.json](evidence.json). Native inputs are outside this portable proof.
Native managed rendering, performance, AX/keyboard workflows, variable heights,
all-row selection, editing and large-tree acceptance remain open. This checkpoint
does not complete P6 or the macOS framework milestone. No OS preferences changed.
