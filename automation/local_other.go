//go:build !darwin

package automation

import (
	"context"
	"net"
)

func localSupported() bool                                { return false }
func dialLocal(context.Context, string) (net.Conn, error) { return nil, ErrUnsupported }
