package platformio

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/dshills/vellum/app"
)

func TestFileArgsKeepValuesOutOfCode(t *testing.T) {
	o := app.FileOptions{Mode: app.SaveFile, Title: "$(touch marker); ' quote", Directory: "/tmp/a b", Filename: "-report;雪.txt", Extensions: []string{"txt", "Go"}}
	args, err := LinuxFileArgs(o)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"--file-selection", "--title=" + o.Title, "--save", "--confirm-overwrite", "--filename=/tmp/a b/-report;雪.txt", "--file-filter=Allowed files | *.txt *.Go"}
	if fmt.Sprint(args) != fmt.Sprint(want) {
		t.Fatalf("%q", args)
	}
	directoryOnly := o
	directoryOnly.Filename = ""
	directoryArgs, err := LinuxFileArgs(directoryOnly)
	if err != nil || !strings.Contains(strings.Join(directoryArgs, "\n"), "--filename=/tmp/a b/") {
		t.Fatal("Linux directory separator changed with host OS", directoryArgs, err)
	}
	for _, ext := range []string{"*", "x|y", "é", "a?"} {
		o.Extensions = []string{ext}
		if _, err := LinuxFileArgs(o); err == nil {
			t.Fatal(ext)
		}
	}
}
func TestPowerShellEncoding(t *testing.T) {
	script := "é 雪 🧭"
	raw, err := base64.StdEncoding.DecodeString(EncodePowerShell(script))
	if err != nil {
		t.Fatal(err)
	}
	words := make([]uint16, len(raw)/2)
	for i := range words {
		words[i] = binary.LittleEndian.Uint16(raw[i*2:])
	}
	if string(utf16.Decode(words)) != script {
		t.Fatalf("%x", raw)
	}
	if !strings.Contains(WindowsFileScript, "[Console]::In.ReadToEnd()") {
		t.Fatal("options not read from stdin")
	}
}
func TestDecodeWindowsFile(t *testing.T) {
	path := "C:\\a b\\雪 🧭.txt"
	data, _ := json.Marshal(map[string]any{"path": path})
	if got, err := DecodeWindowsFile(data); err != nil || got != path {
		t.Fatalf("%q %v", got, err)
	}
	if _, err := DecodeWindowsFile([]byte(`{"canceled":true}`)); !errors.Is(err, app.ErrCanceled) {
		t.Fatal(err)
	}
	for _, data := range [][]byte{[]byte(`{`), []byte(`{}`), []byte(`{"path":"x\u0000y"}`), {0xff}} {
		if _, err := DecodeWindowsFile(data); err == nil {
			t.Fatalf("%q", data)
		}
	}
}
func TestLimitedBuffer(t *testing.T) {
	b := &limitedBuffer{max: 2}
	b.Write([]byte("ok"))
	if _, err := b.Write([]byte("!")); !errors.Is(err, ErrOutputLimit) {
		t.Fatal(err)
	}
	if b.String() != "ok" {
		t.Fatal(b.String())
	}
}
func TestRunTransport(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("VELLUM_TRANSPORT_FIXTURE", "1")
	args := func(mode string) []string { return []string{"-test.run=TestTransportFixture", "--", mode} }
	out, err := Run(context.Background(), exe, args("echo"), []byte("'$(nothing);雪\n"))
	if err != nil || string(out) != "'$(nothing);雪\n" {
		t.Fatalf("%q %v", out, err)
	}
	_, err = Run(context.Background(), exe, args("fail"), nil)
	if err == nil || !strings.Contains(err.Error(), "fixture diagnostic") {
		t.Fatal(err)
	}
	var transport *ProcessError
	if !errors.As(err, &transport) || transport.Diagnostic != "fixture diagnostic" {
		t.Fatalf("diagnostic lost: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = Run(ctx, exe, args("echo"), nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	ctx, cancel = context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	_, err = Run(ctx, exe, args("wait"), nil)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
	_, err = Run(context.Background(), "vellum-executable-that-does-not-exist", nil, nil)
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
func TestTransportFixture(t *testing.T) {
	if os.Getenv("VELLUM_TRANSPORT_FIXTURE") != "1" {
		return
	}
	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "echo":
		data, _ := io.ReadAll(os.Stdin)
		os.Stdout.Write(data)
		os.Exit(0)
	case "cancel":
		os.Exit(1)
	case "displayfail":
		fmt.Fprint(os.Stderr, "cannot open display")
		os.Exit(1)
	case "fail":
		fmt.Fprint(os.Stderr, "fixture diagnostic")
		os.Exit(7)
	case "cancelwarning":
		fmt.Fprint(os.Stderr, `Gtk-Message: Failed to load module "canberra-gtk-module"`)
		os.Exit(1)
	case "stdoutlimit":
		fmt.Fprint(os.Stdout, strings.Repeat("x", 1024*1024+8192))
		os.Exit(0)
	case "noisy":
		fmt.Fprint(os.Stderr, strings.Repeat("w", 16384))
		fmt.Fprint(os.Stdout, "selected.txt")
		os.Exit(0)
	case "noisyfail":
		fmt.Fprint(os.Stderr, strings.Repeat("w", 16384))
		os.Exit(7)
	case "wait":
		time.Sleep(time.Hour)
		os.Exit(0)
	}
	os.Exit(9)
}

func TestLinuxFileResponseAndCancellation(t *testing.T) {
	path := " /tmp/a\n雪 .txt"
	value, err := DecodeLinuxFile([]byte(path + "\n"))
	if err != nil || value != path {
		t.Fatalf("%q %v", value, err)
	}
	for _, data := range [][]byte{nil, []byte("\n"), []byte("x\x00y\n"), {0xff}} {
		if _, err := DecodeLinuxFile(data); err == nil {
			t.Fatalf("invalid response %q", data)
		}
	}
	exe, _ := os.Executable()
	t.Setenv("VELLUM_TRANSPORT_FIXTURE", "1")
	args := func(mode string) []string { return []string{"-test.run=TestTransportFixture", "--", mode} }
	_, err = Run(context.Background(), exe, args("cancel"), nil)
	if !errors.Is(LinuxFileError(err), app.ErrCanceled) {
		t.Fatal(err)
	}
	_, err = Run(context.Background(), exe, args("displayfail"), nil)
	if errors.Is(LinuxFileError(err), app.ErrCanceled) || err == nil {
		t.Fatalf("display failure accepted as cancellation: %v", err)
	}
	if !errors.Is(LinuxFileError(context.Canceled), context.Canceled) {
		t.Fatal("context cancellation lost")
	}
}

func TestSharedChooserRejectsFilterPunctuation(t *testing.T) {
	for _, ext := range []string{"txt|All files", "txt;exe", "*", "t?xt", "雪"} {
		if err := ValidateFileOptions(app.FileOptions{Extensions: []string{ext}}); err == nil {
			t.Fatalf("accepted filter syntax %q", ext)
		}
	}
	if err := ValidateFileOptions(app.FileOptions{Extensions: []string{"txt", "Go_1-2"}}); err != nil {
		t.Fatal(err)
	}
}

func TestNoisyChooserRetainsBoundedDiagnosticsWithoutAbort(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("VELLUM_TRANSPORT_FIXTURE", "1")
	args := func(mode string) []string { return []string{"-test.run=TestTransportFixture", "--", mode} }
	out, err := Run(context.Background(), exe, args("noisy"), nil)
	if err != nil || string(out) != "selected.txt" {
		t.Fatalf("noisy successful helper: %q %v", out, err)
	}
	_, err = Run(context.Background(), exe, args("noisyfail"), nil)
	var process *ProcessError
	var exit *exec.ExitError
	if !errors.As(err, &process) || !errors.As(err, &exit) {
		t.Fatalf("native failure type: %T", err)
	}
	if exit.ExitCode() != 7 || len(process.Diagnostic) != 8192 {
		t.Fatalf("native exit=%d diagnostic bytes=%d", exit.ExitCode(), len(process.Diagnostic))
	}
}

func TestNativePipeLimitsAndCancelWarnings(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("VELLUM_TRANSPORT_FIXTURE", "1")
	args := func(mode string) []string { return []string{"-test.run=TestTransportFixture", "--", mode} }
	_, err = Run(context.Background(), exe, args("stdoutlimit"), nil)
	if !errors.Is(err, ErrOutputLimit) {
		t.Fatalf("stdout limit was bypassed: %v", err)
	}
	_, err = Run(context.Background(), exe, args("cancelwarning"), nil)
	if !errors.Is(LinuxFileError(err), app.ErrCanceled) {
		t.Fatalf("benign cancel warning: %v", err)
	}
	for _, diag := range []string{"Gtk-WARNING **: cannot open display", "Unable to init server", "Gtk-WARNING **: unexplained failure", "unknown diagnostic"} {
		if ignorableChooserDiagnostics(diag) {
			t.Fatalf("ignored native failure: %q", diag)
		}
	}
}

func TestMissingHelperPathsAreUnavailable(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-helper")
	// Construct a relative path directly: the checkout and temporary directory
	// may be on different Windows volumes, where filepath.Rel cannot work.
	relative := filepath.Join(fmt.Sprintf("vellum-missing-%d", time.Now().UnixNano()), "missing-helper")
	if _, err := os.Stat(relative); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("relative missing-helper fixture must not exist: %v", err)
	}
	for _, executable := range []string{"vellum-executable-that-does-not-exist", missing, relative} {
		t.Run(executable, func(t *testing.T) {
			_, err := Run(context.Background(), executable, nil, nil)
			if !errors.Is(err, app.ErrUnavailable) {
				t.Fatalf("missing helper %q: %v", executable, err)
			}
		})
	}
	// A path that exists but cannot execute is a native startup failure, not a
	// missing dependency callers should silently replace with a fallback.
	directory := t.TempDir()
	_, err := Run(context.Background(), directory, nil, nil)
	var process *ProcessError
	if errors.Is(err, app.ErrUnavailable) || !errors.As(err, &process) {
		t.Fatalf("non-executable existing helper: %v", err)
	}
}
