# macOS external file-drop checkpoint

The public `ui.Node.FileDrop` API and Mac adapter receive external file references
on the application owner. Payload validation is atomic, bounded and preserves
names. Current retained callbacks, window lifetime, disabled ancestors, active
tabs/workbench panels, clipped bounds and visible modal hosts gate admission.
Position uses Dp relative to the unclipped outer layout origin. Returning false
lets another eligible ancestor try with its own path slice. Returning true marks
the native event handled with copy semantics and invalidates the window.

The workbench demonstrates replacing file references on Board A with journaled
undo/redo and dirty state. Application chrome displays at most three abbreviated filenames;
the complete references remain in the model. It does not read/import contents or
save a document. No production backend patch was required for this checkpoint.

## Verification

Run from the repository root with `VELLUM_CACHE_DIR=/tmp/vellum-p0-cache` and the
native `drivers/cogent/env.sh` sourced. Commands, exit codes, exact changed-source
hashes, host/tool identity and relevant prepared backend hashes are recorded in
[source-and-check-proof.json](source-and-check-proof.json).

- `go clean -testcache` followed by `./scripts/check.sh` on the final source — exit 0. All 87 Python regressions, portable/native race
  tests and vet, selected backend checks, native build and independent consumers
  pass. The SDK consumer exercises the public file-drop API.
- Native routing and workbench file-reference regressions, ten repetitions under
  the race detector — exit 0. Fractional scrolled origins at 2x Dp scale, clipped
  refusal, private payloads, callback replacement/removal, disabled ancestors,
  closed windows, inactive tabs/panels, and modal host fallback are covered.
  The final full suite also covers native optimistic tab selection before model
  admission; both selected pages must agree before a descendant receives a drop.
- Final affected tests/vet after the bounded chrome change — exit 0.
- `go run -race ./cmd/file-drop-smoke -timeout 20s` from `drivers/cogent` — exit 0.
  Actual GLFW Cocoa `performDragOperation` reads file URLs from a private
  pasteboard. A test-only temporary NSWindow subclass supplies a synthetic
  cursor query during that selector and restores its original class immediately.
  The fixture does not move the user's pointer or create an OS drag session.
  It checks exact Unicode/space/newline paths, child admission/refusal, independent
  parent copies, callback replacement/removal, disabled fallback, two native
  windows, modal host fallback and blocking the former background help target.
  An ordered Escape event and restored AX projection provide the final barrier.
  This does not independently assert that point is outside the entire modal scene.
- `go run -race ./cmd/workbench -smoke -rows 100000 -exit-after 4s` — exit 0.
  Sparse history/dirty state, bounded virtual data and GPU pixel readback pass.

## Failed attempts and acceptance limits

The initial selector fixture timed out because GLFW rereads the real cursor
position, ignoring the fake drag location for routing. A synthetic synchronous
cursor query corrected the fixture without changing production code. The next
fixture reached all intended handlers but failed its final lookup because a modal
changes NSWindow.title. Caching the initial live window number fixed that test.
Both failed logs are retained beside the final passing log. Internal review found
and corrected a controlled-tab ordering gap, with a regression. The final full
suite was rerun after clearing test results and the Cocoa smoke rerun against
that correction. Prism reports and finding dispositions are stored in
../../reviews/macos-file-drop/README.md.

A separate compiled private app ran with `-interactive -timeout 120s` and the
expected disposable `/private/tmp/vellum-physical-drop/Finder space Ω.txt` path.
CUA opened that directory in Finder, exposed the main app target, selected the
file and attempted a drag from the Finder icon to an exposed target coordinate.
No model callback was observed; the app timed out with exit 1. This establishes
an incomplete physical acceptance attempt, not a demonstrated framework failure.
The file remained present. Physical Finder drag qualification remains open.
An earlier workbench smoke omitted its automatic-close flag, verified history and
GPU readback but stayed open; only that test process was stopped and the bounded
command above reran successfully. It is not used as passing shutdown evidence.

The pinned GLFW implementation acknowledges native drops before model admission
and offers generic drag feedback. False callbacks control Vellum propagation,
not Finder refusal feedback. Limits bound delivered payload copies, not prior
GLFW/OS allocations. No input source or VoiceOver preferences changed, general
clipboard content was not used by the receive fixture, and no publication,
pushing, signing or notarization occurred. macOS remains experimental with the
remaining gates listed in [macOS completion](../../../specs/macos-completion.md).
