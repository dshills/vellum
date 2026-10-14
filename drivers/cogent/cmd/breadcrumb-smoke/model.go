package main

import (
	"fmt"
	"github.com/dshills/vellum/ui"
	"slices"
)

type model struct {
	path     ui.Breadcrumb
	reject   bool
	requests int
	failure  error
}

func fullPath() []ui.BreadcrumbItem {
	return []ui.BreadcrumbItem{{ID: "home", Label: "Home"}, {ID: "folder", Label: "Folder"}, {ID: "document", Label: "<Current>"}}
}
func newModel() *model {
	m := &model{reject: true}
	m.path = ui.Breadcrumb{ID: "path", Label: "Location", Items: fullPath(), Navigate: m.navigate}
	return m
}
func (m *model) navigate(e ui.BreadcrumbEvent) {
	m.requests++
	i := slices.IndexFunc(m.path.Items, func(item ui.BreadcrumbItem) bool { return item.ID == e.ItemID })
	if e.ID != m.path.ID || i < 0 || i == len(m.path.Items)-1 {
		m.failure = fmt.Errorf("unexpected ancestor proposal %+v", e)
		return
	}
	if !m.reject {
		m.path.Items = slices.Clone(m.path.Items[:i+1])
	}
}
func (m *model) view() ui.Node {
	n, err := m.path.View()
	if err != nil {
		m.failure = err
		return ui.Node{ID: "failure", Kind: ui.Label, Text: err.Error()}
	}
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{n, {ID: "context", Kind: ui.Label, Text: "Controlled breadcrumb native acceptance"}}}
}
