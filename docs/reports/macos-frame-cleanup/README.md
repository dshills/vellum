# Mac GPU frame failure cleanup

Owned patch 0048 gives ordinary and measured frame rendering one checked
preparation/submission control path. Encoder/pass/finish refusal releases partial
handles and aborts surface ownership. Failed frames are not presented or read as
images. Successful submission keeps the existing pass → finish → queue → buffer
→ encoder release order. Queue admission does not establish GPU completion.
The checked End and drawer paths return errors; legacy void End logs them.

Surface retirement detaches view/texture references before release, releases the
view before the texture and unlocks once. Repeated idle abort/present does nothing;
Surface Release retires an outstanding frame before normal teardown. Partial
acquisition/view failures release their returned handles. This contract requires
one serialized frame owner and encoding stopped before teardown. It does not add
support for stale/concurrent owners, abandoned active passes, broad panic resource
recovery or device-loss recreation. RenderTexture has no acquired surface lock.

## Verified scope

Fresh preparation passes with 6,183 files: seven existing GPU/drawer files change
and two source/test files are added. Nine decoded outputs match the pinned hashes.
The full GPU/gpudraw offscreen race suites pass three repetitions; the adapter
race suite, GPU/gpudraw vet, native-consumer compilation and all 11 preparation
tests pass. Five final check receipts bind unchanged eight selected repository
leaves and all 6,183 prepared identities. The native binary is compiled only and
retains the pre-existing duplicate -lobjc warning.

Counted-reference fault tables execute the actual shared production functions
for acquisition yield/refusal, nil and partial encoder/pass results, finish
failure, nil/partial command buffers, successful queue admission and Go queue
panic unwind. Actual Surface mutex tests verify abort/teardown order, once-only
release/unlock, repeated idle behavior and release-panic cleanup. Actual ordinary
and measured End entrypoints refuse an unsubmitted frame with one abort and zero
presentations. These inert handles do not inject a physical backend failure.
The image-capture guard is source-reviewed and compiled; real failed-backend
readback is not exercised.

## Version and review evidence

A preliminary overlay compile fails because errors.Log1 initially referred to the
standard errors package. That raw exit remains 1; the import is corrected before
V2 and V3 overlay races pass. V1 lacks execution-time overlay-target binding;
frozen copies remain supporting records only. V2/V3 bind all nine candidate hashes
and the overlay, with unchanged 6,181 prepared originals. Both are historical and
precede the final capture guard and owned preparation. They are not relabeled as
final-source checks. Nine command receipts retain raw exits and output hashes.

Internal review covers all eight final source/spec/pin leaves and nine decoded
outputs. One Prism review using Anthropic Claude Sonnet 5.5 selects six patch,
preparation and documentation leaves, completes coverage and returns exit 0 with
no findings. Two static pin maps are checked internally. The indexed archive
preserves candidate versions, receipts, code-check producer, review input/output
and exact source/prepared maps. Binaries, caches and host desktop artifacts are
excluded. This is selected-source evidence, not a compiler/cache/SDK full closure.

No native window, drawable, input source, clipboard or user preference is touched.
Actual graphics-memory reclamation, physical device faults, sustained latency,
P11/endurance, human input/display and public-distribution gates remain open.
Desktop/performance checks remain deferred while the user uses the computer.
