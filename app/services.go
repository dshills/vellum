package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var ErrUnavailable = errors.New("vellum: platform service unavailable")
var ErrCanceled = errors.New("vellum: dialog canceled")
var ErrServiceBusy = errors.New("vellum: platform service dispatch capacity exhausted")
var ErrServiceTooLarge = errors.New("vellum: platform service data exceeds limit")

type FileMode uint8

const (
	OpenFile FileMode = iota
	SaveFile
)

// FileOptions describes a single-file OS dialog. SaveFile chooses a destination;
// it never writes the file. Extensions omit dots (for example, "txt").
type FileOptions struct {
	Mode                       FileMode
	Title, Directory, Filename string
	Extensions                 []string
}

func (o FileOptions) Validate() error {
	if o.Mode != OpenFile && o.Mode != SaveFile {
		return fmt.Errorf("vellum: invalid file mode")
	}
	for _, value := range append([]string{o.Title, o.Directory, o.Filename}, o.Extensions...) {
		if !utf8.ValidString(value) {
			return fmt.Errorf("vellum: invalid UTF-8 in file options")
		}
		if strings.ContainsRune(value, 0) {
			return fmt.Errorf("vellum: NUL in file options")
		}
	}
	for _, ext := range o.Extensions {
		if ext == "" || strings.ContainsAny(ext, ". /\\") {
			return fmt.Errorf("vellum: invalid extension %q", ext)
		}
	}
	return nil
}

// MenuGroup contains a named native application menu; entries reference command
// IDs in the driver's registry. Menus capture labels/enabled state at SetMenu.
type MenuGroup struct {
	Label    string
	Commands []string
}

// Services is implemented by the native adapter. Methods are worker-safe and
// marshal OS calls onto the native main thread. Call all methods from workers,
// outside Do, Drain, View, and widget callbacks. ChooseFile blocks its caller
// until selection or cancellation; call it in a worker and deliver via PostWindow.
// Services are unavailable until the driver shows a window, and after shutdown.
type Services interface {
	ReadClipboard(context.Context) (string, error)
	WriteClipboard(context.Context, string) error
	ChooseFile(context.Context, FileOptions) (string, error)
	SetMenu(context.Context, []MenuGroup) error
}

func (a *Application) SetServices(s Services) { a.mu.Lock(); a.services = s; a.mu.Unlock() }
func (a *Application) Services() (Services, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return nil, ErrClosed
	}
	if a.services == nil {
		return nil, ErrUnavailable
	}
	return a.services, nil
}
