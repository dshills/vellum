package automation

import (
	"context"
	"errors"
	"net"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestAcceptTransientRetryAndReset(t *testing.T) {
	left, right := net.Pipe()
	defer left.Close()
	defer right.Close()
	transient := []error{
		&net.OpError{Op: "accept", Net: "unix", Err: &os.SyscallError{Syscall: "accept", Err: syscall.EMFILE}},
		syscall.ENFILE, syscall.ECONNABORTED, syscall.EINTR, syscall.EAGAIN,
	}
	var observed []time.Duration
	attempts := 0
	conn, err := nextAccepted(context.Background(), func() (net.Conn, error) {
		attempts++
		if attempts <= len(transient) {
			return nil, transient[attempts-1]
		}
		return left, nil
	}, func(_ context.Context, delay time.Duration) error { observed = append(observed, delay); return nil })
	if err != nil || conn != left || attempts != 6 {
		t.Fatal(conn, err, attempts)
	}
	for i, delay := range observed {
		if delay != (5*time.Millisecond)<<i {
			t.Fatal(observed)
		}
	}
	// Next connection begins a fresh backoff sequence, without sleeping.
	attempts = 0
	observed = nil
	_, err = nextAccepted(context.Background(), func() (net.Conn, error) {
		attempts++
		if attempts == 1 {
			return nil, syscall.EMFILE
		}
		return left, nil
	}, func(_ context.Context, delay time.Duration) error { observed = append(observed, delay); return nil })
	if err != nil || len(observed) != 1 || observed[0] != 5*time.Millisecond {
		t.Fatal(err, observed)
	}
}

func TestAcceptBackoffCapCancellationAndFatalRefusal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var delays []time.Duration
	_, err := nextAccepted(ctx, func() (net.Conn, error) { return nil, syscall.EMFILE }, func(ctx context.Context, delay time.Duration) error {
		delays = append(delays, delay)
		if len(delays) == 12 {
			cancel()
			return ctx.Err()
		}
		return nil
	})
	if !errors.Is(err, context.Canceled) || len(delays) != 12 {
		t.Fatal(err, delays)
	}
	for _, delay := range delays {
		if delay > time.Second || delay < 5*time.Millisecond {
			t.Fatal(delays)
		}
	}
	if delays[len(delays)-1] != time.Second {
		t.Fatal("uncapped backoff", delays)
	}
	for _, fatal := range []error{errors.New("permanent"), syscall.EINVAL, net.ErrClosed} {
		attempts, waits := 0, 0
		_, err = nextAccepted(context.Background(), func() (net.Conn, error) { attempts++; return nil, fatal }, func(context.Context, time.Duration) error { waits++; return nil })
		if !errors.Is(err, fatal) || attempts != 1 || waits != 0 {
			t.Fatal(fatal, err, attempts, waits)
		}
	}
	// The actual timer path exits immediately for an already canceled context;
	// the fixture never sleeps or changes process file-descriptor limits.
	if err = acceptBackoff(ctx, time.Minute); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	attempts := 0
	_, err = nextAccepted(ctx, func() (net.Conn, error) { attempts++; return nil, nil }, acceptBackoff)
	if !errors.Is(err, context.Canceled) || attempts != 0 {
		t.Fatal("accept after cancellation", err, attempts)
	}
}
