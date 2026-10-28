package main

import (
	"fmt"

	"github.com/dshills/vellum/ui"
)

type counter struct{ value int }

func (c *counter) view(content ui.Node) ui.Node {
	return ui.Node{ID: "hello", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "heading", Kind: ui.Label, Text: "Vellum · independent Go application"},
		{ID: "count", Kind: ui.Label, Text: fmt.Sprintf("Count: %d", c.value)},
		{ID: "increment", Kind: ui.Button, Text: "Increment", Activate: func() { c.value++ }},
		content,
	}}
}
