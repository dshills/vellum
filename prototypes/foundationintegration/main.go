//go:build darwin && cgo

package main

import (
	"cogentcore.org/core/core"
	"cogentcore.org/core/events"
	"cogentcore.org/core/events/key"
	"cogentcore.org/core/math32"
	"cogentcore.org/core/paint"
	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/states"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/system/driver/desktop"
	"cogentcore.org/core/text/lines"
	"cogentcore.org/core/text/spell"
	"cogentcore.org/core/text/textcore"
	"cogentcore.org/core/tree"
	"cogentcore.org/core/xyz"
	"cogentcore.org/core/xyz/xyzcore"
	"encoding/json"
	"example.com/vellum/foundationcontracts/mailbox"
	"example.com/vellum/foundationintegration/inputbridge"
	"example.com/vellum/foundationintegration/macbridge"
	"example.com/vellum/foundationintegration/semantics"
	"example.com/vellum/foundationintegration/transport"
	"example.com/vellum/nativeinput/textbridge"
	"flag"
	"fmt"
	"github.com/go-gl/glfw/v3.4/glfw"
	"golang.org/x/sys/unix"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

const corpus = "Latin: office AV\nArabic: العربية 123\nHebrew: שלום 456\nIndic: नमस्ते\nCJK: 日本語 中文\nCombining: é å\nEmoji: 👩‍💻 👨‍👩‍👧‍👦 🇺🇳"

type sample struct {
	Name, Text   string
	Runes, Bytes int
	Editing      core.TextFieldEditingState
}
type idleSample struct {
	Seconds, CPUSeconds, CPUPercentOneCore float64
	CanvasRepaints                         uint64
	PaintSchedulerWakes                    uint64
	Note                                   string
}
type wake struct{}

type message struct {
	Idle          idleSample
	Kind          string
	Action        semantics.Action
	Width, Height int
	Input         inputbridge.Update
	Caret         image.Rectangle
	Step          int
}
type fixture struct {
	nativePump                             atomic.Pointer[transport.Pump[message]]
	pump                                   *transport.Pump[message]
	body                                   *core.Body
	source                                 *events.Source
	bridge                                 *macbridge.Bridge
	input                                  *inputbridge.Input
	inputAck                               uint64
	inputHistory                           []inputbridge.Update
	inputHistoryBytes, inputHistoryDropped int
	inputProbes                            []json.RawMessage
	inputSource                            string
	inputEnvironment                       json.RawMessage
	nativeInput                            bool
	widgets                                map[uint64]core.Widget
	roles                                  map[uint64]string
	field                                  *core.TextField
	editor                                 *textcore.Editor
	buffer                                 *lines.Lines
	statusWidget                           *core.Text
	rowsFrame                              *core.Frame
	treeFrame                              *core.Frame
	rowWidgets                             map[uint64]*core.Button
	treeRoot                               *core.Button
	enabled, expanded                      bool
	value, status                          string
	selected, page, rotations, commits     int
	focus                                  uint64
	shown                                  atomic.Bool
	closed                                 atomic.Bool
	paints                                 atomic.Uint64
	updates, actions, rejected             int
	samples                                []sample
	resizes                                []image.Point
	idle                                   []idleSample
	report                                 string
	launched                               time.Time
	shownMS                                float64
	ax                                     bool
	warmSpelling                           bool
	onDemand                               bool
}

func main() {
	report := flag.String("report", "", "write experiment report")
	exitAfter := flag.Duration("exit-after", 90*time.Second, "close experiment automatically")
	ax := flag.Bool("accessibility", true, "attach native AccessKit semantic adapter")
	idle := flag.Duration("idle-sample", 0, "sample process CPU over this interval after 3 seconds settling")
	lifecycle := flag.Bool("lifecycle", false, "exercise six public window resize requests")
	onDemand := flag.Bool("on-demand", true, "use experimental requested-paint scheduler")
	warmSpelling := flag.Bool("warm-spelling", false, "initialize optional spell checker explicitly for comparison")
	nativeInput := flag.Bool("native-input", true, "attach experimental single-field Cocoa input client")
	saturation := flag.Bool("saturation", false, "verify native refusal while ordered UI delivery is stalled")
	inputProbe := flag.Bool("input-probe", false, "exercise native protocol deterministically, not a live IME")
	trace := flag.Bool("trace", false, "log window lifecycle events")
	flag.Parse()
	if *saturation && !*nativeInput {
		panic("saturation requires native input")
	}
	core.DebugSettings.WindowEventTrace = *trace
	core.DebugSettings.KeyEventTrace = *trace
	f := &fixture{widgets: map[uint64]core.Widget{}, roles: map[uint64]string{}, rowWidgets: map[uint64]*core.Button{}, enabled: true, expanded: true, value: "10k", selected: 17, focus: 2, report: *report, launched: time.Now(), ax: *ax, warmSpelling: *warmSpelling, onDemand: *onDemand, nativeInput: *nativeInput}
	f.build()
	desktop.TheApp.NativeAttach = func(dw *desktop.Window) (desktop.NativeAttachment, error) {
		if dw.Glw.GetAttrib(glfw.Visible) != glfw.False {
			return nil, fmt.Errorf("attachment occurred after show")
		}
		if f.ax {
			initial := semantics.Snapshot{Focus: 1, Nodes: []semantics.Node{{ID: 1, Role: "window", Label: "Vellum Native Input"}}}
			f.bridge = macbridge.Attach(dw.Glw.GetCocoaView(), dw.Glw.GetCocoaWindow(), initial, func(a semantics.Action) {
				if !f.closed.Load() && f.nativePump.Load() != nil {
					if err := f.nativePump.Load().Send(message{Kind: "action", Action: a}); err != nil {
						fmt.Println("native action refused:", err)
					}
				}
			})
		}
		if f.ax && f.bridge == nil {
			return nil, fmt.Errorf("accessibility attachment failed")
		}
		if f.nativeInput {
			var pending *transport.Ticket[message]
			f.input = inputbridge.Attach(dw.Glw.GetCocoaView(), func(u inputbridge.Update) bool {
				pump := f.nativePump.Load()
				if pump == nil || f.closed.Load() {
					return false
				}
				var err error
				pending, err = pump.Preflight(message{Kind: "input", Input: u})
				return err == nil
			}, func() {
				if err := pending.Commit(); err != nil {
					panic(err)
				}
				pending = nil
			})
			if f.input == nil {
				return &attachment{f}, fmt.Errorf("native attachment failed")
			}
			f.inputSource = inputbridge.Source()
			f.inputEnvironment = json.RawMessage(inputbridge.Environment())
		}
		fmt.Println("native_attach before_show=true")
		return &attachment{f}, nil
	}
	f.body.OnShow(func(e events.Event) {
		if !f.shown.CompareAndSwap(false, true) {
			return
		}
		f.shownMS = float64(time.Since(f.launched).Microseconds()) / 1000
		f.source = f.body.SystemEvents()
		f.pump, _ = transport.New[message](mailbox.Limits{Slots: 256, PayloadBytes: 4 * 1024 * 1024, MessageBytes: 64 * 1024}, func() {
			e := &events.CustomEvent{}
			e.Typ = events.Custom
			e.Data = wake{}
			e.Init()
			e.SetUnique()
			f.source.Deque.Send(e)
		})
		f.nativePump.Store(f.pump)
		f.body.Scene.RenderWindow().SystemWindow.(*desktop.Window).SetOnDemand(f.onDemand)
		f.body.Scene.On(events.Custom, func(e events.Event) {
			ce, ok := e.(*events.CustomEvent)
			if !ok {
				return
			}
			_, ok = ce.Data.(wake)
			if !ok {
				return
			}
			e.SetHandled()
			if err := f.pump.Drain(16, f.processMessage); err != nil {
				panic(err)
			}
		})
		f.body.Scene.SetGeometry(false, image.Point{}, image.Pt(1120, 960), -1)
		f.body.Defer(func() {
			if f.nativeInput {
				f.field.SetNativeEditingListener(func(state core.TextFieldEditingState) {
					if !f.closed.Load() {
						if err := f.pump.Latest(message{Kind: "caret"}); err != nil {
							fmt.Println("caret refused:", err)
						}
					}
				})
			}
			if *inputProbe {
				f.field.SetText("A👩Z")
				f.field.SetEditingSelection(1, 1)
				f.field.SetFocus()
			}
			f.publish()
			fmt.Printf("foundation shown callback %.1f ms; adapter=%v\n", f.shownMS, f.ax)
		})
		if *exitAfter > 0 {
			time.AfterFunc(*exitAfter, func() {
				if !f.closed.Load() {
					f.send(message{Kind: "exit"})
				}
			})
		}
		if *inputProbe {
			for step := 1; step <= 9; step++ {
				step := step
				time.AfterFunc(time.Duration(step)*time.Second, func() {
					if !f.closed.Load() {
						f.send(message{Kind: "probe", Step: step})
					}
				})
			}
		}
		if *saturation {
			time.AfterFunc(10*time.Second, func() { f.send(message{Kind: "saturation"}) })
			time.AfterFunc(11*time.Second, func() { f.send(message{Kind: "probe", Step: 5}) })
			time.AfterFunc(12*time.Second, func() { f.send(message{Kind: "probe", Step: 1}) }) // close during marked text
		}
		if *lifecycle {
			for i, sz := range []image.Point{{1000, 850}, {1180, 980}, {950, 820}, {1120, 960}, {980, 840}, {1120, 960}} {
				sz := sz
				time.AfterFunc(time.Duration(i+1)*2*time.Second, func() {
					if !f.closed.Load() {
						f.send(message{Kind: "resize", Width: sz.X, Height: sz.Y})
					}
				})
			}
		}
		paintCounter := &f.body.Scene.RenderWindow().SystemWindow.(*desktop.Window).Paint.Wakes
		if *idle > 0 {
			time.AfterFunc(3*time.Second, func() {
				start := time.Now()
				cpu := cpuSeconds()
				paints := f.paints.Load()
				if f.closed.Load() {
					return
				}
				wakes := paintCounter.Load()
				time.AfterFunc(*idle, func() {
					if f.closed.Load() {
						return
					}
					s := idleSample{Seconds: time.Since(start).Seconds(), CPUSeconds: cpuSeconds() - cpu, CanvasRepaints: f.paints.Load() - paints, PaintSchedulerWakes: paintCounter.Load() - wakes, Note: "Whole-process CPU; canvas repaint proxy, not compositor presents or wakeups. Small fixture, not the 1000-control budget workload."}
					s.CPUPercentOneCore = s.CPUSeconds / s.Seconds * 100
					data, _ := json.Marshal(s)
					fmt.Printf("idle_sample %s\n", data)
					f.send(message{Kind: "idle", Idle: s})
				})
			})
		}
	})
	f.body.OnClose(func(e events.Event) {
		if !f.closed.CompareAndSwap(false, true) {
			return
		}
		if f.input != nil {
			core.TheApp.RunOnMain(func() { f.input.Close() })
		}
		if f.bridge != nil {
			core.TheApp.RunOnMain(func() { f.bridge.Close() })
		}
		if f.pump == nil {
			return
		} // Window may close before the first OnShow.
		f.pump.Close()
		for f.pump.Stats().Queue.Outstanding > 0 {
			if err := f.pump.Drain(16, f.processMessage); err != nil {
				panic(err)
			}
		}
		f.save()
	})
	f.body.NewWindow().SetUseMinSize(true).RunMain()
}
func cpuSeconds() float64 {
	var r unix.Rusage
	if unix.Getrusage(unix.RUSAGE_SELF, &r) != nil {
		return 0
	}
	return float64(r.Utime.Sec+r.Stime.Sec) + float64(r.Utime.Usec+r.Stime.Usec)/1e6
}
func (f *fixture) reg(id uint64, role string, w core.Widget) {
	f.widgets[id] = w
	f.roles[id] = role
	w.AsWidget().OnFocus(func(e events.Event) { f.focus = id; f.deferPublish() })
	w.AsWidget().OnFocusLost(func(e events.Event) {
		if f.focus == id {
			f.focus = 1
		}
		f.deferPublish()
	})
}
func (f *fixture) deferPublish() {
	if f.shown.Load() && !f.closed.Load() {
		f.body.Defer(f.publish)
		f.body.NeedsRender()
	}
}
func (f *fixture) button(parent core.Widget, id uint64, label, shortcut string, command string) *core.Button {
	b := core.NewButton(parent).SetText(label)
	if shortcut != "" {
		b.SetShortcut(key.Chord(shortcut))
	}
	b.OnClick(func(e events.Event) { f.command(command, ""); f.sync(); f.deferPublish() })
	f.reg(id, "button", b)
	return b
}
func (f *fixture) build() {
	b := core.NewBody("Vellum Native Input — composition")
	f.body = b
	b.Styler(func(s *styles.Style) { s.Grow.Set(1, 1); s.Min.Set(units.Dp(850), units.Dp(740)) })
	core.NewText(b).SetText("Vellum native input · single-field composition").SetType(core.TextHeadlineSmall)
	bar := core.NewFrame(b)
	bar.Styler(func(s *styles.Style) { s.Direction = styles.Row })
	f.button(bar, 2, "Rotate selected", "Meta+R", "rotate")
	f.button(bar, 3, "Apply value", "Meta+E", "commit")
	f.button(bar, 4, "Toggle enabled", "Meta+T", "toggle")
	f.roles[4] = "checkbox"
	f.button(bar, 5, "Next 8 rows", "Meta+N", "next")
	f.button(bar, 6, "Record text", "Meta+S", "capture")
	f.button(bar, 7, "Finish", "Meta+Shift+X", "quit")
	f.field = core.NewTextField(b).SetText("10k").SetPlaceholder("Component value / Unicode deletion probe")
	f.field.Styler(func(s *styles.Style) { s.Grow.X = 1; s.SetTextWrap(false) })
	f.field.OnKeyChord(func(e events.Event) { f.deferPublish() })
	f.reg(8, "text", f.field)
	f.field.OnInput(func(e events.Event) { f.deferPublish() })
	f.field.OnFocusLost(func(e events.Event) {

		f.deferPublish()
	})
	foc := core.NewFrame(b)
	foc.Styler(func(s *styles.Style) { s.Direction = styles.Row })
	focusButton := core.NewButton(foc).SetText("Focus single-line ⌘I").SetShortcut("Meta+I")
	focusButton.OnClick(func(e events.Event) { f.field.Defer(func() { f.field.SetFocus() }) })
	f.reg(9, "button", focusButton)
	focusEditor := core.NewButton(foc).SetText("Focus multiline ⌘M").SetShortcut("Meta+M")
	focusEditor.OnClick(func(e events.Event) { f.editor.Defer(func() { f.editor.SetFocus() }) })
	f.reg(10, "button", focusEditor)
	row := core.NewFrame(b)
	row.Styler(func(s *styles.Style) { s.Direction = styles.Row })
	f.treeFrame = core.NewFrame(row)
	f.treeFrame.Styler(func(s *styles.Style) { s.Grow.X = 1; s.Direction = styles.Column })
	f.reg(20, "tree", f.treeFrame)
	f.treeRoot = core.NewButton(f.treeFrame).SetText("Components · expanded")
	f.treeRoot.OnClick(func(e events.Event) { f.command("expand", ""); f.sync(); f.deferPublish() })
	f.reg(21, "treeitem", f.treeRoot)
	for i := 0; i < 2; i++ {
		i := i
		leaf := core.NewButton(f.treeFrame).SetText(fmt.Sprintf("R%d", 17+i))
		leaf.Styler(func(s *styles.Style) {
			s.Display = styles.Flex
			s.SetState(!f.expanded, states.Invisible)
			if !f.expanded {
				s.Display = styles.DisplayNone
			}
		})
		f.reg(uint64(22+i), "treeitem", leaf)
		leaf.OnClick(func(e events.Event) { f.selected = 17 + i; f.sync(); f.deferPublish() })
	}

	f.rowsFrame = core.NewFrame(row)
	f.rowsFrame.Styler(func(s *styles.Style) { s.Grow.X = 1; s.Direction = styles.Column })
	f.reg(30, "table", f.rowsFrame)
	f.rebuildRows()
	f.buffer = lines.NewLinesExact().SetText([]byte(corpus))
	f.editor = textcore.NewEditor(b).SetLines(f.buffer)
	if f.warmSpelling {
		f.editor.SpellCheckLineErrors(0)
	}
	f.editor.Styler(func(s *styles.Style) { s.Min.Y = units.Dp(150); s.Max.Y = units.Dp(180); s.Grow.X = 1 })
	f.reg(40, "multiline", f.editor)
	f.editor.OnInput(func(e events.Event) { f.deferPublish() })
	canvas := core.NewCanvas(b).SetDraw(func(pc *paint.Painter) {
		f.paints.Add(1)
		pc.Fill.Color = image.NewUniform(color.RGBA{20, 44, 57, 255})
		pc.Stroke.Color = nil
		pc.Rectangle(0, 0, 1, 1)
		pc.Draw()
		for i := 0; i < 2; i++ {
			pc.Fill.Color = image.NewUniform(color.RGBA{67, 197, 166, 255})
			pc.Rectangle(float32(i)*0.4+0.1, 0.25, 0.2, 0.5)
			pc.Draw()
		}
	})
	canvas.Styler(func(s *styles.Style) { s.Min.Y = units.Dp(70); s.Max.Y = units.Dp(70); s.Grow.X = 1 })
	f.reg(50, "tree", canvas)
	// Reuse the public shared-device GPU path exercised in P0, now under resize.
	sw := xyzcore.NewScene(b)
	sw.Styler(func(s *styles.Style) { s.Min.Y = units.Dp(100); s.Max.Y = units.Dp(140); s.Grow.X = 1 })
	sc := sw.XYZ
	xyz.NewAmbient(sc, "ambient", 0.3, xyz.DirectSun)
	xyz.NewDirectional(sc, "sun", 1, xyz.DirectSun).Pos.Set(0, 2, 1)
	mesh := xyz.NewBox(sc, "cube", 1, 1, 1)
	tree.AddChild(sc, func(s *xyz.Solid) { s.SetMesh(mesh).SetColor(color.RGBA{67, 197, 166, 255}) })
	sc.Camera.Pose.Pos.Set(0, 1, 4)
	sc.Camera.LookAt(math32.Vector3{}, math32.Vec3(0, 1, 0))
	f.statusWidget = core.NewText(b)
	f.reg(60, "status", f.statusWidget)
	f.sync()
}
func (f *fixture) command(cmd, value string) {
	switch cmd {
	case "rotate":
		if !f.enabled {
			f.rejected++
			f.status = "Rotate unavailable while disabled"
			return
		}
		f.rotations++
		f.status = "Rotated selected component"
	case "commit":
		if value == "" {
			value = f.field.Text() // Explicit commit; inspection uses EditingState instead.
		}
		if strings.TrimSpace(value) == "" {
			f.rejected++
			f.status = "Error: value must not be empty"
			return
		}
		f.value = value
		f.commits++
		f.status = "Committed value " + value
	case "toggle":
		f.enabled = !f.enabled
		f.status = "Enabled state changed"
	case "next":
		f.page = (f.page + 8) % 100000
		f.rebuildRows()
		f.status = "Virtual range changed"
	case "expand":
		f.expanded = !f.expanded
		f.status = "Tree expansion changed"
	case "capture":
		f.recordInputProbe(0)
		f.samples = append(f.samples, sample{Name: fmt.Sprintf("sample-%d", len(f.samples)+1), Text: f.field.EditingState().Text, Runes: utf8.RuneCountInString(f.field.EditingState().Text), Bytes: len(f.field.EditingState().Text), Editing: f.field.EditingState()})
		data, _ := json.Marshal(f.samples[len(f.samples)-1])
		fmt.Printf("text_capture %s\n", data)
		state := f.field.EditingState()
		f.status = fmt.Sprintf("Text recorded cursor %d selection %d..%d", state.Cursor, state.Anchor, state.Focus)
	case "quit":
		f.save()
		go core.TheApp.QuitReq()
	}
	fmt.Printf("command %s selected=%d rotations=%d commits=%d rejected=%d\n", cmd, f.selected, f.rotations, f.commits, f.rejected)
}
func (f *fixture) handle(a semantics.Action) {
	f.actions++
	snap := f.snapshot()
	if !snap.Allows(a) {
		f.rejected++
		f.status = "Rejected stale or unsupported accessibility action"
		f.sync()
		f.deferPublish()
		return
	}
	if a.Kind == "focus" {
		if w := f.widgets[a.Target]; w != nil {
			w.AsWidget().SetFocus()
			f.focus = a.Target
		}
		f.deferPublish()
		return
	}
	switch {
	case a.Target == 2:
		f.command("rotate", "")
	case a.Target == 3:
		f.command("commit", "")
	case a.Target == 4:
		f.command("toggle", "")
	case a.Target == 5:
		f.command("next", "")
	case a.Target == 6:
		f.command("capture", "")
	case a.Target == 7:
		f.command("quit", "")
	case a.Target == 8 && a.Kind == "set_selection":
		run, err := textbridge.New(f.field.EditingState().Text)
		sel := a.Selection
		valid := err == nil && sel != nil && sel.Anchor.Node == 80 && sel.Focus.Node == 80
		anchor, focus := 0, 0
		if valid {
			var okA, okF bool
			anchor, okA = run.RuneAtCharacter(sel.Anchor.Character)
			focus, okF = run.RuneAtCharacter(sel.Focus.Character)
			valid = okA && okF
		}
		if !valid || !f.field.SetEditingSelection(anchor, focus) {
			f.rejected++
			f.status = "Rejected invalid text selection"
		} else {
			f.field.SetFocus()
			f.focus = 8
			f.status = "Native text selection changed"
		}
	case a.Target == 8 && a.Kind == "set_value":
		f.field.SetText(a.Value).Update()
		f.field.SetFocus()
		f.focus = 8
		f.status = "Accessible field value changed"
	case a.Target == 9:
		f.field.SetFocus()
	case a.Target == 10:
		f.editor.SetFocus()
	case a.Target == 40 && a.Kind == "set_value":
		f.buffer.SetText([]byte(a.Value))
		f.editor.Update()
		f.status = "Multiline value replaced"
	case a.Target == 21:
		if a.Kind == "click" {
			f.command("expand", "")
		} else {
			f.expanded = a.Kind != "collapse"
		}
	case a.Target == 22 || a.Target == 23:
		f.selected = int(a.Target) - 5
	case a.Target == 51 || a.Target == 52:
		f.selected = int(a.Target) - 34
	case a.Target >= 1000 && a.Target < 101000:
		f.selected = int(a.Target - 999)
		f.status = fmt.Sprintf("Selected R%d", f.selected)
	default:
		f.rejected++
	}
	f.sync()
	f.deferPublish()
}
func (f *fixture) rebuildRows() {
	for id := range f.rowWidgets {
		delete(f.widgets, id)
		delete(f.roles, id)
	}
	f.rowWidgets = map[uint64]*core.Button{}
	f.rowsFrame.DeleteChildren()
	for i := 0; i < 8; i++ {
		idx := f.page + i
		id := uint64(1000 + idx)
		b := core.NewButton(f.rowsFrame).SetType(core.ButtonText).SetText(fmt.Sprintf("R%d · 0603", idx+1))
		f.rowWidgets[id] = b
		f.reg(id, "row", b)
		b.OnClick(func(e events.Event) {
			f.selected = idx + 1
			f.status = fmt.Sprintf("Selected R%d", f.selected)
			f.sync()
			f.deferPublish()
		})
	}
	f.rowsFrame.Update()
	if f.focus >= 1000 {
		f.widgets[5].AsWidget().SetFocus()
	}
}
func (f *fixture) sync() {
	f.statusWidget.SetText(fmt.Sprintf("R%d · value %s · rotations %d · commits %d · enabled %v · rows %d–%d/100000 · %s", f.selected, f.value, f.rotations, f.commits, f.enabled, f.page+1, f.page+8, f.status)).Update()
	text := "Components · collapsed"
	if f.expanded {
		text = "Components · expanded"
	}
	f.treeRoot.SetText(text).Update()
	f.treeFrame.Update()
}
func boolp(b bool) *bool { return &b }
func (f *fixture) snapshot() semantics.Snapshot {
	scale := float64(1)
	if rw := f.body.Scene.RenderWindow(); rw != nil {
		scale = float64(rw.SystemWindow.(*desktop.Window).DevicePixelRatio)
	}
	if scale <= 0 {
		scale = 1
	}
	s := semantics.Snapshot{Focus: f.focus}
	root := semantics.Node{ID: 1, Role: "window", Label: "Vellum foundation", Bounds: [4]float64{0, 0, float64(f.body.Scene.SceneGeom.Size.X) / scale, float64(f.body.Scene.SceneGeom.Size.Y) / scale}, Children: []uint64{2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 30, 40, 50, 60}}
	s.Nodes = append(s.Nodes, root)
	for _, id := range []uint64{2, 3, 4, 5, 6, 7, 8, 9, 10, 20, 21, 22, 23, 30, 40, 50, 60} {
		if (id == 22 || id == 23) && !f.expanded {
			continue
		}
		w := f.widgets[id]
		bb := w.AsWidget().Geom.TotalBBox
		n := semantics.Node{ID: id, Role: f.roles[id], Bounds: [4]float64{float64(bb.Min.X) / scale, float64(bb.Min.Y) / scale, float64(bb.Max.X) / scale, float64(bb.Max.Y) / scale}}
		if b, ok := w.(*core.Button); ok {
			n.Label = b.Text
			n.Actions = []string{"click", "focus"}
		}
		switch id {
		case 2:
			if !f.enabled {
				n.Actions = []string{"focus"}
			}
		case 4:
			n.Checked = boolp(f.enabled)
		case 8:
			n.Label = "Component value"
			n.Invalid = strings.HasPrefix(f.status, "Error:")
			n.Value = f.field.EditingState().Text
			n.Actions = []string{"set_value", "focus"}
			state := f.field.EditingState()
			if run, err := textbridge.New(state.Text); err == nil {
				anchor, okA := run.CharacterAtRune(state.Anchor)
				focus, okF := run.CharacterAtRune(state.Focus)
				if okA && okF {
					n.Children = []uint64{80}
					n.Actions = append(n.Actions, "set_selection")
					n.TextSelection = &semantics.TextSelection{Anchor: semantics.TextPosition{Node: 80, Character: anchor}, Focus: semantics.TextPosition{Node: 80, Character: focus}}
					s.Nodes = append(s.Nodes, semantics.Node{ID: 80, Role: "textrun", Value: state.Text, CharacterLengths: run.ByteLengths, Bounds: n.Bounds})
				}
			}
		case 20:
			n.Label = "Component tree"
			n.Children = []uint64{21}
		case 21:
			n.Expanded = boolp(f.expanded)
			if f.expanded {
				n.Children = []uint64{22, 23}
			}
			n.Actions = []string{"click", "focus", "expand", "collapse"}
		case 22, 23:
			n.Selected = boolp(f.selected == int(id)-5)
		case 30:
			n.Label = "Components virtual table"
			n.RowCount = 100000
			for i := 0; i < 8; i++ {
				n.Children = append(n.Children, uint64(1000+f.page+i))
			}
		case 40:
			n.Label = "Multilingual text editor"
			n.Value = string(f.buffer.Text())
			n.Actions = []string{"focus", "set_value"}
		case 50:
			n.Label = "Schematic objects"
			n.Children = []uint64{51, 52}
		case 60:
			n.Label = f.statusWidget.Text
		}
		s.Nodes = append(s.Nodes, n)
	}
	for i := 0; i < 8; i++ {
		idx := f.page + i
		id := uint64(1000 + idx)
		bb := f.rowWidgets[id].Geom.TotalBBox
		s.Nodes = append(s.Nodes, semantics.Node{ID: id, Role: "row", Label: fmt.Sprintf("R%d package 0603", idx+1), RowIndex: idx, Selected: boolp(f.selected == idx+1), Actions: []string{"click", "focus"}, Bounds: [4]float64{float64(bb.Min.X) / scale, float64(bb.Min.Y) / scale, float64(bb.Max.X) / scale, float64(bb.Max.Y) / scale}})
	}
	for i := 0; i < 2; i++ {
		bb := f.widgets[50].AsWidget().Geom.TotalBBox
		w, h := float64(bb.Dx())/scale, float64(bb.Dy())/scale
		x, y := float64(bb.Min.X)/scale, float64(bb.Min.Y)/scale
		s.Nodes = append(s.Nodes, semantics.Node{ID: uint64(51 + i), Role: "treeitem", Label: fmt.Sprintf("Canvas component R%d", 17+i), Selected: boolp(f.selected == 17+i), Actions: []string{"click"}, Bounds: [4]float64{x + w*(float64(i)*0.4+0.1), y + h*0.25, x + w*(float64(i)*0.4+0.3), y + h*0.75}})
	}
	found := false
	for _, n := range s.Nodes {
		found = found || n.ID == s.Focus
	}
	if !found {
		s.Focus = 1
	}
	return s
}
func (f *fixture) publish() {
	if f.closed.Load() || f.bridge == nil && f.input == nil {
		return
	}
	s := f.snapshot()
	if err := s.Validate(); err != nil {
		fmt.Println("snapshot validation:", err)
		return
	}
	state := f.field.EditingState()
	a, _ := inputbridge.UTF16Offset(state.Text, state.Anchor)
	b, _ := inputbridge.UTF16Offset(state.Text, state.Focus)
	markStart, markEnd, marked := f.field.NativeCompositionRange()
	if marked {
		markStart, _ = inputbridge.UTF16Offset(state.Text, markStart)
		markEnd, _ = inputbridge.UTF16Offset(state.Text, markEnd)
	}
	rect := f.caretPoints(f.field.NativeCaretBounds())
	active, ack := f.field.StateIs(states.Focused), f.inputAck
	core.TheApp.RunOnMain(func() {
		if f.bridge != nil {
			f.bridge.Focus(f.body.Scene.RenderWindow().SystemWindow.(*desktop.Window).Glw.GetAttrib(glfw.Focused) == glfw.True)
			f.bridge.Update(s)
		}
		if f.input != nil {
			f.input.Update(state.Text, a, b, markStart, markEnd, active, ack, rect)
		}
	})
	f.updates++
}
func (f *fixture) save() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	beforeGC := mem.HeapAlloc
	runtime.GC()
	runtime.ReadMemStats(&mem)
	s := f.snapshot()
	r := map[string]any{"idle_samples": f.idle, "transport": f.pump.Stats(), "native_input_enabled": f.nativeInput, "input_source": f.inputSource, "input_environment": f.inputEnvironment, "input_environment_scope": "captured at first show; each recorded native probe reports its current source", "input_updates": f.inputHistory, "input_history_bytes": f.inputHistoryBytes, "input_history_dropped": f.inputHistoryDropped, "input_probes": f.inputProbes, "native_input_ack": f.inputAck, "platform": "darwin/arm64", "editing_state": f.field.EditingState(), "experimental_toolkit_patch": true, "spell_checker_initialized": spell.Spell != nil, "on_demand": f.onDemand, "paint_scheduler_wakes_total": f.body.Scene.RenderWindow().SystemWindow.(*desktop.Window).Paint.Wakes.Load(), "shown_callback_ms": f.shownMS, "presentation_timestamp": false, "accessibility_enabled": f.ax, "semantic_node_count": len(s.Nodes), "virtual_row_total": 100000, "virtual_rows_materialized": 8, "selected_id": f.selected, "value": f.value, "field_text": f.field.EditingState().Text, "multiline_text": string(f.buffer.Text()), "rotations": f.rotations, "commits": f.commits, "rejected_actions": f.rejected, "native_actions": f.actions, "semantic_updates": f.updates, "canvas_repaints_proxy": f.paints.Load(), "text_samples": f.samples, "resize_content_sizes_pixels": f.resizes, "go_heap_alloc_bytes_after_forced_gc": mem.HeapAlloc, "go_heap_alloc_bytes_before_gc": beforeGC, "resident_bytes_after_gc": macbridge.ResidentBytes(), "go_heap_inuse_bytes": mem.HeapInuse, "goroutines": runtime.NumGoroutine(), "status": f.status, "limitations": []string{"small fixture, not full performance workload", "single-line logical selection and current IME caret only; arbitrary text-range geometry, bidi and multiline ranges absent", "candidate-window placement and full screen-reader workflows require separate live evidence; deterministic protocol calls are not live IME tests", "Windows/Linux untested"}}
	if f.bridge != nil {
		r["adapter_active"] = f.bridge.Active()
		r["adapter_snapshot_requests"] = f.bridge.Activations.Load()
		r["adapter_action_callbacks"] = f.bridge.Actions.Load()
	}
	data, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(data))
	if f.report != "" {
		if file, err := os.Create(f.report + ".heap.pprof"); err == nil {
			pprof.WriteHeapProfile(file)
			file.Close()
		}
		if err := os.WriteFile(f.report, append(data, '\n'), 0644); err != nil {
			fmt.Println(err)
		}
	}
}

func (f *fixture) caretPoints(caret image.Rectangle) [4]float64 {
	scale := float64(f.body.Scene.RenderWindow().SystemWindow.(*desktop.Window).DevicePixelRatio)
	if scale <= 0 {
		scale = 1
	}
	return [4]float64{float64(caret.Min.X) / scale, float64(caret.Min.Y) / scale, float64(caret.Dx()) / scale, float64(caret.Dy()) / scale}
}
func (f *fixture) applyInput(u inputbridge.Update) {
	if u.Sequence <= f.inputAck {
		f.rejected++
		return
	}
	decoded, err := inputbridge.Decode(u)
	if err != nil || !f.field.ApplyNativeEdit(u.Text, decoded.RuneAnchor, decoded.RuneFocus, decoded.RuneMarkedStart, decoded.RuneMarkedEnd, u.Kind) {
		f.rejected++
		f.status = "Rejected native input update"
		fmt.Printf("input_rejected %+v error=%v\n", u, err)
	} else {
		f.recordInput(u)
		f.status = fmt.Sprintf("Native %s · sequence %d", u.Kind, u.Sequence)
		fmt.Printf("input_update %+v\n", u)
	}
	f.inputAck = u.Sequence
	f.sync()
	f.deferPublish()
}

func (f *fixture) recordInputProbe(step int) {
	if f.input == nil {
		return
	}
	var data string
	core.TheApp.RunOnMain(func() { data = f.input.Probe(step) })
	if json.Valid([]byte(data)) {
		f.inputProbes = append(f.inputProbes, json.RawMessage(data))
		fmt.Printf("input_probe %s\n", data)
	}
}

func (f *fixture) processMessage(m message) {
	switch m.Kind {
	case "idle":
		f.idle = append(f.idle, m.Idle)
	case "saturation":
		core.TheApp.RunOnMain(func() {
			before := f.input.Probe(1) // accepted mark, before stall
			admitted := 0
			for f.pump.Send(message{Kind: "noop"}) == nil {
				admitted++
			}
			var states []json.RawMessage
			for _, step := range []int{2, 3, 5, 7, 9} {
				states = append(states, json.RawMessage(f.input.Probe(step)))
			}
			data, _ := json.Marshal(map[string]any{"before": json.RawMessage(before), "refused_states": states, "fill_admitted": admitted, "transport": f.pump.Stats()})
			fmt.Printf("saturation %s\n", data)
			if f.report != "" {
				path := filepath.Join(filepath.Dir(f.report), "saturation.json")
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					panic(err)
				}
				if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
					panic(err)
				}
			}
		})
	case "action":
		f.handle(m.Action)
	case "exit":
		f.save()
		go core.TheApp.QuitReq()
	case "input":
		f.applyInput(m.Input)
	case "caret":
		// Publish one complete current text/selection/geometry snapshot. A queued
		// rectangle alone can belong to text older than optimistic native state.
		f.publish()
	case "probe":
		f.recordInputProbe(m.Step)
	case "resize":
		f.body.Scene.SetGeometry(false, image.Point{}, image.Pt(m.Width, m.Height), -1)
		f.body.Defer(func() { f.resizes = append(f.resizes, f.body.Scene.SceneGeom.Size); f.publish(); f.recordInputProbe(0) })
	}
}
func (f *fixture) send(m message) {
	if err := f.pump.Send(m); err != nil {
		fmt.Printf("message refused kind=%s: %v\n", m.Kind, err)
		if m.Kind == "exit" {
			go core.TheApp.QuitReq()
		}
	}
}

// These callbacks run on the Cocoa thread and never inspect widgets.
type attachment struct{ f *fixture }

func (a *attachment) Focus(focused bool) {
	if a.f.bridge != nil {
		a.f.bridge.Focus(focused)
	}
}
func (a *attachment) Close() {
	if a.f.input != nil {
		a.f.input.Close()
	}
	if a.f.bridge != nil {
		a.f.bridge.Close()
	}
	fmt.Println("native_detach before_gpu_view_release=true")
}

// Diagnostic history has separate retention limits from delivery credit.
func (f *fixture) recordInput(u inputbridge.Update) {
	f.inputHistory = append(f.inputHistory, u)
	f.inputHistoryBytes += len(u.Text) + len(u.Kind)
	for len(f.inputHistory) > 128 || f.inputHistoryBytes > 2*1024*1024 {
		f.inputHistoryBytes -= len(f.inputHistory[0].Text) + len(f.inputHistory[0].Kind)
		f.inputHistory[0] = inputbridge.Update{}
		f.inputHistory = f.inputHistory[1:]
		f.inputHistoryDropped++
	}
}
