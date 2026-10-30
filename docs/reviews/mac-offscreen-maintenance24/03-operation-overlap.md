# Checkpoint 3: overlapping document operations and retired settlements

Internal review, 30 October 2026. App-local maintenance, no root stable API,
native adapter, schema, backend or feature addition. Desktop checks deferred.

Three focused tests extend existing coverage rather than replacing the lifecycle
suite. A channel-gated real Save refuses Open/Quit/Close until completion, keeps
the exact captured notebook, and allows a subsequent Quit/Cancel. Serialized
metadata completion queues one external Open, rejects another plus Quit/Close,
retains the dirty decision, cancels without replay on unrelated metadata success,
and retries Save-before-Open preserving the prior notebook. Existing owner Drain
helpers admit completions; correctness does not depend on sleep timing.

Reproduced defect: force-close a dirty window awaiting Open text settlement while
another window keeps the application alive; reuse the old ID; deliver the old
completion. Before correction it created an Open prompt/destination on the retired
model. Initial fixture incorrectly tried to reopen the last closed application's
window; adding a survivor isolated the actual lifetime defect before changing code.

Correction: Model.request refuses an attached retired window both at initial
admission and immediately before the deferred action. It checks the exact retained
WindowRef context, not a reusable string ID, and returns app.ErrWindowClosed.
Duplicate completions remain deduplicated, pending state is cleared, existing
input errors keep precedence, and unattached portable models retain their behavior.

Internal review covered owner confinement, error precedence, Quit epoch interaction,
ID reuse, unchanged live history and persistence. Four late-action subcases cover
Open, Save, Close and Delete; no retired prompt/destination/history mutation,
notebook/recovery writes or replacement-model changes occur. New requests on the
retired attached model are refused too. No additional issue remained in scope.

Checks with Go 1.25.6: focused TestOperationOverlap race run; full Notes race suite
and vet; generator Python tests (3); freshly generated independent document
consumer overlap/Quit/pending-request race tests and notebook vet; diff whitespace.
All pass. No native UI or personal file used. Existing installed build23 remains
source e0e011c and does NOT contain this source correction; no repackage/install
or live verification is claimed for it.
