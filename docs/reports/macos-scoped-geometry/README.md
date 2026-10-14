# Managed outline ordinary geometry projection

Managed outlines already publish host/row/page frames atomically through the
scoped table/tree transaction, but the ordinary bounds traversal still included
them. Cocoa ignores those scoped IDs in that channel, and their presence causes
the ordinary geometry cache to bypass reuse. The projection now omits bound
managed outline subtrees along with bound managed grids. Full inspection,
ordinary controls, legacy virtual trees and workbench namespaces retain geometry.

The independent ordinary projection oracle subtracts actual scoped semantic IDs
from complete inspection bounds. Four DPI fixtures cover mixed table/outline
content, ready/scrolled/pending/disabled states and workbench namespaces. Legacy
virtual-tree paging compares full and ordinary projections. The earlier grid
editing/private text-geometry fixture remains included. Exact-source checks,
intentional old-behavior fault, Cocoa workflows and reviews are recorded before
commit in the checkpoint qualification and evidence archive.

A separate frozen `8c1709e`/6,216 diagnostic records 644 ordinary grid geometry
decisions per CPU/GPU process: 641 reuses and three sender-wrapper entries. Each
uses 512 completed native samples after 128 warm callbacks. The 59-byte payload
contains two ordinary bounds; the first two decisions precede recorded semantic
membership and bypass reuse. The diagnostic preserves exact overlay, compiled
binary, controller/raw logs/profiles and independently reviewed evidence. Sender
entries precede native invocation and are not application acknowledgements.
Its profiling and diagnostic writes exclude ordinary latency qualification.

The projection fix does not establish a managed-outline frame-time improvement.
The independent grid diagnostic shows its existing reuse works in that fixture;
it does not explain mixed earlier ordinary timings or qualify physical GPU work,
native/GPU memory, presentation, final-source endurance, human accessibility or
full Mac/P6/P11 completion. No performance threshold is changed.
