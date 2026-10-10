package cogent

import (
	"image"

	"cogentcore.org/core/styles"
	"cogentcore.org/core/styles/units"
	"cogentcore.org/core/text/rich"
	"example.com/vellum/ui"
)

// applyStyle runs after native default stylers, on a freshly reset native Style.
// An omitted option therefore restores the native default on each update.
func applyStyle(s *styles.Style, n ui.Node, theme ui.Theme) {
	value, err := ui.ResolveStyle(n, theme)
	if err != nil {
		return
	} // validated before mounting or syncing a record
	legacy := n.Kind == ui.Column || n.Kind == ui.Row || n.Kind == ui.Label || n.Kind == ui.Button
	if legacy {
		s.Grow.Set(0, 0)
		s.Min.Set(units.Dp(0), units.Dp(0))
	}
	if n.Kind == ui.Row {
		s.Direction = styles.Row
	} else if n.Kind == ui.Column {
		s.Direction = styles.Column
	}
	if value.Padding.Set {
		p := value.Padding.Value
		s.Padding.Set(units.Dp(p.Top), units.Dp(p.Right), units.Dp(p.Bottom), units.Dp(p.Left))
	}
	if value.Gap.Set {
		s.Gap.Set(units.Dp(value.Gap.Value), units.Dp(value.Gap.Value))
	}
	if value.Align.Set {
		s.Align.Items = nativeAlignment(value.Align.Value)
		s.Align.Content = nativeAlignment(value.Align.Value)
	}
	if value.Justify.Set {
		s.Justify.Content = nativeAlignment(value.Justify.Value)
	}
	if value.MinSize.Set {
		v := value.MinSize.Value
		explicit := theme.Default.MinSize.Set || theme.Kinds[n.Kind].MinSize.Set || n.Style.MinSize.Set
		if legacy || explicit {
			s.Min.Set(units.Dp(v.Width), units.Dp(v.Height))
		} else {
			if n.MinWidth > 0 {
				s.Min.X = units.Dp(n.MinWidth)
			}
			if n.MinHeight > 0 {
				s.Min.Y = units.Dp(n.MinHeight)
			}
		}
	}
	if value.MaxSize.Set {
		v := value.MaxSize.Value
		s.Max.Set(units.Dp(v.Width), units.Dp(v.Height))
	}
	if value.Grow.Set {
		s.Grow.Set(value.Grow.Value, value.Grow.Value)
	}
	if value.Foreground.Set {
		s.Color = image.NewUniform(value.Foreground.Value)
	}
	if value.Background.Set {
		s.Background = image.NewUniform(value.Background.Value)
	}
	if value.Typography.Set {
		p := value.Typography.Value
		s.Font.Size = units.Dp(p.Size)
		if p.Family != "" {
			s.Font.Family = rich.Custom
			s.Font.CustomFont = rich.FontName(p.Family)
		}
		if p.Weight != 0 {
			s.Font.Weight = rich.Weights(p.Weight/100 - 1)
		}
		s.Font.Slant = rich.SlantNormal
		if p.Italic {
			s.Font.Slant = rich.Italic
		}
	}
}
func nativeAlignment(a ui.Alignment) styles.Aligns {
	switch a {
	case ui.AlignCenter:
		return styles.Center
	case ui.AlignEnd:
		return styles.End
	default:
		return styles.Start
	}
}
