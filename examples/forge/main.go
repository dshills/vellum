// Forge is an independently runnable consumer of public Vellum workbench APIs.
package main

import (
	"cogentcore.org/core/core"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
	"os"
	"time"
)

func main() {
	smoke := flag.Bool("smoke", false, "bounded app-local native rendering and input smoke")
	exitAfter := flag.Duration("exit-after", 0, "optional bounded lifetime, max 1m")
	writeIcon := flag.String("write-icon", "", "write original app icon to a new PNG, without launching")
	flag.Parse()
	if *writeIcon != "" {
		if e := writeAppIcon(*writeIcon); e != nil {
			fail(e)
		}
		return
	}
	if *exitAfter < 0 || *exitAfter > time.Minute {
		fail(fmt.Errorf("exit-after must be 0..1m"))
	}
	if e := run(*smoke, *exitAfter); e != nil {
		fail(e)
	}
}
func fail(e error) { fmt.Fprintln(os.Stderr, e); os.Exit(1) }
func run(smoke bool, exitAfter time.Duration) error {
	if smoke {
		old := core.DebugSettings.DisableWindowGeometrySaver
		core.DebugSettings.DisableWindowGeometrySaver = true
		defer func() { core.DebugSettings.DisableWindowGeometrySaver = old }()
	}
	a, e := app.New(app.Config{})
	if e != nil {
		return e
	}
	defer a.Close()
	m := newWorkbench()
	m.application = a
	m.surface.Owner = a
	m.surface.Window = "forge"
	ref, e := a.OpenWindow(app.Window{ID: "forge", Title: "Vellum Forge — Design Workbench", Width: 1440, Height: 920, DisableRootScrolling: true, Chrome: app.WindowChrome{Style: app.ChromeStandard, Appearance: app.AppearanceLight}, View: m.View, CloseRequest: m.requestClose})
	if e != nil {
		return e
	}
	m.ref = ref
	if !smoke {
		if e := m.bindFiles(); e != nil {
			return e
		}
		defer m.recovery.stop(m)
	}
	t := forgeTheme()
	toolbar := &cogent.TitlebarToolbar{Window: ref, Items: m.toolbar(), Commands: m.commands}
	result := make(chan error, 1)
	if smoke {
		go func() { result <- smokeCheck(a, ref, m); a.Close() }()
	}
	if exitAfter > 0 {
		timer := time.AfterFunc(exitAfter, a.Close)
		defer timer.Stop()
	}
	if smoke {
		timer := time.AfterFunc(60*time.Second, func() { fmt.Fprintln(os.Stderr, "Forge smoke watchdog"); os.Exit(1) })
		defer timer.Stop()
	}
	e = a.Run(context.Background(), cogent.Driver{Theme: t, Commands: m.commands, TitlebarToolbar: toolbar, OnError: func(e error) { fmt.Fprintln(os.Stderr, e) }})
	if e != nil {
		a.Close()
		if smoke {
			select {
			case smokeErr := <-result:
				return errors.Join(e, smokeErr)
			case <-time.After(time.Second):
				return errors.Join(e, fmt.Errorf("smoke cleanup timeout"))
			}
		}
		return e
	}
	if smoke {
		return <-result
	}
	return nil
}

func forgeTheme() ui.Theme {
	t := ui.MacTheme(false)
	for _, k := range []ui.Kind{ui.Column, ui.Row} {
		s := t.Kinds[k]
		s.Padding = ui.Some(ui.Insets{})
		s.Gap = ui.Some(float32(0))
		t.Kinds[k] = s
	}
	return t
}
