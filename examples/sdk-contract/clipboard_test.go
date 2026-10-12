package sdkcontract_test

import (
	"context"
	"testing"

	"github.com/dshills/vellum/app"
)

// Compile and consume the optional public contract independently of the driver.
// This in-memory adapter does not establish native clipboard interoperability.
type memoryClipboard struct{ data []app.ClipboardRepresentation }

func (*memoryClipboard) ReadClipboard(context.Context) (string, error) { return "", nil }
func (*memoryClipboard) WriteClipboard(context.Context, string) error  { return nil }
func (*memoryClipboard) ChooseFile(context.Context, app.FileOptions) (string, error) {
	return "", app.ErrUnavailable
}
func (*memoryClipboard) SetMenu(context.Context, []app.MenuGroup) error { return nil }
func (*memoryClipboard) ClipboardCapabilities() app.ClipboardCapabilities {
	return app.ClipboardCapabilities{Text: true, CustomMIME: true}
}
func (m *memoryClipboard) WriteClipboardData(_ context.Context, data []app.ClipboardRepresentation) error {
	if err := app.ValidateClipboardData(data); err != nil {
		return err
	}
	m.data = data
	return nil
}
func (m *memoryClipboard) ReadClipboardData(_ context.Context, formats []string) ([]app.ClipboardRepresentation, error) {
	if err := app.ValidateClipboardFormats(formats); err != nil {
		return nil, err
	}
	var result []app.ClipboardRepresentation
	for _, mime := range formats {
		for _, r := range m.data {
			if r.MIME == mime {
				result = append(result, app.ClipboardRepresentation{MIME: r.MIME, Data: append([]byte(nil), r.Data...)})
			}
		}
	}
	return result, nil
}
func TestPublicClipboardConsumer(t *testing.T) {
	a, err := app.New(app.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	a.SetServices(&memoryClipboard{})
	c, err := a.Clipboard()
	if err != nil || !c.ClipboardCapabilities().CustomMIME {
		t.Fatal(err)
	}
	if err = c.WriteClipboardData(context.Background(), []app.ClipboardRepresentation{{MIME: "text/plain", Data: []byte("Component R17")}, {MIME: "application/vnd.vellum+json", Data: []byte(`{"id":"R17"}`)}}); err != nil {
		t.Fatal(err)
	}
	result, err := c.ReadClipboardData(context.Background(), []string{"image/png", "application/vnd.vellum+json", "text/plain"})
	if err != nil || len(result) != 2 || result[0].MIME != "application/vnd.vellum+json" || string(result[1].Data) != "Component R17" {
		t.Fatal(result, err)
	}
}
