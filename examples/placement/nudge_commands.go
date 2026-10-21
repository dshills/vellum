package main

import (
	"context"
	"fmt"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/ui"
)

func exactNudgeOptions() []ui.Option {
	out := []ui.Option{}
	for _, step := range []int{1, 10} {
		for _, dir := range []string{"left", "right", "up", "down"} {
			out = append(out, ui.Option{ID: fmt.Sprintf("nudge-%d-%s", step, dir), Label: fmt.Sprintf("%s %d unit(s)", map[string]string{"left": "Left", "right": "Right", "up": "Up", "down": "Down"}[dir], step)})
		}
	}
	return out
}
func (m *model) bindExactNudges() error {
	for _, step := range []int{1, 10} {
		for _, dir := range []struct {
			id     string
			dx, dy float64
		}{{"left", -1, 0}, {"right", 1, 0}, {"up", 0, -1}, {"down", 0, 1}} {
			label := fmt.Sprintf("Move selection %d unit(s)", step)
			if err := m.commands.Register(command.Command{ID: fmt.Sprintf("nudge-%d-%s", step, dir.id), Label: label, Available: m.editAvailable, Execute: func(context.Context) error { return m.nudgeWorld(float64(step)*dir.dx, float64(step)*dir.dy, label) }}); err != nil {
				return err
			}
		}
	}
	return nil
}
