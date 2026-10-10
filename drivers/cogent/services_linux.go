//go:build linux && !android && cgo

package cogent

import (
	"context"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/drivers/cogent/internal/platformio"
)

func nativeChooseFile(ctx context.Context, o app.FileOptions) (string, error) {
	args, err := platformio.LinuxFileArgs(o)
	if err != nil {
		return "", err
	}
	if err := nativeServiceContextErr(ctx); err != nil {
		return "", err
	}
	data, err := platformio.Run(ctx, "zenity", args, nil)
	if err != nil {
		return "", platformio.LinuxFileError(err)
	}
	return platformio.DecodeLinuxFile(data)
}
