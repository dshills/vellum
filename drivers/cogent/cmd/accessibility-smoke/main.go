// accessibility-smoke queries and exercises the actual Cocoa accessibility tree
// from an app worker. It never changes input sources or VoiceOver preferences.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"cogentcore.org/core/core"
	"cogentcore.org/core/events/key"
	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

type rows struct{}

func (rows) Len() int { return 100000 }
func (rows) Row(i int) virtual.Row {
	return virtual.Row{ID: fmt.Sprintf("row-%d", i), Cells: []string{fmt.Sprintf("Row %d", i)}}
}

type model struct {
	checked, dialog     bool
	text, tab, selector string
	number              float32
	presses             int
	table               *cogent.VirtualTable
	tree                *cogent.VirtualTree
}

func (m *model) view() ui.Node {
	return ui.Node{ID: "root", Kind: ui.Column, Grow: true, Children: []ui.Node{
		{ID: "button", Kind: ui.Button, Text: fmt.Sprintf("Press count %d", m.presses), Accessibility: ui.Accessibility{Name: "Increment counter"}, Activate: func() { m.presses++ }},
		{ID: "enabled", Kind: ui.Checkbox, Text: "Enable controls", Checked: m.checked, Change: func(e ui.ChangeEvent) { m.checked = e.Checked }},
		{ID: "disabled-scope", Kind: ui.Column, Disabled: !m.checked, Children: []ui.Node{
			{ID: "text", Kind: ui.TextField, Text: m.text, Accessibility: ui.Accessibility{Name: "Project name"}, Change: func(e ui.ChangeEvent) { m.text = e.Text }},
			{ID: "number", Kind: ui.NumberField, Number: m.number, Numeric: ui.Numeric{Min: 0, Max: 10, Step: 2}, Change: func(e ui.ChangeEvent) { m.number = e.Number }},
		}},
		{ID: "tabs", Kind: ui.Tabs, SelectedID: m.tab, Options: []ui.Option{{ID: "first", Label: "First"}, {ID: "second", Label: "Second"}}, Change: func(e ui.ChangeEvent) { m.tab = e.SelectedID }, Children: []ui.Node{
			{ID: "first", Kind: ui.Column, Children: []ui.Node{{ID: "first-label", Kind: ui.Label, Text: "First page"}}},
			{ID: "second", Kind: ui.Column, Children: []ui.Node{{ID: "second-label", Kind: ui.Label, Text: "Second page"}}},
		}},
		{ID: "overlap-selector", Kind: ui.Selector, SelectedID: m.selector, Options: []ui.Option{{ID: "overlap-first", Label: "Earlier choice"}, {ID: "overlap-last", Label: "Later choice"}}, Change: func(e ui.ChangeEvent) { m.selector = e.SelectedID }},
		{ID: "table", Kind: ui.Extension, Custom: m.table},
		{ID: "outline", Kind: ui.Extension, Custom: m.tree},
		{ID: "show-dialog", Kind: ui.Button, Text: "Preferences", Activate: func() { m.dialog = true }},
		{ID: "dialog", Kind: ui.Dialog, Text: "Preferences", Open: m.dialog, Dismiss: func() { m.dialog = false }, Children: []ui.Node{
			{ID: "dialog-text", Kind: ui.TextField, Text: m.text, Accessibility: ui.Accessibility{Name: "Name in preferences"}, Change: func(e ui.ChangeEvent) { m.text = e.Text }},
			{ID: "dialog-apply", Kind: ui.Button, Text: "Apply", Activate: func() { m.dialog = false }},
		}},
	}}
}
func element(s cogent.AccessibilityInspection, id string) (cogent.AccessibilityElement, bool) {
	for _, x := range s.Elements {
		if x.DomainID == id {
			return x, true
		}
	}
	return cogent.AccessibilityElement{}, false
}
func wait(ctx context.Context, predicate func(cogent.AccessibilityInspection) bool) (cogent.AccessibilityInspection, error) {
	return waitWindow(ctx, "ax", predicate)
}
func waitWindow(ctx context.Context, window string, predicate func(cogent.AccessibilityInspection) bool) (cogent.AccessibilityInspection, error) {
	for {
		s, err := cogent.InspectAccessibility(ctx, window)
		if err == nil && s.Reachable && predicate(s) {
			return s, nil
		}
		select {
		case <-ctx.Done():
			return s, fmt.Errorf("accessibility wait: %w; last=%+v; last query error=%v", ctx.Err(), s, err)
		case <-time.After(20 * time.Millisecond):
		}
	}
}
func action(ctx context.Context, id, kind, value string) error {
	ok, err := cogent.PerformAccessibilityAction(ctx, "ax", id, kind, value)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("native accessibility refused %s %s", id, kind)
	}
	return nil
}
func check(ctx context.Context, a *app.Application, m *model) (result error) {
	stage := "initial native tree and geometry"
	defer func() {
		if result != nil {
			result = fmt.Errorf("%s: %w", stage, result)
		}
	}()
	first, err := wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "text")
		return ok && x.Frame[2] > 0 && x.Frame[3] > 0
	})
	if err != nil {
		return err
	}
	if first.WindowRole != "AXWindow" || !first.NativeWindow {
		return fmt.Errorf("native window identity or role lost: %+v", first)
	}
	textElement, _ := element(first, "text")
	if textElement.HitTestID != textElement.ID {
		return fmt.Errorf("native text hit test disagrees with published geometry: %+v", textElement)
	}
	if len(first.Elements) > 40 {
		return fmt.Errorf("virtual tree unbounded: %d", len(first.Elements))
	}
	button, _ := element(first, "button")
	if button.Role != "AXButton" || button.Name != "Increment counter" {
		return fmt.Errorf("native button semantics: %+v", button)
	}
	if _, ok := element(first, "second-label"); ok {
		return fmt.Errorf("inactive tab published")
	}
	stage = "deterministic equal-frame native hit testing"
	for i := 0; i < 8; i++ {
		snapshot, err := wait(ctx, func(s cogent.AccessibilityInspection) bool {
			parent, parentOK := element(s, "overlap-selector")
			earlier, earlierOK := element(s, "overlap-first")
			later, laterOK := element(s, "overlap-last")
			return parentOK && earlierOK && laterOK && later.ID != "" && later.Frame[2] > 0 && later.Frame[3] > 0 && parent.Frame == earlier.Frame && earlier.Frame == later.Frame && parent.HitTestID == later.ID && earlier.HitTestID == later.ID && later.HitTestID == later.ID
		})
		if err != nil {
			return fmt.Errorf("equal-frame retained sibling probe %d: %w; tree=%+v", i, err, snapshot)
		}
	}
	stage = "counter press"
	if err = action(ctx, "button", "press", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, _ := element(s, "button")
		accepted := false
		if err := a.Do(func() { accepted = m.presses == 1 }); err != nil {
			return false
		}
		return accepted && x.Name == "Increment counter"
	})
	if err != nil {
		return err
	}
	stage = "text value proposal"
	if err = action(ctx, "text", "set", "Vellum accepted"); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, _ := element(s, "text")
		return x.Value == "Vellum accepted"
	})
	if err != nil {
		return err
	}
	stage = "numeric proposal and clamp"
	if err = action(ctx, "number", "set", "100"); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, _ := element(s, "number")
		return x.Value == float64(10)
	})
	if err != nil {
		return err
	}
	stage = "disabled refusal and re-enable"
	if err = action(ctx, "enabled", "press", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool { x, _ := element(s, "text"); return !x.Enabled })
	if err != nil {
		return err
	}
	if ok, err := cogent.PerformAccessibilityAction(ctx, "ax", "text", "set", "blocked"); err != nil || ok {
		return fmt.Errorf("disabled edit admission: %v %v", ok, err)
	}
	if err = action(ctx, "enabled", "press", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, _ := element(s, "text")
		return x.Enabled && x.Value == "Vellum accepted"
	})
	if err != nil {
		return err
	}
	stage = "native text focus and Tab"
	if err = action(ctx, "text", "focus", ""); err != nil {
		return err
	}
	focused, err := wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "text")
		return ok && x.ID != "" && s.FocusedID == x.ID
	})
	if err != nil {
		return err
	}
	native := core.TheApp.WindowByName("Vellum accessibility smoke")
	if native == nil {
		return fmt.Errorf("native window unavailable")
	}
	native.Events().KeyChord('	', key.CodeTab, 0)
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "number")
		return ok && x.ID != "" && s.Generation > focused.Generation && s.FocusedID == x.ID
	})
	if err != nil {
		return err
	}
	stage = "selected tab visibility"
	if err = action(ctx, "tabs", "set", "second"); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		_, visible := element(s, "second-label")
		_, hidden := element(s, "first-label")
		return visible && !hidden
	})
	if err != nil {
		return err
	}
	stage = "virtual rows and paging"
	table, err := wait(ctx, func(s cogent.AccessibilityInspection) bool { _, ok := element(s, "row-0"); return ok })
	if err != nil {
		return err
	}
	row, _ := element(table, "row-0")
	if err = action(ctx, row.ID, "press", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool { x, _ := element(s, "row-0"); return x.Selected })
	if err != nil {
		return err
	}
	tableElement, ok := element(table, "table")
	if !ok || tableElement.ID == "" {
		return fmt.Errorf("table container missing")
	}
	var next string
	for _, x := range table.Elements {
		if x.Parent == tableElement.ID && x.Name == "Next page" {
			next = x.ID
		}
	}
	if next == "" {
		return fmt.Errorf("virtual paging semantics absent")
	}
	if err = action(ctx, next, "press", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		_, current := element(s, "row-4")
		_, old := element(s, "row-0")
		return current && !old
	})
	if err != nil {
		return err
	}
	stage = "native outline disclosure independent of selection"
	outlineState, err := wait(ctx, func(s cogent.AccessibilityInspection) bool { _, ok := element(s, "tree-row-0"); return ok })
	if err != nil {
		return err
	}
	parentRow, _ := element(outlineState, "tree-row-0")
	leafRow, leafPresent := element(outlineState, "tree-row-1")
	if !leafPresent || leafRow.ID == "" {
		return fmt.Errorf("outline leaf absent")
	}
	if admitted, err := cogent.PerformAccessibilityAction(ctx, "ax", leafRow.ID, "expand", ""); err != nil || admitted {
		return fmt.Errorf("leaf disclosure admission: %v %v", admitted, err)
	}
	if err = action(ctx, parentRow.ID, "expand", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "tree-row-0")
		_, child := element(s, "tree-child")
		return ok && x.Expanded && !x.Selected && child
	})
	if err != nil {
		return err
	}
	if err = action(ctx, parentRow.ID, "press", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "tree-row-0")
		_, child := element(s, "tree-child")
		return ok && x.Expanded && x.Selected && child
	})
	if err != nil {
		return err
	}
	if err = action(ctx, parentRow.ID, "collapse", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "tree-row-0")
		_, child := element(s, "tree-child")
		return ok && !x.Expanded && x.Selected && !child
	})
	if err != nil {
		return err
	}
	stage = "native outline pager geometry and controlled paging"
	outlineTree, err := wait(ctx, func(s cogent.AccessibilityInspection) bool {
		outline, ok := element(s, "outline")
		if !ok {
			return false
		}
		for _, x := range s.Elements {
			if x.Parent == outline.ID && x.Name == "Next page" {
				return x.ID != "" && x.Frame[2] > 0 && x.Frame[3] > 0 && x.HitTestID == x.ID
			}
		}
		return false
	})
	if err != nil {
		return err
	}
	outline, _ := element(outlineTree, "outline")
	var outlineNext, outlinePrevious string
	for _, x := range outlineTree.Elements {
		if x.Parent == outline.ID {
			switch x.Name {
			case "Next page":
				outlineNext = x.ID
			case "Previous page":
				outlinePrevious = x.ID
			}
		}
	}
	if outlineNext == "" || outlinePrevious == "" {
		return fmt.Errorf("outline native pager targets missing")
	}
	if err = action(ctx, outlineNext, "press", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "tree-row-2")
		_, old := element(s, "tree-row-0")
		return ok && !old && x.Frame[2] > 0
	})
	if err != nil {
		return err
	}
	if err = action(ctx, outlinePrevious, "press", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool { _, ok := element(s, "tree-row-0"); return ok })
	if err != nil {
		return err
	}
	stage = "modal open and geometry"
	if err = action(ctx, "show-dialog", "press", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "dialog-text")
		_, outside := element(s, "button")
		return ok && !outside && x.Frame[2] > 0
	})
	if err != nil {
		return err
	}
	stage = "modal native focus Tab wrap Escape"
	if err = action(ctx, "dialog-text", "focus", ""); err != nil {
		return err
	}
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "dialog-text")
		return ok && x.ID != "" && s.FocusedID == x.ID
	})
	if err != nil {
		return err
	}
	native.Events().KeyChord('	', key.CodeTab, 0)
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "dialog-apply")
		return ok && x.ID != "" && s.FocusedID == x.ID
	})
	if err != nil {
		return err
	}
	native.Events().KeyChord('	', key.CodeTab, 0)
	_, err = wait(ctx, func(s cogent.AccessibilityInspection) bool {
		x, ok := element(s, "dialog-text")
		return ok && x.ID != "" && s.FocusedID == x.ID
	})
	if err != nil {
		return err
	}
	native.Events().KeyChord(0, key.CodeEscape, 0)
	final, err := wait(ctx, func(s cogent.AccessibilityInspection) bool {
		_, outside := element(s, "button")
		_, modal := element(s, "dialog-text")
		return outside && !modal
	})
	if err != nil {
		return err
	}
	var modelErr error
	if err = a.Do(func() {
		if m.presses != 1 || !m.checked || m.text != "Vellum accepted" || m.number != 10 || m.tab != "second" || m.dialog || m.table.Model.StartRow() != 4 || m.table.Model.Selected != "row-0" || m.tree.Start != 0 || m.tree.Model.Selected != "tree-row-0" {
			modelErr = fmt.Errorf("accepted owner state incorrect: presses=%d enabled=%v text=%q number=%v tab=%s dialog=%v start=%d selected=%s", m.presses, m.checked, m.text, m.number, m.tab, m.dialog, m.table.Model.StartRow(), m.table.Model.Selected)
		}
	}); err != nil {
		return err
	}
	if modelErr != nil {
		return modelErr
	}
	stage = "live workbench bridge"
	if err := checkWorkbench(ctx, a); err != nil {
		return err
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"result": "PASS", "native": "macOS NSAccessibility", "reachable_elements": len(final.Elements), "checks": []string{"reachable Cocoa tree", "native NSWindow identity and AXWindow role", "roles labels values native hit testing", "deterministic equal-frame descendant and sibling hit testing", "controlled native actions", "disabled refusal", "Tab order", "selected tab visibility", "bounded stable virtual rows and paging", "real outline pager geometry and native hit testing", "independent outline selection disclosure and leaf refusal", "modal isolation Tab wrap Escape", "live dock panel namespaces activation focus modal geometry and retirement"}, "voiceover_tested": false, "preferences_changed": false})
	return nil
}
func main() {
	if os.Getenv("VELLUM_AX_EVENT_TRACE") == "1" {
		core.DebugSettings.EventTrace = true
		core.DebugSettings.FocusTrace = true
	}
	a, err := app.New(app.Config{})
	if err != nil {
		panic(err)
	}
	defer a.Close()
	table, err := virtual.NewTable(rows{}, 4)
	if err != nil {
		panic(err)
	}
	treeNodes := make([]virtual.TreeNode, 10)
	for i := range treeNodes {
		treeNodes[i] = virtual.TreeNode{ID: fmt.Sprintf("tree-row-%d", i), Label: fmt.Sprintf("Outline row %d", i)}
	}
	treeNodes = append(treeNodes, virtual.TreeNode{ID: "tree-child", Parent: "tree-row-0", Label: "Outline child"})
	treeModel, err := virtual.NewTree(treeNodes)
	if err != nil {
		panic(err)
	}
	m := &model{tree: &cogent.VirtualTree{Owner: a, Window: "ax", Model: treeModel, Count: 2}, checked: true, text: "Vellum", number: 4, tab: "first", selector: "overlap-first", table: &cogent.VirtualTable{Owner: a, Window: "ax", Model: table}}
	if err = a.Open(app.Window{ID: "ax", Title: "Vellum accessibility smoke", Width: 700, Height: 750, View: m.view}); err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	outcome := make(chan error, 1)
	go func() {
		err := check(ctx, a, m)
		outcome <- err
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		cancel()
	}()
	err = a.Run(ctx, cogent.Driver{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "native driver:", err)
		cancel()
	}
	result := <-outcome
	if result == nil && err != nil {
		result = err
	}
	if result != nil {
		fmt.Fprintln(os.Stderr, result)
		os.Exit(1)
	}
}
