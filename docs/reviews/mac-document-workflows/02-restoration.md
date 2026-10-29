# Document/window restoration — internal review

Added opt-in app-owned WindowSession persistence and startup restoration for
Notes, generated documents and A/B Documents. The record is separate from
notebook/recovery schemas, bounded to 16 KiB, with strict fields and path checks.
IO and restoration use existing document workers; metadata callbacks coalesce.
Native geometry persistence is reused. A/B creation follows saved activation
order. Forge reuses workspace storage with optional last-study restoration.

Reviewed startup/owner/worker boundaries, malformed/missing preservation, recovery
precedence and destination, independent histories, metadata/quit ordering and
retirement. Review corrected a nil test document file reference and a case where
Quit could finish before a metadata failure. Metadata failure now keeps saved
windows open. Save after recovery retains the original external destination.

Focused race tests passed for last external reopen, unchanged private default,
recovery precedence/origin, malformed/missing records, path bounds, metadata write
failure during Quit, Forge study/camera restore and missing study preservation,
plus existing Quit, file and workspace regressions. The initial Forge fixture
lacked the required window title; fixing that fixture established the intended
behavior. Final native checks cover actual geometry and restart; software tests
do not claim physical/other-host qualification. No external review service.
