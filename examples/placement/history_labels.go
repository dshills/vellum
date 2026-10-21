package main

func (m *model) historyCaption(action string) string {
	s := m.journal.Snapshot()
	label := s.UndoLabel
	if action == "Redo" {
		label = s.RedoLabel
	}
	if label == "" {
		return action
	}
	return action + " " + label
}
