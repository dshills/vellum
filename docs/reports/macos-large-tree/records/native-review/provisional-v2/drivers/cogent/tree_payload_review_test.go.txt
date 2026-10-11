//go:build offscreen

package cogent

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"strings"
	"testing"
)

func TestReviewerTreeGeneratedPayloadRefusedBeforeAllocation(t *testing.T) {
	for _, name := range []string{"key", "control-id", "escaped-total"} {
		t.Run(name, func(t *testing.T) {
			count := 1
			if name == "escaped-total" {
				count = 40
			}
			nodes := make([]virtual.TreeNode, count)
			for n := range nodes {
				nodes[n] = virtual.TreeNode{ID: fmt.Sprintf("key-%d", n), Label: "Small label"}
				if name == "key" {
					nodes[n].ID = strings.Repeat("k", 5000)
				}
				if name == "escaped-total" {
					nodes[n].Label = strings.Repeat("<", 64000)
				}
			}
			data, err := virtual.NewTreeData(context.Background(), "review-payload", 1, nodes, virtual.TreeLimits{})
			if err != nil {
				t.Fatal(err)
			}
			index, err := virtual.BuildTreeIndex(context.Background(), data, virtual.TreeIndexOptions{})
			if err != nil {
				t.Fatal(err)
			}
			model, err := virtual.NewOutline(index, virtual.OutlineConfig{RowHeight: 23, MaxRows: 500}, virtual.OutlineViewport{Width: 400, Height: count * 23})
			if err != nil {
				t.Fatal(err)
			}
			id := "outline"
			if name == "control-id" {
				id = strings.Repeat("c", 5000)
			}
			a, err := app.New(app.Config{})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			if err = a.Open(app.Window{ID: "one", Title: "Review", Width: 800, Height: 600, View: func() ui.Node { return ui.Node{ID: id, Kind: ui.DataTree, Tree: model, Grow: true} }}); err != nil {
				t.Fatal(err)
			}
			manual, err := NewManual(testContext, a, manualConfig(t, 7))
			if manual != nil {
				if closeErr := manual.Close(testContext); closeErr != nil {
					t.Fatal(closeErr)
				}
			}
			if !errors.Is(err, ErrTreeBudget) {
				t.Errorf("native payload %s admitted, expected ErrTreeBudget; err=%v", name, err)
			}
			outlineRegistry.Lock()
			remaining := outlineRegistry.models[model]
			outlineRegistry.Unlock()
			if remaining != nil {
				t.Error("refused payload retained binding claim")
			}
		})
	}
}
