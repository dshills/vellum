# Deep-tree presentation origin validation

The paced validator now accepts the existing deep-tree consumer's declared
nonzero starting position: y=68,076 for 10,000 nodes and y=675,576 for 100,000
nodes. It previously required zero for every mode, rejecting valid deep-tree
captures while admitting a translated zero-origin workload. Grid and wide-tree
origins remain exactly zero. The correction changes one initial-Y predicate;
relative deltas, clocks, strict prefix accounting, cutoffs, warmups, sample
populations, latency and cadence thresholds remain unchanged.

All 25 synthetic-journal test methods pass. Four new methods cover both node
counts and requested render modes, wrong origins, configuration crossovers,
and unchanged grid/wide behavior. The original-source counterfactual confirms
four supported deep variants are refused by the old predicate and zero-translated
deep trajectories are wrongly accepted at both sizes. Expected original failures
are retained separately from the passing candidate result.

Both owned command groups terminate and are observed absent. The bound 5,898-file
source vectors, 6,322 prepared dependency files and 8,293 retained SDK files
match before and after. No Go build, native application, OS setting or historical
journal changes occur. This validates the offline checker; it does not qualify
a native capture, performance gate or Mac completion.

The [evidence index](records/index.json) binds the actual logs, command receipts,
source proposal, counterfactual and internal review in the retained archive.
