package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/dshills/vellum/ui"
)

const toolbarPreferenceLimit = 64 << 10

var personalToolbarIDs = []string{"toolbar-Small", "toolbar-Regular", "toolbar-Large", "toolbar-mixed"}

type macToolbarPreferences struct {
	Version  int                         `json:"version"`
	Layouts  map[string]ui.ToolbarLayout `json:"layouts"`
	Displays map[string]string           `json:"displays,omitempty"`
	Display  string                      `json:"display,omitempty"`
}

// Persistence belongs to this application. Portable components perform no I/O.
// Pure gallery constructors/tests stay isolated; only main opts into this file.
func (m *macToolbarGallery) enablePreferences() {
	dir, err := os.UserConfigDir()
	if err != nil {
		m.preferencesError = err
		return
	}
	appearance := "light"
	if m.dark {
		appearance = "dark"
	}
	m.loadPreferences(filepath.Join(dir, "Vellum", "MacWidgets", "toolbars-"+appearance+".json"))
}
func (m *macToolbarGallery) loadPreferences(path string) {
	m.preferencesPath = path
	m.preferencesError = nil
	defer func() {
		if m.preferencesError != nil {
			m.preferencesPath = ""
		}
	}()
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		m.preferencesError = err
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, toolbarPreferenceLimit+1))
	if err != nil {
		m.preferencesError = err
		return
	}
	if len(data) > toolbarPreferenceLimit {
		m.preferencesError = fmt.Errorf("toolbar preferences exceed 64 KiB")
		return
	}
	var value macToolbarPreferences
	if err = json.Unmarshal(data, &value); err != nil {
		m.preferencesError = err
		return
	}
	if value.Version != 1 {
		// Do not overwrite a newer schema with an older application.
		m.preferencesPath = ""
		m.preferencesError = fmt.Errorf("unsupported toolbar preference version %d; file left unchanged", value.Version)
		return
	}
	if len(value.Layouts) > len(personalToolbarIDs) {
		m.preferencesError = fmt.Errorf("too many saved toolbar layouts")
		return
	}
	next := make(map[string]ui.ToolbarLayout, len(value.Layouts))
	for id, layout := range value.Layouts {
		if !slices.Contains(personalToolbarIDs, id) {
			continue
		}
		catalogue := m.toolbar(id, macToolbarSizes[1], "both", false)
		if len(layout.Order) > ui.MaxMenuOptions || len(layout.Hidden) > ui.MaxMenuOptions {
			m.preferencesError = fmt.Errorf("saved toolbar layout exceeds item limit")
			return
		}
		// Catalogue evolution: remove retired/duplicate IDs, append new defaults.
		known := map[string]bool{}
		for _, item := range catalogue.Children {
			known[item.ID] = true
		}
		prune := func(ids []string) []string {
			out := []string{}
			seen := map[string]bool{}
			for _, key := range ids {
				if known[key] && !seen[key] {
					out = append(out, key)
					seen[key] = true
				}
			}
			return out
		}
		next[id] = ui.ToolbarLayout{Order: prune(layout.Order), Hidden: prune(layout.Hidden)}
	}
	if len(value.Displays) > len(personalToolbarIDs) {
		m.preferencesError = fmt.Errorf("too many toolbar display settings")
		return
	}
	displays := map[string]string{}
	for id, mode := range value.Displays {
		if slices.Contains(personalToolbarIDs, id) && toolbarDisplayValid(mode) {
			displays[id] = mode
		}
	}
	display := "both"
	if toolbarDisplayValid(value.Display) {
		display = value.Display
	}
	m.layouts, m.displays, m.display = next, displays, display
	m.preferencesError = nil
}
func (m *macToolbarGallery) savePreferences() {
	if m.preferencesPath == "" {
		return
	}
	value := macToolbarPreferences{Version: 1, Layouts: m.layouts, Displays: m.displays, Display: m.display}
	data, err := json.MarshalIndent(value, "", "  ")
	if err == nil && len(data) > toolbarPreferenceLimit {
		err = fmt.Errorf("toolbar preferences exceed 64 KiB")
	}
	if err == nil {
		err = writeToolbarPreferences(m.preferencesPath, append(data, '\n'))
	}
	m.preferencesError = err
}
func writeToolbarPreferences(path string, data []byte) (err error) {
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".toolbars-*.tmp")
	if err != nil {
		return err
	}
	name := f.Name()
	defer func() { f.Close(); os.Remove(name) }()
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
