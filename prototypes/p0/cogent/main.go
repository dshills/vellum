package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"cogentcore.org/core/colors"
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"cogentcore.org/core/xyz/xyzcore"
	"example.com/vellum/p0/model"
)

type view struct {
	s              *model.Session
	index          int
	frame          *core.Frame
	rows           []model.Row
	table          *core.Table
	treeFrame      *core.Frame
	value          *core.TextField
	title, status  *core.Text
	canvas         *core.Canvas
	commands       [6]*core.Button
	palette        *core.Frame
	paletteShown   bool
	query          *core.TextField
	paletteButtons [6]*core.Button
	wake           func()
	filter         *core.TextField
}

func main() {
	count := flag.Int("rows", 100000, "components per document")
	report := flag.String("report", "", "write structural metrics on window close")
	gpuProbe := flag.Bool("gpu-probe", false, "run public shared-device 3D integration probe")
	trace := flag.Bool("trace", false, "log native event routing for P0 diagnostics")
	exitAfter := flag.Duration("exit-after", 0, "automatically finish after this duration")
	flag.Parse()
	core.DebugSettings.EventTrace = *trace
	core.DebugSettings.KeyEventTrace = *trace
	if *gpuProbe {
		runGPUProbe(*exitAfter)
		return
	}
	if *count < 96 {
		fmt.Fprintln(os.Stderr, "rows must be at least 96")
		os.Exit(2)
	}
	started := time.Now()
	s := model.New(*count)
	defer s.Close()
	b := core.NewBody("Vellum P0 — Cogent Core")
	b.Styler(func(st *styles.Style) { st.Min.Set(units.Dp(1180), units.Dp(760)); st.Grow.Set(1, 1) })
	core.NewText(b).SetText("Vellum / Cogent Core baseline").SetType(core.TextHeadlineSmall)
	core.NewButton(b).SetText("Finish smoke run (⌘⇧X)").SetShortcut("Meta+Shift+X").OnClick(func(e events.Event) { go core.TheApp.QuitReq() })
	commandBar := core.NewFrame(b)
	commandBar.Styler(func(st *styles.Style) { st.Direction = styles.Row })
	tabs := core.NewTabs(b).SetType(core.FunctionalTabs)
	tabs.SetCloseIcon("")
	tabs.Styler(func(st *styles.Style) { st.Grow.Set(1, 1) })
	var source *events.Source
	var pending, closed atomic.Bool
	wake := func() {
		if !closed.Load() && source != nil && pending.CompareAndSwap(false, true) {
			source.Custom("p0.worker")
		}
	}
	views := make([]*view, 2)
	var saveReport func()
	for i := 0; i < 2; i++ {
		s.Switch(i)
		fr, _ := tabs.NewTab(s.Doc().Name)
		v := &view{s: s, index: i, frame: fr, wake: wake}
		views[i] = v
		v.build()
		fr.OnShow(func(e events.Event) {
			_, current := tabs.CurrentTab()
			if current == v.index {
				s.Switch(v.index)
				v.sync()
			}
		})
	}
	shortcuts := []key.Chord{"Meta+R", "Meta+E", "Meta+1", "Meta+2", "Meta+B", "Meta+K"}
	for i, c := range model.Commands {
		i, c := i, c
		core.NewButton(commandBar).SetText(c.Label).SetShortcut(shortcuts[i]).OnClick(func(e events.Event) { _, current := tabs.CurrentTab(); views[current].execute(c.ID) })
	}
	core.NewButton(commandBar).SetText("Filter ⌘0").SetShortcut("Meta+0").OnClick(func(e events.Event) {
		fmt.Println("P0 focus filter")
		_, current := tabs.CurrentTab()
		views[current].activate()
		views[current].filter.Defer(func() { views[current].filter.SetFocus() })
	})
	core.NewButton(commandBar).SetText("Value ⌘I").SetShortcut("Meta+I").OnClick(func(e events.Event) {
		_, current := tabs.CurrentTab()
		views[current].activate()
		views[current].value.Defer(func() { views[current].value.SetFocus() })
	})
	core.NewButton(commandBar).SetText("Palette ⌘P").SetShortcut("Meta+P").OnClick(func(e events.Event) {
		fmt.Println("P0 palette toggle")
		_, current := tabs.CurrentTab()
		v := views[current]
		v.paletteShown = !v.paletteShown
		v.palette.Update()
		if v.paletteShown {
			v.palette.Defer(func() { v.query.SetFocus() })
		}
	})
	for i, shortcut := range []key.Chord{"Meta+3", "Meta+4"} {
		i := i
		core.NewButton(commandBar).SetText(fmt.Sprintf("Board %c", 'A'+i)).SetShortcut(shortcut).OnClick(func(e events.Event) {
			tabs.SelectTabIndex(i)
			views[i].activate()
			views[i].sync()
		})
	}
	s.Switch(0)
	b.OnShow(func(e events.Event) {
		if source != nil {
			return
		}
		source = b.SystemEvents()
		b.Scene.On(events.Custom, func(e events.Event) {
			ce, ok := e.(*events.CustomEvent)
			if !ok {
				return
			}
			if ce.Data == "p0.exit" {
				saveReport()
				go core.TheApp.QuitReq()
				e.SetHandled()
				return
			}
			if ce.Data != "p0.worker" {
				return
			}
			pending.Store(false)
			s.Poll()
			_, current := tabs.CurrentTab()
			views[current].syncStatus()
			e.SetHandled()
		})
		_, current := tabs.CurrentTab()
		s.Switch(current)
		views[current].value.SetFocus()
		if *exitAfter > 0 {
			time.AfterFunc(*exitAfter, func() {
				if !closed.Load() {
					source.Custom("p0.exit")
				}
			})
		}
		fmt.Printf("Cogent workspace shown; %d projected rows; %.1f ms since fixture construction started\n", len(s.View), float64(time.Since(started).Microseconds())/1000)
	})
	saveReport = func() {
		_, current := tabs.CurrentTab()
		r := map[string]any{"candidate": "cogent", "rows_per_document": *count, "final_selected_id": s.Doc().SelectedID, "final_status": s.Status, "visible_rows": views[current].table.VisibleRows, "table_child_count": views[current].table.NumChildren(), "notes": "Structural counts from interactive smoke run; not a performance benchmark or native accessibility test."}
		data, _ := json.MarshalIndent(r, "", "  ")
		fmt.Println(string(data))
		if *report != "" {
			os.WriteFile(*report, append(data, '\n'), 0644)
		}
	}
	b.OnClose(func(e events.Event) { closed.Store(true); s.Close(); saveReport() })
	b.NewWindow().SetUseMinSize(true).RunMain()
}

func column(st *styles.Style) { st.Direction = styles.Column; st.Grow.Set(1, 1) }
func (v *view) build() {
	bar := core.NewFrame(v.frame)
	bar.Styler(func(st *styles.Style) { st.Direction = styles.Row; st.Grow.X = 1 })
	filter := core.NewTextField(bar).SetPlaceholder("Filter by reference, value or net")
	v.filter = filter
	filter.Styler(func(st *styles.Style) { st.Grow.X = 1 })
	filter.OnInput(func(e events.Event) { v.activate(); v.s.SetFilter(filter.Text()); v.sync() })
	core.NewButton(bar).SetText("Focus filter").OnClick(func(e events.Event) { v.activate(); filter.Defer(func() { filter.SetFocus() }) })
	for _, i := range []int{2, 3} {
		i := i
		core.NewButton(bar).SetText(model.Commands[i].Label).OnClick(func(e events.Event) { v.execute(model.Commands[i].ID) })
	}
	core.NewButton(bar).SetText("Commands").OnClick(func(e events.Event) {
		v.paletteShown = !v.paletteShown
		v.palette.Update()
		if v.paletteShown {
			v.palette.Defer(func() { v.query.SetFocus() })
		}
	})
	panes := core.NewSplits(v.frame).SetSplits(0.18, 0.62, 0.20)
	panes.Styler(func(st *styles.Style) { st.Grow.Set(1, 1); st.Min.Y = units.Dp(550) })
	v.treeFrame = core.NewFrame(panes)
	v.treeFrame.Styler(func(st *styles.Style) { column(st); st.Overflow.Set(styles.OverflowAuto) })
	center := core.NewFrame(panes)
	center.Styler(column)
	core.NewText(center).SetText("Board canvas · 96 component glyphs")
	v.canvas = core.NewCanvas(center).SetDraw(v.drawCanvas)
	v.canvas.Styler(func(st *styles.Style) {
		st.Min.Set(units.Dp(0), units.Dp(220))
		st.Max.Y = units.Dp(220)
		st.Grow.X = 1
	})
	v.rows = append([]model.Row(nil), v.s.View...)
	v.table = core.NewTable(center)
	v.table.SetReadOnly(true)
	v.table.SetSlice(&v.rows)
	v.table.Styler(func(st *styles.Style) { st.Grow.Set(1, 1); st.Min.Y = units.Dp(200) })
	v.table.OnSelect(func(e events.Event) {
		v.activate()
		i := v.table.SelectedIndex
		if i >= 0 && i < len(v.rows) {
			v.s.Select(v.rows[i].ID)
			v.syncInspector()
			v.syncStatus()
		}
	})
	inspector := core.NewFrame(panes)
	inspector.Styler(func(st *styles.Style) { column(st); st.Overflow.Set(styles.OverflowAuto) })
	v.title = core.NewText(inspector)
	v.value = core.NewTextField(inspector).SetPlaceholder("Value")
	core.NewButton(inspector).SetText("Focus value").OnClick(func(e events.Event) { v.activate(); v.value.Defer(func() { v.value.SetFocus() }) })
	for _, i := range []int{1, 0, 4, 5} {
		i := i
		v.commands[i] = core.NewButton(inspector).SetText(model.Commands[i].Label)
		v.commands[i].OnClick(func(e events.Event) { v.execute(model.Commands[i].ID) })
	}
	v.palette = core.NewFrame(inspector)
	v.palette.Styler(func(st *styles.Style) {
		column(st)
		st.Display = styles.Flex
		st.SetState(!v.paletteShown, states.Invisible)
		if !v.paletteShown {
			st.Display = styles.DisplayNone
		}
	})
	core.NewText(v.palette).SetText("Command palette")
	v.query = core.NewTextField(v.palette).SetPlaceholder("Search commands")
	v.query.OnInput(func(e events.Event) { v.palette.Update() })
	for i, c := range model.Commands {
		i, c := i, c
		btn := core.NewButton(v.palette).SetText(c.Label)
		v.paletteButtons[i] = btn
		btn.Styler(func(st *styles.Style) {
			st.Display = styles.Flex
			hidden := !strings.Contains(strings.ToLower(c.Label+" "+c.ID), strings.ToLower(v.query.Text()))
			st.SetState(hidden, states.Invisible)
			if hidden {
				st.Display = styles.DisplayNone
			}
			st.SetState(!v.s.Enabled(c.ID), states.Disabled)
		})
		btn.OnClick(func(e events.Event) { v.execute(c.ID); v.paletteShown = false; v.palette.Update() })
	}
	v.status = core.NewText(v.frame)
	v.rebuildTree()
	v.syncInspector()
	v.syncStatus()
}
func (v *view) execute(id string) {
	v.activate()
	fmt.Printf("P0 command %s document=%d selected=%d\n", id, v.index, v.s.Doc().SelectedID)
	if err := v.s.Execute(id, v.value.Text(), v.wake); err != nil {
		v.s.Status = err.Error()
	}
	if strings.HasPrefix(id, "component.") || strings.HasPrefix(id, "view.") {
		v.sync()
	} else {
		v.syncStatus()
	}
}
func (v *view) sync() {
	if v.filter.Text() != v.s.Filter {
		v.filter.SetText(v.s.Filter).Update()
	}
	v.rows = append(v.rows[:0], v.s.View...)
	v.table.SetSlice(&v.rows)
	v.table.ResetSelectedIndexes()
	v.table.SelectedIndex = -1
	for i, r := range v.rows {
		if r.ID == v.s.Doc().SelectedID {
			v.table.SelectedIndex = i
			break
		}
	}
	v.table.Update()
	v.rebuildTree()
	v.syncInspector()
	v.syncStatus()
}
func (v *view) rebuildTree() {
	v.treeFrame.DeleteChildren()
	root := core.NewTree(v.treeFrame).SetText("Components / first 64")
	for _, r := range v.s.TreeRows() {
		r := r
		n := core.NewTree(root).SetText(r.Reference + "  " + r.Value)
		n.OnSelect(func(e events.Event) { v.activate(); v.s.Select(r.ID); v.syncInspector(); v.syncStatus() })
	}
	v.treeFrame.Update()
}

func (v *view) activate() {
	if v.s.Active != v.index {
		v.s.Switch(v.index)
	}
}
func (v *view) syncInspector() {
	r := v.s.Selected()
	v.title.SetText(fmt.Sprintf("Inspector · %s<br>Package %s · rotation %d°", r.Reference, r.Package, r.Rotation)).Update()
	v.value.SetText(r.Value).Update()
	v.canvas.NeedsRender()
}
func (v *view) syncStatus() {
	v.status.SetText(fmt.Sprintf("%s · %s · %d/%d rows · revision %d · %s (%d%%)", v.s.Doc().Name, v.s.Selected().Reference, len(v.s.View), len(v.s.Doc().Rows), v.s.Doc().Revision, v.s.Status, v.s.Progress())).Update()
	for i, btn := range v.commands {
		if btn != nil {
			btn.SetState(!v.s.Enabled(model.Commands[i].ID), states.Disabled)
			btn.Update()
		}
	}
}
func (v *view) drawCanvas(pc *paint.Painter) {
	pc.Stroke.Color = nil
	pc.Fill.Color = colors.Uniform(color.RGBA{23, 43, 57, 255})
	pc.Rectangle(0, 0, 1, 1)
	pc.Draw()
	pc.Stroke.Color = colors.Uniform(color.RGBA{45, 69, 80, 255})
	pc.Stroke.Width = units.Px(1)
	pc.Fill.Color = nil
	for i := 0; i < 13; i++ {
		x := float32(i) / 12
		pc.Line(x, 0, x, 1)
		pc.Draw()
	}
	for i := 0; i < 9; i++ {
		y := float32(i) / 8
		pc.Line(0, y, 1, y)
		pc.Draw()
	}
	pc.Stroke.Color = nil
	for i, r := range v.s.CanvasRows() {
		x := float32(i%12)/12 + 1.0/48
		y := float32(i/12)/8 + 1.0/32
		w, h := float32(1.0/24), float32(1.0/16)
		col := color.RGBA{67, 197, 166, 255}
		if r.ID == v.s.Doc().SelectedID {
			col = color.RGBA{255, 190, 69, 255}
		}
		if r.Rotation%180 != 0 {
			w, h = h, w
		}
		pc.Fill.Color = colors.Uniform(col)
		pc.Rectangle(x, y, w, h)
		pc.Draw()
	}
}

// This supplemental probe exercises the documented public direct-render path.
// It is not the equivalent 2D fixture and is not used for speed comparisons.
func runGPUProbe(exitAfter time.Duration) {
	b := core.NewBody("Vellum P0 — Cogent shared GPU probe")
	core.NewText(b).SetText("Public xyzcore direct renderer · drag/scroll to navigate")
	sw := xyzcore.NewScene(b)
	sw.Styler(func(st *styles.Style) { st.Min.Set(units.Dp(640), units.Dp(420)); st.Grow.Set(1, 1) })
	sc := sw.XYZ
	xyz.NewAmbient(sc, "ambient", 0.3, xyz.DirectSun)
	xyz.NewDirectional(sc, "sun", 1, xyz.DirectSun).Pos.Set(0, 2, 1)
	mesh := xyz.NewBox(sc, "cube", 1, 1, 1)
	tree.AddChild(sc, func(s *xyz.Solid) { s.SetMesh(mesh).SetColor(color.RGBA{67, 197, 166, 255}).SetPos(0, 0, 0) })
	sc.Camera.Pose.Pos.Set(0, 1, 4)
	sc.Camera.LookAt(math32.Vector3{}, math32.Vec3(0, 1, 0))
	var shown atomic.Bool
	b.OnShow(func(e events.Event) {
		if shown.CompareAndSwap(false, true) {
			fmt.Println("Cogent public direct-render GPU scene shown")
			if exitAfter > 0 {
				time.AfterFunc(exitAfter, core.TheApp.QuitReq)
			}
		}
	})
	b.RunMainWindow()
}
