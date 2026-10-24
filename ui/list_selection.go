package ui

import "slices"

// ListSelectionEvent proposes a bounded list edit. Apply to the application's
// current accepted selection, so queued Command-clicks compose without relying
// on a last-rendered snapshot. IDs contains one target or an enabled Shift range.
// Toggle flips membership; Add unions; otherwise the edit replaces selection.
// Retaining the current selection rejects the proposal. No history is owned here.
type ListSelectionEvent struct {
	ID, TargetID string
	IDs          []string
	Toggle, Add  bool
}

func (e ListSelectionEvent) Apply(current []string) []string {
	ids := make([]string, 0, len(e.IDs))
	for _, id := range e.IDs {
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	if !e.Toggle && !e.Add {
		return ids
	}
	result := slices.Clone(current)
	for _, id := range ids {
		i := slices.Index(result, id)
		if e.Toggle && i >= 0 {
			result = append(result[:i], result[i+1:]...)
		} else if i < 0 {
			result = append(result, id)
		}
	}
	return result
}

// ListOptionSelected reports accepted membership for single or multiple lists.
func ListOptionSelected(n Node, id string) bool {
	if n.Multiple {
		return slices.Contains(n.SelectedIDs, id)
	}
	return n.SelectedID == id
}
