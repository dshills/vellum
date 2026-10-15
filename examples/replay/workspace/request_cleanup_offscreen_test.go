//go:build offscreen

package workspace

import (
	"context"
	"errors"
	"testing"

	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/virtual"
)

// Unit-only capability tests: no Manual window, rendering, native event loop,
// or large-data construction is needed to prove request retirement.
func reviewerRequestFixture(t *testing.T) (*Fixture, *execution.Manual) {
	t.Helper()
	a, err := app.New(app.Config{QueueCapacity: 16, CompletionCapacity: 4})
	if err != nil {
		t.Fatal(err)
	}
	ex, err := execution.NewManual(8)
	if err != nil {
		a.Close()
		t.Fatal(err)
	}
	f, err := New(a, ex, 10_000)
	if err != nil {
		a.Close()
		ex.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		a.Close()
		for _, id := range ex.PendingIDs() {
			_ = ex.Step(context.Background(), id)
		}
		f.Palette.Close()
		ex.Close()
	})
	return f, ex
}

func reviewerOwner(t *testing.T, f *Fixture, fn func() error) {
	t.Helper()
	var inner error
	err := f.Owner.Do(func() { inner = fn() })
	if err != nil || inner != nil {
		t.Fatal(errors.Join(err, inner))
	}
}

func reviewerFinishWorker(t *testing.T, f *Fixture, ex *execution.Manual) {
	t.Helper()
	ids := ex.PendingIDs()
	if len(ids) != 1 {
		t.Fatal("expected exactly one queued worker", ids)
	}
	if err := ex.Step(context.Background(), ids[0]); err != nil {
		t.Fatal(err)
	}
	if err := f.Owner.Drain(); err != nil {
		t.Fatal(err)
	}
	if !f.task.Snapshot().OwnerDone || !f.task.Snapshot().WorkerDone || !f.manager.Pending().Idle() {
		t.Fatal("worker or reserved delivery did not retire", f.task.Snapshot(), f.manager.Pending())
	}
}

func TestReviewerParentContextCancellationRetiresRequest(t *testing.T) {
	for _, kind := range []string{"query", "tree"} {
		t.Run(kind, func(t *testing.T) {
			f, ex := reviewerRequestFixture(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var grid virtual.GridProjectionRequest
			var tree virtual.OutlineRequest
			reviewerOwner(t, f, func() error {
				if kind == "query" {
					if err := f.query(ctx, false, false); err != nil {
						return err
					}
					grid = f.gridRequest
				} else {
					if err := f.changeTree(ctx, virtual.OutlineChange{Kind: virtual.OutlineFilter, Filter: "missing"}); err != nil {
						return err
					}
					tree = f.treeRequest
				}
				return nil
			})
			// Do not call Fixture.cancel: this must exercise parent cancellation.
			cancel()
			reviewerFinishWorker(t, f, ex)
			reviewerOwner(t, f, func() error {
				var err error
				if kind == "query" {
					err = f.Grid.CancelProjection(grid)
				} else {
					err = f.Tree.Cancel(tree)
				}
				if kind == "query" && !errors.Is(err, virtual.ErrGridStale) {
					t.Error("parent cancellation left grid request live", err)
				}
				if kind == "tree" && !errors.Is(err, virtual.ErrTreeStale) {
					t.Error("parent cancellation left tree request live", err)
				}
				state, err := f.Tree.State()
				if err != nil || state.Pending {
					t.Error("tree still pending", state, err)
				}
				return nil
			})
		})
	}
}

func TestReviewerFinishRefusalRetiresOnlyItsExactRequest(t *testing.T) {
	for _, kind := range []string{"query", "tree"} {
		for _, supersede := range []bool{false, true} {
			name := kind + "/current"
			if supersede {
				name = kind + "/superseded"
			}
			t.Run(name, func(t *testing.T) {
				f, ex := reviewerRequestFixture(t)
				refusal := errors.New("reviewer: deliberate finish refusal")
				var oldGrid, newGrid virtual.GridProjectionRequest
				var oldTree, newTree virtual.OutlineRequest
				finished := 0
				reviewerOwner(t, f, func() error {
					var err error
					if kind == "query" {
						oldGrid, err = f.Grid.PrepareProjection()
					} else {
						oldTree, err = f.Tree.Prepare(virtual.OutlineChange{Kind: virtual.OutlineFilter, Filter: "old"})
					}
					if err != nil {
						return err
					}
					err = f.start(context.Background(), kind, func(context.Context) (jobResult, error) { return jobResult{kind: kind}, nil }, func(jobResult) error {
						finished++
						if supersede {
							if kind == "query" {
								newGrid, err = f.Grid.PrepareProjection()
							} else {
								newTree, err = f.Tree.Prepare(virtual.OutlineChange{Kind: virtual.OutlineFilter, Filter: "new"})
							}
							if err != nil {
								return err
							}
						}
						return refusal
					})
					if err != nil {
						return err
					}
					// Exactly the same retained-request publication as query/changeTree.
					f.gridRequest, f.treeRequest = oldGrid, oldTree
					return nil
				})
				reviewerFinishWorker(t, f, ex)
				if finished != 1 || f.failure != refusal.Error() {
					t.Fatal("finish refusal was not delivered", finished, f.failure)
				}
				reviewerOwner(t, f, func() error {
					var err error
					if supersede {
						// Cancel(old) must not revoke a newer public-API capability.
						if kind == "query" {
							err = f.Grid.CancelProjection(newGrid)
						} else {
							err = f.Tree.Cancel(newTree)
						}
						if err != nil {
							t.Error("cleanup revoked newer request", err)
						}
					} else {
						if kind == "query" {
							err = f.Grid.CancelProjection(oldGrid)
						} else {
							err = f.Tree.Cancel(oldTree)
						}
						if kind == "query" && !errors.Is(err, virtual.ErrGridStale) {
							t.Error("finish refusal left grid request live", err)
						}
						if kind == "tree" && !errors.Is(err, virtual.ErrTreeStale) {
							t.Error("finish refusal left tree request live", err)
						}
					}
					return nil
				})
			})
		}
	}
}
