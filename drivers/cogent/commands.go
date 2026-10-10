package cogent

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/tree"
	"example.com/vellum/app"
	"example.com/vellum/command"
	"fmt"
	"html"
	"image"
	"strings"
	"sync"
)

type commandItem struct {
	ID, Label string
	Disabled  bool
}

// CommandBar renders immutable command snapshots. Cogent also constructs menus
// during widget updates to discover shortcuts, so menu builders must not reenter
// the application owner. Actual activation rechecks the live command registry.
type CommandBar struct {
	Owner    *app.Application
	Registry *command.Registry
	OnError  func(error)
	Toolbar  []string
	mu       sync.RWMutex
	items    []commandItem
	ids      []string
	buttons  []*core.Button
}

func (c *CommandBar) invoke(id string) {
	c.Owner.Do(func() {
		if err := c.Registry.Invoke(c.Owner.Context(), id); err != nil && c.OnError != nil {
			c.OnError(err)
		}
		c.Owner.Invalidate("")
	})
}
func (c *CommandBar) capture() {
	var items []commandItem
	for _, cmd := range c.Registry.Search("") {
		items = append(items, commandItem{ID: cmd.ID, Label: cmd.Label, Disabled: c.Registry.Enabled(cmd.ID) != nil})
	}
	c.mu.Lock()
	c.items = items
	c.mu.Unlock()
}
func (c *CommandBar) snapshot(query string) []commandItem {
	c.mu.RLock()
	defer c.mu.RUnlock()
	q := strings.ToLower(query)
	var out []commandItem
	for _, item := range c.items {
		if strings.Contains(strings.ToLower(item.ID+" "+item.Label), q) {
			out = append(out, item)
		}
	}
	return out
}
func (c *CommandBar) populate(parent tree.Node, query string) {
	for _, item := range c.snapshot(query) {
		id := item.ID
		b := core.NewButton(parent).SetText(html.EscapeString(item.Label))
		b.SetState(item.Disabled, states.Disabled)
		b.OnClick(func(events.Event) { c.invoke(id) })
	}
}
func (c *CommandBar) Mount(parent tree.Node) core.Widget {
	if c == nil || c.Owner == nil || c.Registry == nil {
		return nil
	}
	c.capture()
	c.ids = nil
	c.buttons = nil
	f := core.NewFrame(parent)
	f.Styler(func(s *styles.Style) { s.Direction = styles.Row; s.Grow.X = 1 })
	core.NewButton(f).SetText("Commands").SetMenu(func(m *core.Scene, _ image.Point) { c.populate(m, "") })
	core.NewButton(f).SetText("Command palette").SetMenu(func(m *core.Scene, _ image.Point) {
		query := core.NewTextField(m).SetPlaceholder("Search commands")
		rows := core.NewFrame(m)
		rebuild := func() { rows.DeleteChildren(); c.populate(rows, query.Text()); rows.Update() }
		query.OnInput(func(events.Event) { rebuild() })
		rebuild()
	})
	for _, item := range c.snapshot("") {
		id := item.ID
		if c.Toolbar != nil {
			include := false
			for _, wanted := range c.Toolbar {
				include = include || wanted == id
			}
			if !include {
				continue
			}
		}
		b := core.NewButton(f).SetText(html.EscapeString(item.Label))
		b.OnClick(func(events.Event) { c.invoke(id) })
		c.ids = append(c.ids, id)
		c.buttons = append(c.buttons, b)
	}
	return f
}
func (c *CommandBar) Sync(core.Widget) error {
	if c.Owner == nil || c.Registry == nil {
		return fmt.Errorf("cogent: command owner and registry required")
	}
	c.capture()
	for i, id := range c.ids {
		c.buttons[i].SetState(c.Registry.Enabled(id) != nil, states.Disabled)
		c.buttons[i].Update()
	}
	return nil
}
