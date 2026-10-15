//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/drivers/cogent/internal/acceptance"
)

func exerciseComposed(ctx context.Context, a *app.Application, id string, w *acceptance.Workspace, cycle int) error {
	value, err := ownerProbe(ctx, a, func() ownerReply { return ownerReply{err: errors.Join(w.Exercise(), a.Invalidate(id))} })
	if err != nil || value.err != nil {
		return errors.Join(err, value.err)
	}
	if err = wait(ctx, func(inner context.Context) (bool, error) {
		value, e := ownerProbe(inner, a, func() ownerReply {
			grid, ge := cogent.InspectGrid(a, w.Grid)
			outline, te := cogent.InspectTree(a, w.Tree)
			if ge != nil || te != nil {
				return ownerReply{err: errors.Join(ge, te)}
			}
			if len(grid.Cells) == 0 || len(outline.Rows) == 0 || outline.Selected != "n1" {
				return ownerReply{err: errors.New("composed models lack completed geometry")}
			}
			return ownerReply{}
		})
		if e != nil {
			return false, e
		}
		if errors.Is(value.err, cogent.ErrGridLayout) || errors.Is(value.err, cogent.ErrTreeLayout) {
			return false, nil
		}
		return value.err == nil, value.err
	}); err != nil {
		return err
	}
	treeAXRows := 0
	if err = wait(ctx, func(inner context.Context) (bool, error) {
		var geometry cogent.TreeInspection
		reply, e := ownerProbe(inner, a, func() ownerReply {
			var err error
			geometry, err = cogent.InspectTree(a, w.Tree)
			return ownerReply{err: err}
		})
		if e != nil {
			return false, e
		}
		if errors.Is(reply.err, cogent.ErrTreeLayout) {
			return false, nil
		}
		if reply.err != nil {
			return false, reply.err
		}
		native, e := cogent.InspectAccessibility(inner, id)
		if e != nil {
			return false, e
		}
		ready, rows := treeEnvelopeCompleted(geometry, native)
		treeAXRows = rows
		return ready, nil
	}); err != nil {
		return fmt.Errorf("composed tree AX completion: %w", err)
	}
	nativeActions, err := exerciseNativeActions(ctx, a, id, w)
	if err != nil {
		return err
	}
	if len(nativeActions) != nativeActionPlanSteps {
		return fmt.Errorf("native plan %d completed %d of %d steps", nativeActionPlanVersion, len(nativeActions), nativeActionPlanSteps)
	}

	var inspection cogent.AccessibilityInspection
	root, clip := [4]float64{}, [4]float64{}
	// Scope publication can precede the first main-scene geometry after a modal
	// closes. Await completed nonempty container geometry before observing labels.
	if err = wait(ctx, func(inner context.Context) (bool, error) {
		var e error
		inspection, e = cogent.InspectAccessibility(inner, id)
		if e != nil {
			return false, e
		}
		root, clip = [4]float64{}, [4]float64{}
		for _, element := range inspection.Elements {
			if element.DomainID == "root" {
				root = element.Frame
			}
			if element.DomainID == "density-scroll" {
				clip = element.Frame
			}
		}
		return root[2] > 0 && root[3] > 0 && clip[2] > 0 && clip[3] > 0, nil
	}); err != nil {
		return fmt.Errorf("composed density geometry: %w", err)
	}

	for _, e := range inspection.Elements {
		if e.DomainID == "root" {
			root = e.Frame
		}
		if e.DomainID == "density-scroll" {
			clip = e.Frame
		}
	}
	defined, visible := 0, 0
	for _, e := range inspection.Elements {
		key := e.DomainID
		if key == "" {
			key = e.ID
		}
		if len(key) != 10 || !strings.HasPrefix(key, "dense-") {
			continue
		}
		defined++
		if frameInside(e.Frame, root) && frameInside(e.Frame, clip) {
			visible++
		}
	}
	if defined != acceptance.DenseLabels {
		return fmt.Errorf("composed catalog contains %d of %d dense labels", defined, acceptance.DenseLabels)
	}
	// Every cycle emits bounded scalar evidence; no 1000-node dump per cycle.
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"event": "vellum.composed-cycle", "cycle": cycle + 1, "native_action_plan": nativeActionPlanVersion, "native_list_requests": 3, "native_list_rejected": true, "native_canvas_selections": 2, "native_text_accepted": true, "native_color_edits": 2, "native_breadcrumb_requests": 2, "native_breadcrumb_rejected": true, "native_breadcrumb_truncated": true, "native_actions": nativeActions, "native_tree_selection": true, "native_tree_disclosures": 2, "editor_history_replayed": true, "grid_tree_completed": true, "tree_accessibility_completed": true, "tree_accessibility_rows": treeAXRows, "modal_open_close": true, "dense_defined": defined, "dense_fully_visible": visible, "root_screen_points": root, "density_clip_screen_points": clip, "all_dense_labels_fully_visible": visible == acceptance.DenseLabels, "requested_logical_size": [2]int{1920, 1080}, "full_p11_qualified": false})
}
func frameInside(frame, bounds [4]float64) bool {
	return frame[2] > 0 && frame[3] > 0 && bounds[2] > 0 && bounds[3] > 0 && frame[0] >= bounds[0] && frame[1] >= bounds[1] && frame[0]+frame[2] <= bounds[0]+bounds[2]+.01 && frame[1]+frame[3] <= bounds[1]+bounds[3]+.01
}

// The native inspector must contain the same completed bounded tree view as
// the owner inspection. It never counts last-readable pending records as ready.
func treeEnvelopeCompleted(tree cogent.TreeInspection, native cogent.AccessibilityInspection) (bool, int) {
	if !native.Reachable || !native.NativeWindow || tree.Generation == 0 {
		return false, 0
	}
	var host cogent.AccessibilityElement
	for _, e := range native.Elements {
		if e.DomainID == tree.ID {
			host = e
			break
		}
	}
	current := func(e cogent.AccessibilityElement) bool {
		return e.ID != "" && e.ObjectIdentity != 0 && !e.GridPending && e.GridScope == host.ID && e.GridIncarnation == tree.Incarnation && e.GridGeneration == tree.Generation && e.GridTargetGeneration == tree.Generation && e.Frame[2] > 0 && e.Frame[3] > 0
	}
	if host.Role != "AXOutline" || !current(host) || host.RowCount != tree.VisibleRows {
		return false, 0
	}
	rows := map[string]cogent.AccessibilityElement{}
	for _, e := range native.Elements {
		if e.GridScope == host.ID && e.Role == "AXRow" {
			if _, duplicate := rows[e.DomainID]; duplicate {
				return false, 0
			}
			rows[e.DomainID] = e
		}
	}
	count := 0
	for _, row := range tree.Rows {
		if row.Bounds.Empty() {
			continue
		}
		e, ok := rows[row.ID]
		if !ok || !current(e) || e.Parent != host.ID || e.Name != row.Label || e.Selected != row.Selected || row.HasChildren && e.Expanded != row.Expanded || e.RowIndex != [2]int{row.Index, 1} || e.RowCount != tree.VisibleRows {
			return false, count
		}
		count++
	}
	return count > 0 && count == len(rows), count
}
