# Fixed initial viewport

An explicitly sized non-resizable Stage retains its requested client Dp size
instead of restoring an earlier saved size/maximized state. Saved position and
monitor selection remain independent. Resizable windows and legacy fixed
windows without an explicit Dp size retain their existing restore behavior.
Tests exercise both the new policy and both compatibility cases through the
actual creation-options seam. This patch changes no OS preferences.

New test and authored changes are MIT; mainstage.go retains its upstream notice.
Preparation verifies original/output hashes and the patch digest.
