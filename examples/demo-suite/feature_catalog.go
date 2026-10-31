package main

type featureDemo struct{ ID, Name, Description, Bundle string }

var featureDemos = []featureDemo{
	{"forge", "Forge design workbench", "3D assembly, expression dimensions, move/resize handles, workspace persistence and Undo/Redo.", "Vellum Forge.app"},
	{"documents", "Documents — Save, Quit and recovery", "Two notebooks with independent toolbars, Save/Open, coordinated Quit, restart and interrupted-session recovery.", "Vellum Suite Documents.app"},
	{"notes", "Mac Notes", "Mac prose editor, folders, Find/Replace, Save/Open, recovery and native commands.", "Vellum Suite Notes.app"},
	{"reminders", "Reminders", "Controlled checklists, scheduled date ranges, detail editing and Undo/Redo.", "Vellum Reminders.app"},
	{"forma", "Forma poster studio", "Poster canvas, layers, color and size inspector, pan/zoom and Undo/Redo.", "Vellum Forma.app"},
	{"luma", "Luma image library", "Filmstrip, collections, favorites, search, color variants and image histogram.", "Vellum Luma.app"},
	{"pulse", "Pulse telemetry explorer", "Retained charts, virtual table, filtering and fictional service details.", "Vellum Pulse.app"},
	{"placement", "Palette → canvas", "Drag Block, Connector or Note onto the canvas. Pan/zoom, grid snapping, target rejection, selection, keyboard placement and Undo/Redo.", "Vellum Placement.app"},
	{"workspace", "Replayable workspace", "Interactive managed workspace with explicit scheduling, tasks, document state and replay infrastructure.", "Vellum Workspace.app"},
	{"gallery", "Controls gallery", "Breadcrumbs, color picker, closable tabs, images, split panes, icons, menus and progress shapes.", "Vellum Controls.app"},
	{"workbench", "Professional workbench", "Docking, floating windows, inspector, commands, undo/redo and cancellable tasks.", "Vellum Workbench.app"},
	{"editor", "Document workspace", "Two native editors with exact text persistence, Save/Open, find and unsaved-change handling.", "Vellum Editor.app"},
	{"grid", "Editable grid", "Frozen columns, captured row selection, variable heights and versioned cell edits.", "Vellum Grid.app"},
	{"loadedgrid", "Stored-data grid", "Real stored records with cancellable worker sorting/filtering and stable selection.", "Vellum Loaded Grid.app"},
	{"loadedtree", "Large indexed tree", "100,000 stored nodes, worker-built indexes and bounded native rows.", "Vellum Tree.app"},
	{"drawing", "GPU drawing", "Retained meshes, tiled images and shared-device canvas composition.", "Vellum Drawing.app"},
	{"imaging", "Compute + HDR imaging", "Compute, floating-point textures, exposure controls and asynchronous PNG capture/save.", "Vellum Imaging.app"},
	{"viewport3d", "3D viewport", "Custom shaders, depth testing, retained geometry and camera orbit/dolly.", "Vellum 3D.app"},
}

func featureByID(id string) featureDemo {
	for _, f := range featureDemos {
		if f.ID == id {
			return f
		}
	}
	panic("unknown feature demo: " + id)
}
