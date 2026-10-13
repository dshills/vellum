//go:build darwin && cgo && !offscreen && vellum_native_checks

// clipboard-smoke must be launched by scripts/check-clipboard.py. That guardian
// preserves the original raw pasteboard independently of this owned test process.
package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dshills/vellum/app"
	cogent "github.com/dshills/vellum/drivers/cogent"
	"github.com/dshills/vellum/ui"
)

const markerMIME = "application/x-vellum-clipboard-test-owner"
const customMIME = "application/x-vellum-clipboard-test-data"

type config struct {
	Revision                int64
	Nonce                   string
	StatePath, ExpectedPath string
	ReadyPath               string
	Alias                   []string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err) // Errors never contain original clipboard data.
		os.Exit(1)
	}
}
func run() error {
	path := flag.String("fixture-config", "", "protected guardian configuration")
	failure := flag.String("failure-mode", "none", "none, exit-after-owned or hold-after-owned")
	safety := flag.String("safety-state", "", "private-board native guard self-test state file")
	flag.Parse()
	if *safety != "" {
		if err := cogent.CheckNativeClipboardGuardPrivate(*safety); err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]any{"result": "PASS_NATIVE_CLIPBOARD_GUARD_INTERPOSITION", "general_pasteboard_access": false, "foreign_after_admission_refused": true, "foreign_after_empty_refused": true, "wrong_clear_revision_refused": true, "owned_write_accepted": true, "stale_journal_ignored": true})
	}
	file, err := os.Open(*path)
	if err != nil {
		return errors.New("clipboard guardian configuration missing")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.Size() > 8192 || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
		return errors.New("unsafe clipboard configuration")
	}
	var cfg config
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&cfg); err != nil {
		return errors.New("invalid clipboard configuration")
	}
	nonce, err := hex.DecodeString(cfg.Nonce)
	if err != nil || len(nonce) != 32 || len(cfg.Nonce) != 64 || hex.EncodeToString(nonce) != cfg.Nonce || cfg.Revision < 0 || cfg.Revision > 9007199254740991 {
		return errors.New("invalid clipboard check identity")
	}
	dir := filepath.Dir(*path)
	if !filepath.IsAbs(*path) || cfg.StatePath != filepath.Join(dir, "state.json") || cfg.ExpectedPath != filepath.Join(dir, "expected.json") || cfg.ReadyPath != filepath.Join(dir, "child.pid") || len(cfg.Alias) != 0 && len(cfg.Alias) != 2 || *failure != "none" && *failure != "exit-after-owned" && *failure != "hold-after-owned" {
		return errors.New("invalid clipboard check paths/alias")
	}
	a, err := app.New(app.Config{})
	if err != nil {
		return err
	}
	defer a.Close()
	if err = a.Open(app.Window{ID: "clipboard", Title: "Vellum clipboard acceptance", Width: 400, Height: 180, View: func() ui.Node {
		return ui.Node{ID: "root", Kind: ui.Column, Children: []ui.Node{{ID: "label", Kind: ui.Label, Text: "Guarded clipboard acceptance"}}}
	}}); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- check(ctx, a, cfg, *failure); a.Close() }()
	runErr := a.Run(ctx, cogent.Driver{})
	cancel()
	select {
	case err = <-result:
	case <-time.After(5 * time.Second):
		err = errors.New("clipboard worker did not stop")
	}
	return errors.Join(runErr, err)
}
func check(ctx context.Context, a *app.Application, cfg config, failure string) error {
	for {
		s, err := cogent.InspectAccessibility(ctx, "clipboard")
		if err == nil && s.Reachable && len(s.Elements) > 0 && s.Elements[0].Frame[2] > 0 {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(15 * time.Millisecond):
		}
	}
	clipboard, err := a.Clipboard()
	if err != nil {
		return err
	}
	legacy, err := a.Services()
	if err != nil {
		return err
	}
	cap := clipboard.ClipboardCapabilities()
	if !cap.Text || !cap.PNG || !cap.HTML || !cap.CustomMIME {
		return errors.New("native clipboard capabilities missing")
	}
	for {
		ready, e := os.ReadFile(cfg.ReadyPath)
		if e == nil && strings.TrimSpace(string(ready)) == strconv.Itoa(os.Getpid()) {
			break
		}
		select {
		case <-ctx.Done():
			return errors.New("guardian did not durably record child identity")
		case <-time.After(15 * time.Millisecond):
		}
	}
	if _, err = cogent.ConfigureNativeClipboardCheck(ctx, cogent.NativeClipboardCheckConfig{Revision: cfg.Revision, Nonce: cfg.Nonce, StatePath: cfg.StatePath}); err != nil {
		return fmt.Errorf("configure guarded clipboard: %w", err)
	}
	icon := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	icon.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 128})
	icon.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})
	var encoded bytes.Buffer
	if err = png.Encode(&encoded, icon); err != nil {
		return err
	}
	data := []app.ClipboardRepresentation{
		{MIME: "text/plain", Data: []byte("Vellum clipboard check\nA🙂")},
		{MIME: "image/png", Data: encoded.Bytes()},
		{MIME: "text/html", Data: []byte("<b>Vellum clipboard check — A🙂</b>")},
		{MIME: customMIME, Data: []byte{0, 255, 17, 128, 0, 42}},
		{MIME: markerMIME, Data: []byte(cfg.Nonce)},
	}
	if err = clipboard.WriteClipboardData(ctx, data); err != nil {
		return fmt.Errorf("rich clipboard write: %w", err)
	}
	if failure == "exit-after-owned" {
		return errors.New("intentional clipboard fixture failure after owned write")
	}
	if failure == "hold-after-owned" {
		<-ctx.Done()
		return ctx.Err()
	}
	request := []string{customMIME, "application/x-vellum-clipboard-test-missing", "text/html", "image/png", "text/plain", markerMIME}
	expected := []app.ClipboardRepresentation{data[3], data[2], data[1], data[0], data[4]}
	read, err := clipboard.ReadClipboardData(ctx, request)
	if err != nil {
		return err
	}
	if err = equal(read, expected); err != nil {
		return err
	}
	read[0].Data[0] = 99
	read, err = clipboard.ReadClipboardData(ctx, request)
	if err != nil {
		return err
	}
	if err = equal(read, expected); err != nil {
		return errors.New("native clipboard returned borrowed bytes")
	}
	text, err := legacy.ReadClipboard(ctx)
	if err != nil || text != string(data[0].Data) {
		return errors.Join(err, errors.New("legacy text read mismatch"))
	}
	legacyText := "Vellum legacy clipboard check " + cfg.Nonce
	if err = legacy.WriteClipboard(ctx, legacyText); err != nil {
		return fmt.Errorf("guarded legacy write: %w", err)
	}
	text, err = legacy.ReadClipboard(ctx)
	if err != nil || text != legacyText {
		return errors.Join(err, errors.New("legacy text roundtrip mismatch"))
	}
	read, err = clipboard.ReadClipboardData(ctx, []string{"image/png", "text/plain", markerMIME})
	if err != nil {
		return err
	}
	if err = equal(read, []app.ClipboardRepresentation{{MIME: "text/plain", Data: []byte(legacyText)}}); err != nil {
		return err
	}
	if _, err = cogent.SeedNativeClipboardCheckItems(ctx, encoded.Bytes()); err != nil {
		return fmt.Errorf("seed owned items: %w", err)
	}
	read, err = clipboard.ReadClipboardData(ctx, []string{"image/png", markerMIME, "text/plain"})
	if err != nil {
		return err
	}
	if err = equal(read, []app.ClipboardRepresentation{data[4], {MIME: "text/plain", Data: []byte("Vellum first item " + cfg.Nonce)}}); err != nil {
		return errors.New("clipboard merged multiple items")
	}
	if err = clipboard.WriteClipboardData(ctx, data); err != nil {
		return err
	}
	before, err := cogent.InspectNativeClipboardCheck(ctx)
	if err != nil {
		return err
	}
	if len(cfg.Alias) == 2 {
		err = clipboard.WriteClipboardData(ctx, []app.ClipboardRepresentation{{MIME: cfg.Alias[0], Data: []byte{1}}, {MIME: cfg.Alias[1], Data: []byte{2}}})
		if !errors.Is(err, app.ErrUnavailable) {
			return fmt.Errorf("native MIME alias refusal: %w", err)
		}
		after, e := cogent.InspectNativeClipboardCheck(ctx)
		if e != nil {
			return e
		}
		if before != after {
			return errors.New("alias refusal changed clipboard ownership")
		}
	}
	read, err = clipboard.ReadClipboardData(ctx, request)
	if err != nil {
		return err
	}
	if err = equal(read, expected); err != nil {
		return err
	}
	file, err := os.OpenFile(cfg.ExpectedPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return errors.New("expected clipboard output unavailable")
	}
	err = json.NewEncoder(file).Encode(data)
	err = errors.Join(err, file.Sync(), file.Close())
	if err != nil {
		return errors.New("expected clipboard output failed")
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"result": "PASS_GO_CLIPBOARD_TRANSPORT", "request_order": true, "missing_omitted": true, "detached_bytes": true, "legacy_text": true, "first_item_consistency": true, "alias_tested": len(cfg.Alias) == 2, "owned_state": before, "original_content_received": false, "preferences_changed": false})
}
func equal(got, want []app.ClipboardRepresentation) error {
	if len(got) != len(want) {
		return errors.New("clipboard representation count mismatch")
	}
	for i := range want {
		if got[i].MIME != want[i].MIME || !bytes.Equal(got[i].Data, want[i].Data) {
			return errors.New("clipboard representation bytes/order mismatch")
		}
	}
	return nil
}
