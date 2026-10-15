//go:build darwin && cgo && !offscreen && vellum_native_checks

package main

import (
	"math"
	"reflect"

	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/drivers/cogent/internal/acceptance"
)

type selectionListIdentity struct {
	host uint64
	rows map[string]uint64
}

func finiteNativeFrame(frame [4]float64) bool {
	for _, v := range frame {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return frame[2] >= 0 && frame[3] >= 0
}
func positiveNativeFrame(frame [4]float64) bool {
	return finiteNativeFrame(frame) && frame[2] > 0 && frame[3] > 0
}

func selectionListCompleted(s cogent.AccessibilityInspection, previous *selectionListIdentity) bool {
	host, ok := nativeElementRole(s, acceptance.SelectionListID, "AXList")
	if !s.Reachable || !s.NativeWindow || !ok || host.ID == "" || host.ObjectIdentity == 0 || !host.Enabled || !positiveNativeFrame(host.Frame) || host.Value != "Last component" {
		return false
	}
	names := map[string]string{"a": "Rejected first", "b": "Accepted initial", "locked": "Locked", "c": "Resistor", "d": "Capacitor", "e": "Connector", "f": "Label", "last": "Last component"}
	next := selectionListIdentity{host: host.ObjectIdentity, rows: map[string]uint64{}}
	identities := map[uint64]bool{host.ObjectIdentity: true}
	visible, selected := 0, 0
	for _, row := range s.Elements {
		if row.Parent != host.ID {
			continue
		}
		name, expected := names[row.DomainID]
		if !expected || row.ID == "" || next.rows[row.DomainID] != 0 || row.Role != "AXRow" || row.ObjectIdentity == 0 || identities[row.ObjectIdentity] || !finiteNativeFrame(row.Frame) || row.Name != name || row.Enabled != (row.DomainID != "locked") || row.Selected != (row.DomainID == "last") {
			return false
		}
		next.rows[row.DomainID] = row.ObjectIdentity
		identities[row.ObjectIdentity] = true
		if row.Selected {
			selected++
		}
		if positiveNativeFrame(row.Frame) {
			if !frameInside(row.Frame, host.Frame) {
				return false
			}
			visible++
		}
	}
	if len(next.rows) != 8 || selected != 1 || visible == 0 || visible >= 8 || previous.host != 0 && !reflect.DeepEqual(*previous, next) {
		return false
	}
	*previous = next
	return true
}

type selectionCanvasIdentity struct {
	host, left, right, incarnation uint64
	generation                     uint64
}

func selectionCanvasTarget(s cogent.AccessibilityInspection, e cogent.AccessibilityElement) bool {
	host, ok := nativeElementRole(s, acceptance.SelectionCanvasID, "AXGroup")
	return ok && host.ID != "" && host.GridScope == host.ID && !host.GridPending && host.GridIncarnation > 0 && host.GridGeneration > 0 && host.GridGeneration == host.GridTargetGeneration && positiveNativeFrame(host.Frame) &&
		e.Role == "AXButton" && e.ID != "" && e.Parent == host.ID && e.GridScope == host.ID && !e.GridPending && e.GridIncarnation == host.GridIncarnation && e.GridGeneration == host.GridGeneration && e.GridTargetGeneration == host.GridTargetGeneration && positiveNativeFrame(e.Frame) && frameInside(e.Frame, host.Frame)
}

func selectionCanvasCompleted(s cogent.AccessibilityInspection, selected string, previous *selectionCanvasIdentity) bool {
	host, ok := nativeElementRole(s, acceptance.SelectionCanvasID, "AXGroup")
	label := "Left object"
	if selected == acceptance.SelectionRightID {
		label = "Right object"
	}
	if selected != acceptance.SelectionLeftID && selected != acceptance.SelectionRightID || !s.Reachable || !s.NativeWindow || !ok || host.ID == "" || host.Name != "Selectable components" || host.Value != label || host.ObjectIdentity == 0 || !host.Enabled || !positiveNativeFrame(host.Frame) || host.GridScope != host.ID || host.GridIncarnation == 0 || host.GridPending || host.GridGeneration == 0 || host.GridGeneration != host.GridTargetGeneration {
		return false
	}
	next := selectionCanvasIdentity{host: host.ObjectIdentity, incarnation: host.GridIncarnation, generation: host.GridGeneration}
	children := 0
	for _, e := range s.Elements {
		if e.Parent != host.ID {
			continue
		}
		name := "Left object"
		if e.DomainID == acceptance.SelectionRightID {
			name = "Right object"
		} else if e.DomainID != acceptance.SelectionLeftID {
			return false
		}
		if e.ID == "" || e.Role != "AXButton" || e.ObjectIdentity == 0 || e.ObjectIdentity == host.ObjectIdentity || !e.Enabled || !positiveNativeFrame(e.Frame) || !frameInside(e.Frame, host.Frame) || e.Name != name || e.Selected != (e.DomainID == selected) || e.GridScope != host.ID || e.GridIncarnation != host.GridIncarnation || e.GridPending || e.GridGeneration != host.GridGeneration || e.GridTargetGeneration != host.GridGeneration {
			return false
		}
		if e.DomainID == acceptance.SelectionLeftID {
			if next.left != 0 {
				return false
			}
			next.left = e.ObjectIdentity
		} else {
			if next.right != 0 {
				return false
			}
			next.right = e.ObjectIdentity
		}
		children++
	}
	if children != 2 || next.left == 0 || next.right == 0 || next.left == next.right || previous.host != 0 && (previous.host != next.host || previous.left != next.left || previous.right != next.right || previous.incarnation != next.incarnation || next.generation <= previous.generation) {
		return false
	}
	*previous = next
	return true
}
