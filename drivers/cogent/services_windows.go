//go:build windows && cgo

package cogent

import (
	"context"
	"encoding/json"

	"example.com/vellum/app"
	"example.com/vellum/cogent/internal/platformio"
)

func nativeChooseFile(ctx context.Context, o app.FileOptions) (string, error) {
	if err := nativeServiceContextErr(ctx); err != nil {
		return "", err
	}
	// Apply the same narrow glob syntax to WinForms filters. Options themselves
	// remain JSON data; metacharacters never become PowerShell source.
	if err := platformio.ValidateFileOptions(o); err != nil {
		return "", err
	}
	payload, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	args := []string{"-NoLogo", "-NoProfile", "-NonInteractive", "-STA", "-EncodedCommand", platformio.EncodePowerShell(platformio.WindowsFileScript)}
	data, err := platformio.Run(ctx, "powershell.exe", args, payload)
	if err != nil {
		return "", err
	}
	return platformio.DecodeWindowsFile(data)
}
