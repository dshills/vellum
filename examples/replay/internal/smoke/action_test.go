package smoke

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAdmissionIsNotRepeatedWhileOwnerCompletes(t *testing.T) {
	queries, calls := 0, 0
	r, err := Action(context.Background(), "placement", Hooks{
		Ready:    func(context.Context) (string, error) { return "owned-target", nil },
		Perform:  func(context.Context, string) (bool, error) { calls++; return true, nil },
		Accepted: func(context.Context) (bool, error) { queries++; return queries >= 4, nil },
	})
	if err != nil || calls != 1 || r.Attempts != 1 || r.Admissions != 1 || !r.OwnerCompleted {
		t.Fatalf("%+v %v calls=%d", r, err, calls)
	}
}

func TestActionBoundsContextWithoutExtendingCallerDeadline(t *testing.T) {
	for _, name := range []string{"implicit", "earlier-caller"} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			if name == "earlier-caller" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Second)
				defer cancel()
			}
			parentDeadline, parentBounded := ctx.Deadline()
			queries := 0
			_, err := Action(ctx, "bounded", Hooks{
				Ready:   func(context.Context) (string, error) { return "target", nil },
				Perform: func(context.Context, string) (bool, error) { return true, nil },
				Accepted: func(inner context.Context) (bool, error) {
					deadline, ok := inner.Deadline()
					if !ok || time.Until(deadline) > 30*time.Second {
						t.Fatal("helper supplied no cooperative bound")
					}
					if parentBounded && !deadline.Equal(parentDeadline) {
						t.Fatal("helper changed earlier caller deadline")
					}
					queries++
					return queries > 1, nil
				},
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestIncompleteOwnerTimesOutWithoutRepeatingAdmission(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	calls := 0
	r, err := Action(ctx, "owner-timeout", Hooks{
		Ready:    func(context.Context) (string, error) { return "target", nil },
		Perform:  func(context.Context, string) (bool, error) { calls++; return true, nil },
		Accepted: func(context.Context) (bool, error) { return false, nil },
	})
	if !errors.Is(err, context.DeadlineExceeded) || calls != 1 || r.Admissions != 1 || r.OwnerCompleted {
		t.Fatalf("%+v %v calls=%d", r, err, calls)
	}
}

func TestReadinessDeadlineRefusesBeforeAdmission(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	ready, calls := 0, 0
	r, err := Action(ctx, "readiness-timeout", Hooks{
		Ready:    func(context.Context) (string, error) { ready++; return "", nil },
		Perform:  func(context.Context, string) (bool, error) { calls++; return true, nil },
		Accepted: func(context.Context) (bool, error) { return false, nil },
	})
	if !errors.Is(err, context.DeadlineExceeded) || ready == 0 || calls != 0 || r.Attempts != 0 || r.OwnerCompleted {
		t.Fatalf("%+v %v readiness=%d calls=%d", r, err, ready, calls)
	}
}

func TestMissingInputsRefuseBeforeProposal(t *testing.T) {
	for _, name := range []string{"context", "ready", "perform", "accepted"} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			calls := 0
			h := Hooks{Ready: func(context.Context) (string, error) { return "target", nil }, Perform: func(context.Context, string) (bool, error) { calls++; return true, nil }, Accepted: func(context.Context) (bool, error) { return false, nil }}
			switch name {
			case "context":
				ctx = nil
			case "ready":
				h.Ready = nil
			case "perform":
				h.Perform = nil
			case "accepted":
				h.Accepted = nil
			}
			r, err := Action(ctx, "invalid-input", h)
			if err == nil || calls != 0 || r.Attempts != 0 {
				t.Fatalf("%+v %v calls=%d", r, err, calls)
			}
		})
	}
}

func TestCanceledAfterAdmissionDoesNotRetry(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	r, err := Action(ctx, "uncertain", Hooks{
		Ready:    func(context.Context) (string, error) { return "target", nil },
		Perform:  func(context.Context, string) (bool, error) { calls++; cancel(); return true, nil },
		Accepted: func(context.Context) (bool, error) { return false, nil },
	})
	if !errors.Is(err, context.Canceled) || calls != 1 || r.Admissions != 1 || r.OwnerCompleted {
		t.Fatalf("%+v %v calls=%d", r, err, calls)
	}
}

func TestRefusalAndTransportErrorDoNotRetry(t *testing.T) {
	for _, transportError := range []bool{false, true} {
		t.Run(map[bool]string{false: "refused", true: "uncertain-error"}[transportError], func(t *testing.T) {
			calls := 0
			r, err := Action(context.Background(), "refusal", Hooks{
				Ready: func(context.Context) (string, error) { return "target", nil },
				Perform: func(context.Context, string) (bool, error) {
					calls++
					if transportError {
						return true, errors.New("lost reply")
					}
					return false, nil
				},
				Accepted: func(context.Context) (bool, error) { return false, nil },
			})
			if err == nil || calls != 1 || r.Attempts != 1 || r.OwnerCompleted {
				t.Fatalf("%+v %v calls=%d", r, err, calls)
			}
		})
	}
}

func TestPreexistingSuccessCannotPassWithoutAction(t *testing.T) {
	calls := 0
	r, err := Action(context.Background(), "stale", Hooks{
		Ready:    func(context.Context) (string, error) { calls++; return "target", nil },
		Perform:  func(context.Context, string) (bool, error) { calls++; return true, nil },
		Accepted: func(context.Context) (bool, error) { return true, nil },
	})
	if err == nil || calls != 0 || r.Attempts != 0 || r.OwnerCompleted {
		t.Fatalf("%+v %v calls=%d", r, err, calls)
	}
}

func TestCanceledReadinessNeverSendsProposal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	r, err := Action(ctx, "not-ready", Hooks{
		Ready:    func(context.Context) (string, error) { cancel(); return "", nil },
		Perform:  func(context.Context, string) (bool, error) { calls++; return true, nil },
		Accepted: func(context.Context) (bool, error) { return false, nil },
	})
	if !errors.Is(err, context.Canceled) || calls != 0 || r.Attempts != 0 {
		t.Fatalf("%+v %v calls=%d", r, err, calls)
	}
}
