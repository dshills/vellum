// docking-smoke exercises live native float/redock/close with owner-confined
// shared document state. -interactive leaves the workbench available for
// physical pointer/tab menu acceptance checks; all OS preferences stay intact.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"cogentcore.org/core/system"
	"example.com/vellum/app"
	cogent "example.com/vellum/cogent"
	"example.com/vellum/ui"
	"example.com/vellum/workbench"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Returning errors unwinds application and temporary session cleanup before
// main reports failure; os.Exit is confined to the outer entry point.
func run(args []string) error {
	flags := flag.NewFlagSet("docking-smoke", flag.ContinueOnError)
	interactive := flags.Bool("interactive", false, "leave native UI open for pointer docking/tab menu checks")
	deadline := flags.Duration("timeout", 20*time.Second, "smoke deadline; use a longer duration with -interactive")
	path := flags.String("session", "", "session autosave/restore file (temporary by default)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	if *path == "" {
		directory, err := os.MkdirTemp("", "vellum-docking-smoke-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(directory)
		*path = filepath.Join(directory, "session.json")
	}
	initial := workbench.SessionLayout{Version: workbench.SessionVersion, Windows: []workbench.WindowLayout{{ID: "main", Title: "Vellum Docking Workbench", Bounds: workbench.Bounds{Width: 900, Height: 640}, Root: &workbench.Group{ID: "root", Axis: workbench.Horizontal, Ratio: .7, First: &workbench.Group{ID: "documents", Tabs: []string{"board", "notes"}, Active: "board"}, Second: &workbench.Group{ID: "inspector", Tabs: []string{"inspector"}, Active: "inspector"}}}}, Focus: workbench.Focus{Window: "main", Panel: "board"}}
	session, err := workbench.NewSession(initial, []string{"board", "notes", "inspector"})
	if err != nil {
		return err
	}
	text := "Document state survives floating and redocking"
	edits := 0
	panels := []cogent.Panel{{ID: "board", Title: "Board", View: func() ui.Node {
		return ui.Node{ID: "board", Kind: ui.Column, Children: []ui.Node{{ID: "help", Kind: ui.Label, Text: "Drag tabs to reorder, to group edges to split, or outside to float. Right-click a tab to redock across windows."}, {ID: "revision", Kind: ui.Label, Text: fmt.Sprintf("Shared document revision: %d", edits)}}}
	}}, {ID: "notes", Title: "Notes", View: func() ui.Node {
		return ui.Node{ID: "notes", Kind: ui.Column, Children: []ui.Node{{ID: "editor", Kind: ui.TextField, Text: text, Change: func(change ui.ChangeEvent) { text = change.Text; edits++; a.Invalidate("") }}}}
	}}, {ID: "inspector", Title: "Inspector", View: func() ui.Node {
		return ui.Node{ID: "inspector", Kind: ui.Label, Text: fmt.Sprintf("Document: %s\nRevision: %d", text, edits)}
	}}}
	group, err := cogent.NewWorkbenchGroup(a, session, panels)
	if err != nil {
		return err
	}
	group.PersistencePath = *path
	ready := make(chan string, 32)
	failure := make(chan error, 1)
	group.OnWindowReady = func(id string) {
		nativeID, ok := group.WindowID(id)
		if ok {
			select {
			case ready <- nativeID:
			default:
			}
		}
	}
	group.OnError = func(err error) {
		select {
		case failure <- err:
		default:
		}
		a.Close()
	}
	if err := group.Open(); err != nil {
		return err
	}
	initialNativeID, ok := group.WindowID("main")
	if !ok {
		return errors.New("initial native registration missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), *deadline)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		finish := func(err error) { result <- err; a.Close() }
		latestReady := ""
		waitReady := func(expected string) error {
			nativeID, err := waitNativeReady(ctx, ready, failure, expected)
			if err == nil {
				latestReady = nativeID
			}
			return err
		}
		if err := waitReady(initialNativeID); err != nil {
			finish(err)
			return
		}
		if *interactive {
			fmt.Println("READY native docking workbench; physical pointer and tab context menus available")
			<-ctx.Done()
			result <- nil
			return
		}
		onOwner := func(fn func() error) error {
			reply := make(chan error, 1)
			if err := a.Post(func() { reply <- fn() }); err != nil {
				return err
			}
			select {
			case err := <-reply:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		expectedFloating := ""
		resolveFloating := func() error {
			var ok bool
			expectedFloating, ok = group.WindowID(session.Layout().Focus.Window)
			if !ok {
				return errors.New("floating native registration missing")
			}
			return nil
		}
		if err := onOwner(func() error {
			if err := group.Float("notes", workbench.Bounds{X: 160, Y: 120, Width: 640, Height: 480, Positioned: true}); err != nil {
				return err
			}
			return resolveFloating()
		}); err != nil {
			finish(err)
			return
		}
		if err := waitReady(expectedFloating); err != nil {
			finish(err)
			return
		}
		if err := onOwner(func() error {
			text = "Edited while floating"
			edits++
			a.Invalidate("")
			return group.Apply(func(s *workbench.Session) error { return s.Dock("notes", "main", "documents", 0) })
		}); err != nil {
			finish(err)
			return
		}
		if err := onOwner(func() error {
			if len(session.Layout().Windows) != 1 || text != "Edited while floating" || edits != 1 {
				return errors.New("redock lost document state")
			}
			if err := group.Float("notes", workbench.Bounds{Width: 640, Height: 480}); err != nil {
				return err
			}
			return resolveFloating()
		}); err != nil {
			finish(err)
			return
		}
		if err := waitReady(expectedFloating); err != nil {
			finish(err)
			return
		}
		// Request the real native close path from this worker, matching Cogent's
		// desktop OnCloseReq dispatch. CloseReq synchronously takes the native
		// render lock and runs Scene.Close; the driver's hook then enters a.Do.
		// Calling it inside onOwner would reenter that non-reentrant owner.
		// WindowByName locks the app list, and READY follows stable native name
		// and close-callback setup. The native hook returns these panels before
		// the floating registration is cancelled.
		native := system.TheApp.WindowByName(latestReady)
		if native == nil {
			finish(errors.New("floating native window missing"))
			return
		}
		native.CloseReq()
		for {
			done := false
			err := onOwner(func() error { done = len(session.Layout().Windows) == 1; return nil })
			if err != nil {
				finish(err)
				return
			}
			if done {
				break
			}
			select {
			case <-time.After(30 * time.Millisecond):
			case <-ctx.Done():
				finish(ctx.Err())
				return
			}
		}
		err := onOwner(func() error {
			snapshot := session.Layout()
			if snapshot.Focus.Window != "main" || snapshot.Focus.Panel != "notes" || text != "Edited while floating" {
				return errors.New("native close focus/state transfer")
			}
			copy, err := workbench.NewSession(initial, session.Panels())
			if err != nil {
				return err
			}
			if err := copy.Load(*path); err != nil {
				return err
			}
			if len(copy.Layout().Windows) != 1 || copy.Layout().Focus.Panel != "notes" {
				return errors.New("complete session persistence")
			}
			return nil
		})
		finish(err)
	}()
	err = a.Run(ctx, cogent.Driver{OnError: group.OnError})
	select {
	case check := <-result:
		if err == nil {
			err = check
		}
	case check := <-failure:
		if err == nil {
			err = check
		}
	case <-ctx.Done():
		if err == nil && !(*interactive) {
			err = ctx.Err()
		}
	}
	if err != nil {
		return err
	}
	if !*interactive {
		fmt.Println("PASS native float, cross-window redock, native close return, shared document state, focus transfer and complete session persistence")
	}
	return nil
}

// Native lifetime IDs are resolved on the owner after each accepted Float.
// Ignore every other readiness notification, including duplicate retired IDs.
func waitNativeReady(ctx context.Context, ready <-chan string, failure <-chan error, expected string) (string, error) {
	for {
		select {
		case id, ok := <-ready:
			if !ok {
				return "", errors.New("native readiness channel closed")
			}
			if id == expected {
				return id, nil
			}
		case err := <-failure:
			return "", err
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}
