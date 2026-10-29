// External-consumer compile contract for the explicit native v1 facade.
// This guards source shapes, not native behavior, defaults or platform support.
package cogent_test

import (
	"context"
	"image/color"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/command"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/editor"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
)

var (
	_ app.Driver                                                                        = cogent.Driver{}
	_ func(cogent.Driver, context.Context, *app.Application) error                      = cogent.Driver.Run
	_ func(app.WindowRef) error                                                         = cogent.CustomizeTitlebarToolbar
	_ func(cogent.SystemAppearance) ui.Theme                                            = cogent.MacSystemTheme
	_ func(context.Context, *app.Application, string, string, func(error)) error        = cogent.SettleTextField
	_ func(context.Context, *app.Application, string, *editor.Model, func(error)) error = cogent.SettleEditor
	_ func(context.Context, *app.Application, string, *virtual.Grid, func(error)) error = cogent.SettleGridEdit

	// Both directions preserve the public compatibility alias, not just a cast.
	_ *cogent.TitlebarToolbar          = (*cogent.TitlebarToolbarPrototype)(nil)
	_ *cogent.TitlebarToolbarPrototype = (*cogent.TitlebarToolbar)(nil)
)

// Never executed: compile the public field reads and writes as an application
// would. Do not invoke Run, AppKit, global appearance, or user preferences here.
func stableFacadeConsumer(d cogent.Driver, bar cogent.TitlebarToolbar, item cogent.TitlebarToolbarItem, appearance cogent.SystemAppearance) {
	var (
		_ func()                                               = d.OnQuitRequested
		_ ui.Theme                                             = d.Theme
		_ *command.Registry                                    = d.Commands
		_ []app.MenuGroup                                      = d.CommandMenus
		_ func(error)                                          = d.OnError
		_ bool                                                 = d.FollowSystemAppearance
		_ func(cogent.SystemAppearance)                        = d.OnSystemAppearance
		_ *cogent.TitlebarToolbar                              = d.TitlebarToolbar
		_ *cogent.TitlebarToolbarPrototype                     = d.TitlebarToolbarPrototype
		_ func(app.WindowRef) (*cogent.TitlebarToolbar, error) = d.TitlebarToolbars
		_ app.WindowRef                                        = bar.Window
		_ []cogent.TitlebarToolbarItem                         = bar.Items
		_ string                                               = bar.CustomizationID
		_ *command.Registry                                    = bar.Commands
		_ string                                               = item.ID
		_ string                                               = item.Label
		_ ui.Kind                                              = item.Kind
		_ []byte                                               = item.PNG
		_ bool                                                 = item.Disabled
		_ func()                                               = item.Activate
		_ string                                               = item.Command
		_ func() error                                         = item.Available
		_ bool                                                 = appearance.Known
		_ bool                                                 = appearance.Dark
		_ bool                                                 = appearance.HighContrast
		_ bool                                                 = appearance.ReduceMotion
		_ bool                                                 = appearance.ReduceTransparency
		_ color.NRGBA                                          = appearance.Accent
	)
	_ = cogent.Driver{
		OnQuitRequested: func() {}, Theme: ui.Theme{}, Commands: (*command.Registry)(nil),
		CommandMenus: []app.MenuGroup{}, OnError: func(error) {},
		FollowSystemAppearance: true, OnSystemAppearance: func(cogent.SystemAppearance) {},
		TitlebarToolbar:          (*cogent.TitlebarToolbar)(nil),
		TitlebarToolbarPrototype: (*cogent.TitlebarToolbarPrototype)(nil),
		TitlebarToolbars:         func(app.WindowRef) (*cogent.TitlebarToolbar, error) { return nil, nil },
	}
	_ = cogent.TitlebarToolbar{Window: app.WindowRef{}, Items: []cogent.TitlebarToolbarItem{}, CustomizationID: string(""), Commands: (*command.Registry)(nil)}
	_ = cogent.TitlebarToolbarPrototype{Window: app.WindowRef{}, Items: []cogent.TitlebarToolbarItem{}, CustomizationID: string(""), Commands: (*command.Registry)(nil)}
	_ = cogent.TitlebarToolbarItem{ID: string(""), Label: string(""), Kind: ui.Kind(""), PNG: []byte{}, Disabled: bool(false), Activate: func() {}, Command: string(""), Available: func() error { return nil }}
	_ = cogent.SystemAppearance{Known: bool(false), Dark: bool(false), HighContrast: bool(false), ReduceMotion: bool(false), ReduceTransparency: bool(false), Accent: color.NRGBA{}}
}
