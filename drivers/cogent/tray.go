package cogent

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dshills/vellum/app"
	"image/color"
	"image/png"
	"strings"
	"unicode/utf8"
)

type nativeTrayItem struct {
	Label               string
	Index               int
	Disabled, Separator bool
}
type nativeTray struct {
	Title, Tooltip     string
	AccessibilityLabel string
	Template, Remove   bool
	Width, Height      int
	RGBA               []byte
	Items              []nativeTrayItem
}

func (s *platformServices) TrayCapabilities() app.TrayCapabilities { return nativeTrayCapabilities() }
func (s *platformServices) SetTray(ctx context.Context, o app.TrayOptions) error {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	if err := o.Validate(); err != nil {
		return err
	}
	desc, ids, err := s.captureTray(o)
	if err != nil {
		return err
	}
	return nativeSetTray(scope, desc, o.IconPNG, ids, s)
}

func (s *platformServices) captureTray(o app.TrayOptions) (nativeTray, []string, error) {
	desc := nativeTray{Title: o.Title, Tooltip: o.Tooltip, Template: o.Template}
	desc.AccessibilityLabel = o.AccessibilityLabel
	if strings.TrimSpace(desc.AccessibilityLabel) == "" {
		desc.AccessibilityLabel = o.Tooltip
	}
	if strings.TrimSpace(desc.AccessibilityLabel) == "" {
		desc.AccessibilityLabel = o.Title
	}
	var ids []string
	var captureErr error
	if err := s.owner.Do(func() {
		for _, item := range o.Menu {
			if item.Separator {
				desc.Items = append(desc.Items, nativeTrayItem{Separator: true, Index: -1, Disabled: true})
				continue
			}
			if s.commands == nil {
				captureErr = fmt.Errorf("cogent: tray command unavailable")
				return
			}
			cmd, found := s.commands.Lookup(item.Command)
			if found {
				if len(cmd.Label) > 1024 {
					captureErr = app.ErrServiceTooLarge
					return
				}
				if strings.TrimSpace(cmd.Label) == "" || !utf8.ValidString(cmd.Label) || strings.ContainsRune(cmd.Label, 0) {
					captureErr = fmt.Errorf("cogent: invalid tray command label")
					return
				}
				desc.Items = append(desc.Items, nativeTrayItem{Label: cmd.Label, Index: len(ids), Disabled: s.commands.Enabled(cmd.ID) != nil})
				ids = append(ids, cmd.ID)
			}
			if !found {
				captureErr = fmt.Errorf("cogent: tray command unavailable")
				return
			}
		}
	}); err != nil {
		return nativeTray{}, nil, err
	}
	if captureErr != nil {
		return nativeTray{}, nil, captureErr
	}
	return desc, ids, nil
}
func (s *platformServices) ClearTray(ctx context.Context) error {
	scope, cancel, err := s.serviceContext(ctx)
	if err != nil {
		return err
	}
	defer cancel()
	return nativeSetTray(scope, nativeTray{Remove: true}, nil, nil, s)
}

// Decode on a worker after bounded native admission, without retaining the caller
// PNG for a delayed Cocoa operation. Straight-alpha RGBA avoids native PNG decoding.
func trayPixels(desc *nativeTray, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if len(data) > app.TrayMaxPNGBytes {
		return app.ErrServiceTooLarge
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return err
	}
	if len(data) > app.TrayMaxPNGBytes || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > app.TrayMaxIconSize || cfg.Height > app.TrayMaxIconSize {
		return app.ErrServiceTooLarge
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}
	desc.Width, desc.Height = cfg.Width, cfg.Height
	desc.RGBA = make([]byte, 4*cfg.Width*cfg.Height)
	p := 0
	for y := 0; y < cfg.Height; y++ {
		for x := 0; x < cfg.Width; x++ {
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			desc.RGBA[p], desc.RGBA[p+1], desc.RGBA[p+2], desc.RGBA[p+3] = c.R, c.G, c.B, c.A
			p += 4
		}
	}
	return nil
}
