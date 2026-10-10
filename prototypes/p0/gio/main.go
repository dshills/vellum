package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"strings"
	"time"

	"example.com/vellum/p0/model"
	"gioui.org/app"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type rowState struct {
	button widget.Clickable
	seen   int
}
type workspace struct {
	s                         *model.Session
	w                         *app.Window
	theme                     *material.Theme
	ops                       op.Ops
	table, tree, paletteList  widget.List
	filter, value, query      widget.Editor
	tabs                      [2]widget.Clickable
	buttons                   [6]widget.Clickable
	paletteButton             widget.Clickable
	paletteCommands           [6]widget.Clickable
	rows                      map[int]*rowState
	treeButtons               [64]widget.Clickable
	palette                   bool
	treeExpanded              bool
	treeToggle                widget.Clickable
	frame, maxRows, maxStates int
	currentRows               int
	started                   time.Time
	report                    string
	exitAt                    time.Time
}

func main() {
	count := flag.Int("rows", 100000, "components per document")
	report := flag.String("report", "", "write structural metrics on window close")
	exitAfter := flag.Duration("exit-after", 0, "automatically finish after this duration")
	flag.Parse()
	if *count < 96 {
		fmt.Fprintln(os.Stderr, "rows must be at least 96")
		os.Exit(2)
	}
	go func() {
		w := new(app.Window)
		w.Option(app.Title("Vellum P0 — Gio"), app.Size(unit.Dp(1280), unit.Dp(820)))
		ui := &workspace{s: model.New(*count), w: w, theme: material.NewTheme(), rows: make(map[int]*rowState), started: time.Now(), report: *report, treeExpanded: true}
		ui.theme.TextSize = unit.Sp(14)
		ui.theme.Palette.ContrastBg = color.NRGBA{R: 35, G: 87, B: 170, A: 255}
		ui.table.Axis = layout.Vertical
		ui.tree.Axis = layout.Vertical
		ui.paletteList.Axis = layout.Vertical
		ui.filter.SingleLine = true
		ui.value.SingleLine = true
		ui.query.SingleLine = true
		ui.value.SetText(ui.s.Selected().Value)
		if *exitAfter > 0 {
			ui.exitAt = time.Now().Add(*exitAfter)
			time.AfterFunc(*exitAfter, w.Invalidate)
		}
		if err := ui.run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}()
	app.Main()
}

func (ui *workspace) run() error {
	defer ui.s.Close()
	defer ui.saveReport()
	for {
		switch e := ui.w.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			if !ui.exitAt.IsZero() && !time.Now().Before(ui.exitAt) {
				ui.saveReport()
				ui.w.Perform(system.ActionClose)
			}
			gtx := app.NewContext(&ui.ops, e)
			ui.s.Poll()
			ui.events(gtx)
			ui.frame++
			ui.currentRows = 0
			ui.draw(gtx)
			ui.maxRows = max(ui.maxRows, ui.currentRows)
			for id, state := range ui.rows {
				if state.seen < ui.frame-2 && !state.button.Pressed() && !gtx.Focused(&state.button) {
					delete(ui.rows, id)
				}
			}
			ui.maxStates = max(ui.maxStates, len(ui.rows))
			e.Frame(gtx.Ops)
			if ui.frame == 1 {
				fmt.Printf("Gio first frame submitted; %d projected rows; %.1f ms since workspace constructed\n", len(ui.s.View), float64(time.Since(ui.started).Microseconds())/1000)
			}
		}
	}
}
func (ui *workspace) execute(id string) {
	if err := ui.s.Execute(id, ui.value.Text(), ui.w.Invalidate); err != nil {
		ui.s.Status = err.Error()
	}
	ui.value.SetText(ui.s.Selected().Value)
}
func (ui *workspace) selectID(id int) { ui.s.Select(id); ui.value.SetText(ui.s.Selected().Value) }
func (ui *workspace) events(gtx layout.Context) {
	for {
		_, ok := ui.value.Update(gtx)
		if !ok {
			break
		}
	}
	for {
		_, ok := ui.query.Update(gtx)
		if !ok {
			break
		}
	}
	for ui.treeToggle.Clicked(gtx) {
		ui.treeExpanded = !ui.treeExpanded
	}
	for {
		ev, ok := gtx.Event(key.Filter{Name: "P", Required: key.ModShortcut}, key.Filter{Name: "X", Required: key.ModShortcut | key.ModShift}, key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		e, ok := ev.(key.Event)
		if !ok || e.State != key.Press {
			continue
		}
		if e.Name == "X" {
			ui.saveReport()
			ui.w.Perform(system.ActionClose)
		} else if e.Name == "P" {
			ui.palette = !ui.palette
			gtx.Execute(key.FocusCmd{Tag: &ui.query})
		} else {
			ui.palette = false
		}
	}
	for ui.paletteButton.Clicked(gtx) {
		ui.palette = !ui.palette
		gtx.Execute(key.FocusCmd{Tag: &ui.query})
	}
	for i := range ui.tabs {
		for ui.tabs[i].Clicked(gtx) {
			ui.s.Switch(i)
			ui.value.SetText(ui.s.Selected().Value)
		}
	}
	for i, c := range model.Commands {
		for ui.buttons[i].Clicked(gtx) {
			ui.execute(c.ID)
		}
	}
	for {
		e, ok := ui.filter.Update(gtx)
		if !ok {
			break
		}
		if _, ok := e.(widget.ChangeEvent); ok {
			ui.s.SetFilter(ui.filter.Text())
			ui.table.Position.First = 0
		}
	}
}

func (ui *workspace) label(gtx layout.Context, s string, size unit.Sp) layout.Dimensions {
	return material.Label(ui.theme, size, s).Layout(gtx)
}
func (ui *workspace) button(gtx layout.Context, b *widget.Clickable, text string) layout.Dimensions {
	style := material.Button(ui.theme, b, text)
	style.Inset = layout.Inset{Top: 6, Bottom: 6, Left: 10, Right: 10}
	return style.Layout(gtx)
}
func gap(d unit.Dp) layout.FlexChild { return layout.Rigid(layout.Spacer{Width: d, Height: d}.Layout) }
func (ui *workspace) draw(gtx layout.Context) layout.Dimensions {
	paint.Fill(gtx.Ops, color.NRGBA{R: 245, G: 247, B: 250, A: 255})
	return layout.UniformInset(12).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return ui.label(gtx, "Vellum / Gio baseline", 20) }), gap(16),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return ui.button(gtx, &ui.tabs[0], "Board A") }), gap(6),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return ui.button(gtx, &ui.tabs[1], "Board B") }), gap(16),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return ui.button(gtx, &ui.paletteButton, "Commands ⌘P") }),
				)
			}), gap(10),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return material.Editor(ui.theme, &ui.filter, "Filter components by reference, value or net").Layout(gtx)
					}), gap(12),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return ui.button(gtx, &ui.buttons[2], "Sort reference") }), gap(6),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions { return ui.button(gtx, &ui.buttons[3], "Sort value") }),
				)
			}), gap(10),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.X = gtx.Dp(180)
						return ui.navigation(gtx)
					}), gap(12),
					layout.Flexed(1, ui.center), gap(12),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Max.X = gtx.Dp(240)
						return ui.inspector(gtx)
					}),
				)
			}), gap(8),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return ui.label(gtx, fmt.Sprintf("%s · %s · %d/%d rows · revision %d · %s (%d%%)", ui.s.Doc().Name, ui.s.Selected().Reference, len(ui.s.View), len(ui.s.Doc().Rows), ui.s.Doc().Revision, ui.s.Status, ui.s.Progress()), 12)
			}),
		)
	})
}
func (ui *workspace) navigation(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := "▾ Components / first 64"
			if !ui.treeExpanded {
				label = "▸ Components / first 64"
			}
			return ui.button(gtx, &ui.treeToggle, label)
		}), gap(8),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			rows := ui.s.TreeRows()
			if !ui.treeExpanded {
				rows = nil
			}
			return ui.tree.Layout(gtx, len(rows), func(gtx layout.Context, i int) layout.Dimensions {
				if ui.treeButtons[i].Clicked(gtx) {
					ui.selectID(rows[i].ID)
				}
				return ui.button(gtx, &ui.treeButtons[i], "  └ "+rows[i].Reference+"  "+rows[i].Value)
			})
		}),
	)
}
func (ui *workspace) center(gtx layout.Context) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.label(gtx, "Board canvas · 96 component glyphs", 14)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.Y = gtx.Dp(220)
			gtx.Constraints.Max.Y = gtx.Constraints.Min.Y
			return ui.canvas(gtx)
		}), gap(12),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.label(gtx, "Reference     Value       Package     Rotation      Net        X / Y", 13)
		}), gap(6),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return ui.table.Layout(gtx, len(ui.s.View), func(gtx layout.Context, i int) layout.Dimensions {
				ui.currentRows++
				r := ui.s.View[i]
				state := ui.rows[r.ID]
				if state == nil {
					state = &rowState{}
					ui.rows[r.ID] = state
				}
				state.seen = ui.frame
				if state.button.Clicked(gtx) {
					ui.selectID(r.ID)
				}
				return state.button.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					if r.ID == ui.s.Doc().SelectedID {
						paint.FillShape(gtx.Ops, color.NRGBA{R: 214, G: 230, B: 255, A: 255}, clip.Rect{Max: image.Pt(gtx.Constraints.Max.X, gtx.Dp(28))}.Op())
					}
					return layout.Inset{Top: 5, Bottom: 5, Left: 6}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return ui.label(gtx, fmt.Sprintf("%-12s %-10s %-10s %3d°    %-9s %2.0f / %2.0f", r.Reference, r.Value, r.Package, r.Rotation, r.Net, r.X, r.Y), 13)
					})
				})
			})
		}),
	)
}
func (ui *workspace) canvas(gtx layout.Context) layout.Dimensions {
	size := image.Pt(gtx.Constraints.Max.X, gtx.Constraints.Max.Y)
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, color.NRGBA{R: 23, G: 43, B: 57, A: 255})
	for i := 0; i < 13; i++ {
		x := i * size.X / 12
		paint.FillShape(gtx.Ops, color.NRGBA{R: 45, G: 69, B: 80, A: 255}, clip.Rect{Min: image.Pt(x, 0), Max: image.Pt(x+1, size.Y)}.Op())
	}
	for i := 0; i < 9; i++ {
		y := i * size.Y / 8
		paint.FillShape(gtx.Ops, color.NRGBA{R: 45, G: 69, B: 80, A: 255}, clip.Rect{Min: image.Pt(0, y), Max: image.Pt(size.X, y+1)}.Op())
	}
	for i, r := range ui.s.CanvasRows() {
		x := (i%12)*size.X/12 + size.X/48
		y := (i/12)*size.Y/8 + size.Y/32
		w := size.X / 24
		h := max(3, size.Y/16)
		col := color.NRGBA{R: 67, G: 197, B: 166, A: 255}
		if r.ID == ui.s.Doc().SelectedID {
			col = color.NRGBA{R: 255, G: 190, B: 69, A: 255}
		}
		if r.Rotation%180 != 0 {
			w, h = h, w
		}
		paint.FillShape(gtx.Ops, col, clip.Rect{Min: image.Pt(x, y), Max: image.Pt(x+w, y+h)}.Op())
	}
	return layout.Dimensions{Size: size}
}
func (ui *workspace) inspector(gtx layout.Context) layout.Dimensions {
	children := []layout.FlexChild{
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.label(gtx, "Inspector · "+ui.s.Selected().Reference, 18)
		}), gap(12),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return ui.label(gtx, fmt.Sprintf("Package %s · rotation %d°", ui.s.Selected().Package, ui.s.Selected().Rotation), 13)
		}), gap(12),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Editor(ui.theme, &ui.value, "Value").Layout(gtx)
		}), gap(10),
	}
	for _, i := range []int{1, 0, 4, 5} {
		i := i
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !ui.s.Enabled(model.Commands[i].ID) {
				gtx = gtx.Disabled()
			}
			return ui.button(gtx, &ui.buttons[i], model.Commands[i].Label)
		}), gap(8))
	}
	if ui.palette {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions { return ui.label(gtx, "Command palette", 16) }), gap(8), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return material.Editor(ui.theme, &ui.query, "Search commands").Layout(gtx)
		}), gap(8))
		for i, c := range model.Commands {
			if !strings.Contains(strings.ToLower(c.Label+" "+c.ID), strings.ToLower(ui.query.Text())) {
				continue
			}
			i, c := i, c
			children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if ui.paletteCommands[i].Clicked(gtx) {
					ui.execute(c.ID)
					ui.palette = false
				}
				if !ui.s.Enabled(c.ID) {
					gtx = gtx.Disabled()
				}
				return ui.button(gtx, &ui.paletteCommands[i], c.Label)
			}), gap(4))
		}
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}
func (ui *workspace) saveReport() {
	r := map[string]any{"candidate": "gio", "rows_per_document": len(ui.s.Doc().Rows), "frames": ui.frame, "max_row_layout_calls_per_frame": ui.maxRows, "max_retained_row_states": ui.maxStates, "final_selected_id": ui.s.Doc().SelectedID, "final_status": ui.s.Status, "notes": "Structural counts from an interactive smoke run; not a performance benchmark or presentation timestamp."}
	b, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(b))
	if ui.report != "" {
		os.WriteFile(ui.report, append(b, '\n'), 0644)
	}
}
