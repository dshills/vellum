//go:build darwin

package automation

import (
	"context"
	"net"
)

func localSupported() bool { return true }
func dialLocal(ctx context.Context, path string) (net.Conn, error) {
	return (&net.Dialer{}).DialContext(ctx, "unix", path)
}
