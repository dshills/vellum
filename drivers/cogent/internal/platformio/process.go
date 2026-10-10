// Package platformio provides bounded, cancellation-aware OS service transports.
package platformio

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/dshills/vellum/app"
)

var ErrOutputLimit = errors.New("platform service response exceeded limit")

// ProcessError preserves native exit status and diagnostics for adapters that
// distinguish user cancellation from a helper startup failure.
type ProcessError struct {
	Executable string
	Err        error
	Diagnostic string
}

func (e *ProcessError) Error() string {
	return fmt.Sprintf("platform service %s: %v: %s", e.Executable, e.Err, e.Diagnostic)
}
func (e *ProcessError) Unwrap() error { return e.Err }

type limitedBuffer struct {
	buffer   bytes.Buffer
	max      int
	exceeded bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.max-b.buffer.Len() {
		b.exceeded = true
		return 0, ErrOutputLimit
	}
	return b.buffer.Write(p)
}

// Use a named buffer so promoted ReaderFrom/WriteString methods cannot bypass
// Write limits when os/exec uses io.Copy to drain native pipes.
func (b *limitedBuffer) Bytes() []byte     { return b.buffer.Bytes() }
func (b *limitedBuffer) String() string    { return b.buffer.String() }
func (b *truncatingBuffer) String() string { return b.buffer.String() }

// Diagnostic streams are drained after the retained prefix fills. Warnings
// must not close stderr or turn a successful chooser into a transport failure.
type truncatingBuffer struct {
	buffer bytes.Buffer
	max    int
}

func (b *truncatingBuffer) Write(p []byte) (int, error) {
	length := len(p)
	keep := min(length, max(0, b.max-b.buffer.Len()))
	_, _ = b.buffer.Write(p[:keep])
	return length, nil
}

// Run uses argument vectors and stdin exclusively; user data is never shell code.
func Run(ctx context.Context, executable string, args []string, input []byte) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	cmd := exec.CommandContext(ctx, executable, args...)
	// Bound pipe draining if a helper descendant retains inherited stdout/stderr.
	cmd.WaitDelay = time.Second
	cmd.Stdin = bytes.NewReader(input)
	out, diagnostic := &limitedBuffer{max: 1024 * 1024}, &truncatingBuffer{max: 8192}
	cmd.Stdout = out
	cmd.Stderr = diagnostic
	err := cmd.Run()
	if out.exceeded {
		err = errors.Join(err, ErrOutputLimit)
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) || errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s missing", app.ErrUnavailable, executable)
		}
		return nil, &ProcessError{Executable: executable, Err: err, Diagnostic: strings.TrimSpace(diagnostic.String())}
	}
	return out.Bytes(), nil
}

// ValidateFileOptions restricts glob/filter punctuation for every chooser.
// Native Windows and Linux adapters call it before any process can start.
func ValidateFileOptions(o app.FileOptions) error {
	if err := o.Validate(); err != nil {
		return err
	}
	for _, ext := range o.Extensions {
		for _, r := range ext {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
				return fmt.Errorf("unsupported file extension %q", ext)
			}
		}
	}
	return nil
}

// LinuxFileArgs invokes the GTK chooser supplied by Zenity; paths and labels are
// values of explicit options even if they contain leading dashes or metacharacters.
func LinuxFileArgs(o app.FileOptions) ([]string, error) {
	if err := ValidateFileOptions(o); err != nil {
		return nil, err
	}
	args := []string{"--file-selection", "--title=" + o.Title}
	if o.Mode == app.SaveFile {
		args = append(args, "--save", "--confirm-overwrite")
	}
	path := o.Directory
	if o.Filename != "" {
		path = filepath.Join(path, o.Filename)
	} else if path != "" {
		path += string(filepath.Separator)
	}
	if path != "" {
		args = append(args, "--filename="+path)
	}
	if len(o.Extensions) > 0 {
		var patterns []string
		for _, ext := range o.Extensions {
			patterns = append(patterns, "*."+ext)
		}
		args = append(args, "--file-filter=Allowed files | "+strings.Join(patterns, " "))
	}
	return args, nil
}

// EncodePowerShell is only applied to a fixed implementation script; file options
// travel separately as UTF-8 JSON on stdin.
func EncodePowerShell(script string) string {
	words := utf16.Encode([]rune(script))
	data := make([]byte, len(words)*2)
	for i, w := range words {
		binary.LittleEndian.PutUint16(data[i*2:], w)
	}
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeLinuxFile(data []byte) (string, error) {
	value := strings.TrimSuffix(string(data), "\n")
	if !utf8.ValidString(value) || value == "" || strings.ContainsRune(value, 0) {
		return "", fmt.Errorf("invalid GTK file response")
	}
	return value, nil
}

// Recognize only narrow benign GTK warning fixtures. Unknown diagnostics and
// display/startup failures retain their native errors even with cancel status.
func ignorableChooserDiagnostics(diagnostic string) bool {
	for _, line := range strings.Split(diagnostic, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(lower, "cannot open display") || strings.Contains(lower, "unable to init server") || strings.Contains(lower, "fatal") {
			return false
		}
		gtk := strings.Contains(line, "Gtk-Message:") || strings.Contains(line, "Gtk-WARNING")
		benign := strings.Contains(line, "Failed to load module") || strings.Contains(line, "Theme parsing error")
		if !gtk || !benign {
			return false
		}
	}
	return true
}

func LinuxFileError(err error) error {
	var exit *exec.ExitError
	var transport *ProcessError
	if errors.As(err, &exit) && exit.ExitCode() == 1 && errors.As(err, &transport) && ignorableChooserDiagnostics(transport.Diagnostic) {
		return app.ErrCanceled
	}
	return err
}

func DecodeWindowsFile(data []byte) (string, error) {
	var result struct {
		Path     string `json:"path"`
		Canceled bool   `json:"canceled"`
	}
	if !utf8.Valid(data) {
		return "", fmt.Errorf("invalid UTF-8 in file response")
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("invalid file response: %w", err)
	}
	if result.Canceled {
		return "", app.ErrCanceled
	}
	if result.Path == "" || strings.ContainsRune(result.Path, 0) {
		return "", fmt.Errorf("empty or invalid file response")
	}
	return result.Path, nil
}

const WindowsFileScript = `
$ErrorActionPreference = 'Stop'
[Console]::InputEncoding = New-Object System.Text.UTF8Encoding($false)
[Console]::OutputEncoding = New-Object System.Text.UTF8Encoding($false)
Add-Type -AssemblyName System.Windows.Forms
$o = [Console]::In.ReadToEnd() | ConvertFrom-Json
if ($o.Mode -eq 1) { $d = New-Object System.Windows.Forms.SaveFileDialog } else { $d = New-Object System.Windows.Forms.OpenFileDialog; $d.Multiselect = $false; $d.CheckFileExists = $true }
try {
 $d.Title = $o.Title
 $d.InitialDirectory = $o.Directory
 $d.FileName = $o.Filename
 if ($o.Extensions.Count -gt 0) { $patterns = ($o.Extensions | ForEach-Object { '*.' + $_ }) -join ';'; $d.Filter = 'Allowed files|' + $patterns }
 if ($d.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK) { $result = @{path=$d.FileName; canceled=$false} } else { $result = @{path=''; canceled=$true} }
 [Console]::Out.WriteLine(($result | ConvertTo-Json -Compress))
} finally { $d.Dispose() }
`
