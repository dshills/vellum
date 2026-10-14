package cogent

// Retained native subtrees can change scenes without remounting each data view.
// Revoke completion before moving an ancestor: its old scene's clear layout
// fence cannot qualify coordinates in the destination scene. The scoped AX
// cache remains last-readable; only a completed destination frame restores it.
func withdrawDataGeometry(r *record) {
	if r == nil {
		return
	}
	if r.control != nil {
		if g := r.control.grid; g != nil {
			g.layoutPending.Store(true)
			g.layout.valid = false
			g.withdrawMeasurement()
			g.measurementScene.Store(nil)
		}
		if t := r.control.outline; t != nil {
			if t.measurement.Swap(nil) != nil {
				_ = t.owner.Invalidate(t.window)
			}
			t.measurementScene.Store(nil)
		}
	}
	for _, child := range r.children {
		withdrawDataGeometry(child)
	}
	for _, child := range activeWorkbenchRecords(r) {
		withdrawDataGeometry(child)
	}
}
