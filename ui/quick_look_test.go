package ui

import "testing"

func TestPreviewProposalsRejectRemovedAndHiddenItems(t *testing.T) {
	calls := 0
	d := &DocumentList{ID: "files", Label: "Files", Items: []DocumentItem{{ID: "a", Name: "A"}}, Preview: func(DocumentEvent) { calls++ }}
	n, e := d.View()
	if e != nil {
		t.Fatal(e)
	}
	old := pickerNode(n, DocumentListID(d.ID)).DocumentList
	old.Preview(DocumentEvent{ItemID: "a"})
	d.Items = nil
	old.Preview(DocumentEvent{ItemID: "a"})
	if calls != 1 {
		t.Fatal("removed preview", calls)
	}
	o := &OutlineView{ID: "outline", Label: "Outline", Items: []OutlineItem{{ID: "p", Label: "Parent"}, {ID: "a", Parent: "p", Label: "A"}}, ExpandedIDs: []string{"p"}, Preview: func(OutlineEvent) { calls++ }}
	view, e := o.View()
	if e != nil {
		t.Fatal(e)
	}
	view.Outline.Preview(OutlineEvent{ItemID: "a"})
	o.ExpandedIDs = nil
	view.Outline.Preview(OutlineEvent{ItemID: "a"})
	if calls != 2 {
		t.Fatal("hidden preview", calls)
	}
}
