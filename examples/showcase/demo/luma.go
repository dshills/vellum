package demo

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

type Artwork struct {
	ID, Name, Collection, Description string
	Index                             int
	Favorite                          bool
}
type Luma struct {
	selection []string
	owner
	items                          []Artwork
	selected, query, filter, order string
	look                           int
	collection                     ui.CollectionView
}

func NewLuma() (*Luma, error) {
	m := &Luma{selected: "art-0", selection: []string{"art-0"}, filter: "all", order: "name", collection: ui.CollectionView{ID: "luma-filmstrip", Label: "Artwork filmstrip", ItemWidth: 112}}
	names := []string{"Canyon dawn", "Alpine still", "Tidal blue", "Copper light", "Green horizon", "Moonlit ridge"}
	collections := []string{"earth", "water", "water", "earth", "earth", "water"}
	descriptions := []string{"Soft ridgelines. A little desert warmth.", "An alpine study in quiet blue.", "Where the land meets the tide.", "The last light on a copper horizon.", "Layered greens, from forest to mist.", "A mountain silhouette after dusk."}
	for i, n := range names {
		m.items = append(m.items, Artwork{fmt.Sprintf("art-%d", i), n, collections[i], descriptions[i], i, i%2 == 0})
	}
	_ = landscapeAssets()
	m.collection.Change = func(e ui.ListSelectionEvent) {
		m.selection = e.Apply(m.selection)
		m.selected = ""
		if slices.Contains(m.selection, e.TargetID) {
			m.selected = e.TargetID
		} else if len(m.selection) > 0 {
			m.selected = m.selection[0]
		}
		m.changed()
	}
	m.collection.Activate = func(e ui.CollectionEvent) { _ = m.Select(e.ItemID) }
	m.register("luma.favorite", "Toggle favorite", "Meta+D", func() error {
		if m.selected == "" {
			return errors.New("select artwork")
		}
		return nil
	}, m.ToggleFavorite)
	m.register("luma.original", "Original colors", "Meta+0", nil, func() error { return m.SetLook(0) })
	return m, nil
}
func (m *Luma) Bind(a *app.Application, r app.WindowRef) { m.bind(a, r) }
func (m *Luma) visible() []Artwork {
	out := []Artwork{}
	for _, a := range m.items {
		if m.filter == "favorites" && !a.Favorite {
			continue
		}
		if (m.filter == "earth" || m.filter == "water") && a.Collection != m.filter {
			continue
		}
		if !strings.Contains(strings.ToLower(a.Name+" "+a.Description), strings.ToLower(m.query)) {
			continue
		}
		out = append(out, a)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if m.order == "color" {
			return out[i].Index < out[j].Index
		}
		return out[i].Name < out[j].Name
	})
	return out
}
func (m *Luma) reconcileSelection() {
	visible := m.visible()
	accepted := []string{}
	for _, a := range visible {
		if slices.Contains(m.selection, a.ID) {
			accepted = append(accepted, a.ID)
		}
	}
	m.selection = accepted
	if !slices.Contains(accepted, m.selected) {
		m.selected = ""
		if len(accepted) > 0 {
			m.selected = accepted[0]
		}
	}
	if m.selected == "" && len(visible) > 0 {
		m.selected = visible[0].ID
		m.selection = []string{m.selected}
	}
}
func (m *Luma) Select(id string) error {
	for _, a := range m.visible() {
		if a.ID == id {
			m.selected = id
			m.selection = []string{id}
			m.changed()
			return nil
		}
	}
	return errors.New("unknown visible artwork")
}
func (m *Luma) SetSearch(q string) error {
	if len(q) > 256 || !utf8.ValidString(q) || strings.ContainsRune(q, 0) {
		return errors.New("search exceeds256 bytes")
	}
	m.query = q
	m.reconcileSelection()
	m.changed()
	return nil
}
func (m *Luma) SetFilter(f string) error {
	if f != "all" && f != "favorites" && f != "earth" && f != "water" {
		return errors.New("unknown collection")
	}
	m.filter = f
	m.reconcileSelection()
	m.changed()
	return nil
}
func (m *Luma) SetLook(look int) error {
	if look < 0 || look > 2 {
		return errors.New("unknown look")
	}
	m.look = look
	m.changed()
	return nil
}
func (m *Luma) ToggleFavorite() error {
	for i := range m.items {
		if m.items[i].ID == m.selected {
			m.items[i].Favorite = !m.items[i].Favorite
			m.reconcileSelection()
			m.changed()
			return nil
		}
	}
	return errors.New("select artwork")
}
func (m *Luma) selectedArtwork() (Artwork, bool) {
	for _, a := range m.items {
		if a.ID == m.selected {
			return a, true
		}
	}
	return Artwork{}, false
}
func (m *Luma) Toolbar() []cogent.TitlebarToolbarItem {
	return []cogent.TitlebarToolbarItem{{ID: "favorite", Label: "Toggle favorite", Kind: ui.Button, PNG: toolbarIcon(rgb(243, 173, 51), "add"), Command: "luma.favorite"}, {ID: "space", Kind: ui.ToolbarFlexibleSpace}, {ID: "original", Label: "Original colors", Kind: ui.Button, PNG: toolbarIcon(rgb(100, 133, 223), "reset"), Command: "luma.original"}, {ID: "all", Label: "All artwork", Kind: ui.Button, PNG: toolbarIcon(rgb(44, 163, 177), "folder"), Activate: func() { _ = m.SetFilter("all"); _ = m.SetSearch("") }}}
}

var histograms = sync.OnceValue(func() [6][3]*canvas.TiledImage {
	var result [6][3]*canvas.TiledImage
	for index, looks := range landscapeAssets() {
		for look, img := range looks {
			var bins [3][64]float64
			for y := 0; y < 600; y += 8 {
				for x := 0; x < 960; x += 8 {
					c := img.NRGBAAt(x, y)
					bins[0][int(c.R)/4]++
					bins[1][int(c.G)/4]++
					bins[2][int(c.B)/4]++
				}
			}
			p := image.NewNRGBA(image.Rect(0, 0, 220, 76))
			maximum := 1.
			for _, channel := range bins {
				for _, n := range channel {
					maximum = math.Max(maximum, n)
				}
			}
			colors := []color.NRGBA{rgb(219, 116, 110), rgb(112, 166, 128), rgb(93, 138, 195)}
			for c, channel := range bins {
				for x := range 220 {
					h := int(math.Sqrt(channel[min(63, x*64/220)]/maximum) * 68)
					for y := 76 - h; y < 76; y++ {
						p.SetNRGBA(x, y, colors[c])
					}
				}
			}
			result[index][look] = tiled(p)
		}
	}
	return result
})

func histogram(index, look int) *canvas.TiledImage { return histograms()[index][look] }
func (m *Luma) View() ui.Node {
	search := ui.Node{ID: "luma-search", Kind: ui.TextField, Text: m.query, Placeholder: "Search artwork", Accessibility: ui.Accessibility{Name: "Search artwork"}, Change: func(e ui.ChangeEvent) { _ = m.SetSearch(e.Text) }, Style: ui.Style{MinSize: ui.Some(ui.Size{Width: 100}), GrowX: ui.Some(float32(1))}}
	left := pad(col("luma-sidebar", label("luma-brand", "Luma", 26, 700, ink), label("luma-tagline", "A library of light", 12, 400, muted), search, section("luma-library", "LIBRARY", nav("luma-all", "All artwork", m.filter == "all", rgb(64, 142, 222), func() { _ = m.SetFilter("all") }), nav("luma-favorites", "Favorites", m.filter == "favorites", rgb(241, 172, 53), func() { _ = m.SetFilter("favorites") })), section("luma-collections", "COLLECTIONS", nav("luma-earth", "Earth tones", m.filter == "earth", rgb(211, 131, 95), func() { _ = m.SetFilter("earth") }), nav("luma-water", "Water & sky", m.filter == "water", rgb(80, 174, 175), func() { _ = m.SetFilter("water") })), spacer("luma-side-space"), pill("luma-originals", "Original artwork", rgb(93, 119, 109), rgb(229, 239, 232)), label("luma-local", "Generated locally. Yours to explore.", 10, 400, muted)), 16)
	left = width(left, 208)
	left.Style.GrowY = ui.Some(float32(1))
	left.Style.Background = ui.Some(rgb(247, 248, 250))
	left.Style.Gap = ui.Some(float32(23))
	selected, ok := m.selectedArtwork()
	name := "No matching artwork"
	description := "Try another collection or search."
	if ok {
		name = selected.Name
		description = selected.Description
	}
	title := row("luma-heading", col("luma-headings", label("luma-kicker", "THE LANDSCAPE COLLECTION", 10, 600, muted), label("luma-title", name, 24, 600, ink)), spacer("luma-head-space"), ui.Node{ID: "luma-sort", Kind: ui.Selector, SelectedID: m.order, Options: []ui.Option{{ID: "name", Label: "Name"}, {ID: "color", Label: "Color story"}}, Change: func(e ui.ChangeEvent) {
		if e.SelectedID == "name" || e.SelectedID == "color" {
			m.order = e.SelectedID
			m.changed()
		}
	}, Accessibility: ui.Accessibility{Name: "Sort artwork"}})
	preview := col("luma-preview")
	preview.Style.Background = ui.Some(rgb(30, 34, 40))
	preview.Style.CornerRadius = ui.Some(float32(10))
	preview.Style.Padding = ui.Some(ui.UniformInsets(16))
	preview.Style.GrowY = ui.Some(float32(1))
	preview.Style.MinSize = ui.Some(ui.Size{Height: 240})
	if ok {
		preview.Children = append(preview.Children, imageNode("luma-main-image", selected.Name, landscapeAssets()[selected.Index][m.look], 0, 0))
	} else {
		preview.Children = append(preview.Children, label("luma-no-results", description, 16, 400, white))
	}
	film := m.visible()
	m.collection.Items = nil
	m.collection.SelectedIDs = slices.Clone(m.selection)
	for _, a := range film {
		m.collection.Items = append(m.collection.Items, ui.CollectionItem{ID: a.ID, Label: a.Name, Thumbnail: landscapeAssets()[a.Index][0]})
	}
	m.collection.Style = ui.Style{GrowX: ui.Some(float32(1)), GrowY: ui.Some(float32(0)), MinSize: ui.Some(ui.Size{Height: 132}), MaxSize: ui.Some(ui.Size{Height: 132}), Background: ui.Some(white), Border: ui.Some(ui.Border{})}
	filmNode, e := m.collection.View()
	if e != nil {
		panic(e)
	}
	main := pad(col("luma-main", title, preview, label("luma-description", description, 13, 400, muted), row("luma-strip-title", label("luma-contact-title", "In this collection", 12, 600, ink), spacer("luma-count-space"), label("luma-count", fmt.Sprintf("%d images", len(film)), 11, 400, muted)), filmNode, rule("luma-bottom-rule"), row("luma-footer", label("luma-memory", "In memory · no personal files", 11, 400, muted), spacer("luma-foot-space"), label("luma-size", "960 × 600  /  RGB", 11, 400, muted))), 22)
	main.Style.GrowY = ui.Some(float32(1))
	main.Style.Gap = ui.Some(float32(14))
	main.Style.Background = ui.Some(white)
	lookIDs := []string{"original", "warm", "mono"}
	lookNames := []string{"Original", "Warm", "Mono"}
	options := []ui.Option{}
	for i, id := range lookIDs {
		options = append(options, ui.Option{ID: id, Label: lookNames[i]})
	}
	looks := ui.Node{ID: "luma-looks", Kind: ui.SegmentedControl, SelectedID: lookIDs[m.look], Options: options, Accessibility: ui.Accessibility{Name: "Image color look"}, Change: func(e ui.ChangeEvent) {
		for i, id := range lookIDs {
			if id == e.SelectedID {
				_ = m.SetLook(i)
			}
		}
	}, Style: ui.Style{GrowX: ui.Some(float32(1))}}
	fav := button("luma-favorite", "Add to favorites", func() { _ = m.ToggleFavorite() })
	fav.Disabled = !ok
	if selected.Favorite {
		fav.Text = "Remove favorite"
	}
	fav.IconImage = icon(rgb(240, 174, 51), "add")
	inspector := pad(col("luma-inspector", label("luma-inspector-title", "Image details", 15, 600, ink), rule("luma-inspector-rule"), section("luma-metadata", "ORIGINAL", label("luma-meta-name", name, 17, 600, ink), label("luma-dimensions", "960 × 600 pixels", 12, 400, muted), label("luma-format", "Generated RGB landscape", 12, 400, muted)), section("luma-colors", "COLOR PROFILE", imageNode("luma-histogram", "Sampled RGB histogram", histogram(selected.Index, m.look), 0, 70), label("luma-profile", "sRGB  ·  "+lookNames[m.look], 11, 400, muted)), section("luma-look-section", "LOOK", looks, label("luma-look-note", "Nondestructive color variants.", 11, 400, muted)), fav, spacer("luma-inspector-space"), pill("luma-asset-tag", "PROCEDURAL ORIGINAL", rgb(139, 103, 69), rgb(247, 237, 221)), label("luma-asset-credit", "Created with Go. Displayed by Vellum.", 10, 400, muted)), 18)
	if !ok {
		inspector.Children = []ui.Node{label("luma-inspector-empty", "No image selected", 15, 600, muted)}
	}
	inspector = width(inspector, 226)
	inspector.Style.GrowY = ui.Some(float32(1))
	inspector.Style.Background = ui.Some(rgb(251, 251, 253))
	inspector.Style.Gap = ui.Some(float32(23))
	return ui.Node{ID: "luma-root", Kind: ui.Row, Grow: true, Children: []ui.Node{left, main, inspector}, Style: ui.Style{Gap: ui.Some(float32(1)), Background: ui.Some(line)}}

}
