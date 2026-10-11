package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/dshills/vellum/app"
	"github.com/dshills/vellum/execution"
	"github.com/dshills/vellum/tasks"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"testing"
)

var bg = context.Background()

func owner(t *testing.T, p *program, fn func() error) {
	t.Helper()
	if e := owned(p, fn); e != nil {
		t.Fatal(e)
	}
}
func drain(t *testing.T, p *program, batch int) {
	t.Helper()
	for range 1024 {
		if p.owner.Pending().Ordinary+p.owner.Pending().ReadyCompletions == 0 {
			return
		}
		if _, e := p.owner.DrainN(batch); e != nil {
			t.Fatal(e)
		}
	}
	t.Fatal("ownerdrainbound")
}
func portable(t *testing.T, batch int) (*program, *execution.Manual) {
	t.Helper()
	ex, e := execution.NewManual(8)
	if e != nil {
		t.Fatal(e)
	}
	p, e := newProgram(128, true, ex)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { p.owner.Close(); ex.Close() })
	owner(t, p, p.load)
	ids := ex.PendingIDs()
	if len(ids) != 1 {
		t.Fatal(ids)
	}
	if e = ex.Step(bg, ids[0]); e != nil {
		t.Fatal(e)
	}
	drain(t, p, batch)
	for _, w := range p.windows {
		if w.domain.rows != 128 || w.view == nil {
			t.Fatal("load delivery")
		}
		if e = ui.Validate(w.describe()); e != nil {
			t.Fatal(e)
		}
	}
	return p, ex
}
func TestWorkerLatestQueryCancelCreditsAndRefusedSubmit(t *testing.T) {
	for _, batch := range []int{1, 7, 64} {
		t.Run(fmt.Sprintf("batch-%d", batch), func(t *testing.T) {
			p, ex := portable(t, batch)
			w := p.windows[1]
			var first, latest *tasks.Task
			q := virtual.GridQuery{Sort: []virtual.GridSortColumn{{ColumnID: column(1), Kind: virtual.GridSortInteger}}}
			owner(t, p, func() error {
				if e := w.queryWork(q); e != nil {
					return e
				}
				first = w.queryTask
				q.Sort[0].Descending = true
				if e := w.queryWork(q); e != nil {
					return e
				}
				latest = w.queryTask
				before := w.queryRequest
				token := w.queryToken
				if e := w.queryWork(virtual.GridQuery{}); !errors.Is(e, tasks.ErrCapacity) {
					t.Fatal("full admission", e)
				}
				if w.queryRequest != before || w.queryToken != token || w.queryTask != latest {
					t.Fatal("refused capacity retired acceptedquery")
				}
				return nil
			})
			ids := ex.PendingIDs()
			if len(ids) != 2 {
				t.Fatal(ids)
			}
			if e := ex.Step(bg, ids[1]); e != nil {
				t.Fatal(e)
			}
			if w.manager.Pending().Active != 2 {
				t.Fatal("credit released before owner")
			}
			select {
			case <-latest.OwnerDone():
				t.Fatal("owner completed without drain")
			default:
			}
			drain(t, p, batch)
			revision := w.view.Revision()
			if !w.query.Sort[0].Descending || w.manager.Pending().Active != 1 {
				t.Fatal("latestnotaccepted")
			}
			if e := ex.Step(bg, ids[0]); e != nil {
				t.Fatal(e)
			}
			select {
			case <-first.Done():
			default:
				t.Fatal("worker not terminal")
			}
			if w.manager.Pending().Active != 1 {
				t.Fatal("canceledcreditlostbeforeowner")
			}
			drain(t, p, batch)
			if w.view.Revision() != revision || !w.manager.Pending().Idle() {
				t.Fatal("oldresultinstalled or strandedcredit")
			}
			owner(t, p, func() error { return w.queryWork(q) })
			accepted := w.queryTask
			request, token := w.queryRequest, w.queryToken
			ex.Close()
			owner(t, p, func() error {
				if e := w.queryWork(virtual.GridQuery{}); !errors.Is(e, execution.ErrClosed) {
					t.Fatal("refusedsubmit", e)
				}
				if w.queryTask != accepted || w.queryRequest != request || w.queryToken != token {
					t.Fatal("refusedsubmitretiredprior")
				}
				return nil
			})
			ids = ex.PendingIDs()
			if e := ex.Step(bg, ids[0]); e != nil {
				t.Fatal(e)
			}
			drain(t, p, batch)
			if !w.manager.Pending().Idle() {
				t.Fatal("refusedcreditnotrefunded")
			}
		})
	}
}
func TestOwnerDomainCASStaleDraftAndFilteredEdit(t *testing.T) {
	p, ex := portable(t, 7)
	w := p.windows[1]
	owner(t, p, func() error { return w.queryWork(virtual.GridQuery{}) })
	id := ex.PendingIDs()[0]
	oldView := w.view
	newDomain, e := edited(bg, w.domain, key(0), column(1), "314", 1)
	if e != nil {
		t.Fatal(e)
	}
	owner(t, p, func() error { w.domain = newDomain; return nil })
	if e = ex.Step(bg, id); e != nil {
		t.Fatal(e)
	}
	drain(t, p, 7)
	if w.view != oldView || w.queryTask != nil || !w.manager.Pending().Idle() {
		t.Fatal("late result installed or retained task/credit after full-domain replacement")
	}
	// Install valid full input, then filter the edited column to one exact value.
	restored, err := virtual.BuildGridProjection(bg, input(newDomain, true), oldView.Revision()+1, virtual.GridQuery{}, limits())
	if err != nil {
		t.Fatal(err)
	}
	owner(t, p, func() error {
		if e := w.grid.ReplaceSource(restored); e != nil {
			return e
		}
		w.view = restored
		return nil
	})
	owner(t, p, func() error {
		return w.queryWork(virtual.GridQuery{Filters: []virtual.GridFilter{{ColumnID: column(1), Kind: virtual.GridFilterEquals, Text: "314"}}})
	})
	if _, e = ex.StepNext(bg); e != nil {
		t.Fatal(e)
	}
	drain(t, p, 7)
	if w.view.RowCount() != 1 {
		t.Fatal("exactnumericfilter")
	}
	owner(t, p, func() error {
		if e := w.grid.JumpToKey(key(0), column(1)); e != nil {
			return e
		}
		if e := w.begin(); e != nil {
			return e
		}
		if e := w.grid.SetEditText("315"); e != nil {
			return e
		}
		ed, e := w.grid.Edit()
		if e != nil {
			return e
		}
		return w.editWork(ed)
	})
	if _, e = ex.StepNext(bg); e != nil {
		t.Fatal(e)
	}
	drain(t, p, 7)
	ed, e := w.grid.Edit()
	if e != nil || ed.Active || w.view.RowCount() != 0 {
		t.Fatal("edit changingfiltermembership didnotatomicallyclose", ed, e, w.message)
	}
	v, version, _ := w.domain.value(key(0), column(1))
	if v != 315 || version != 3 {
		t.Fatal(v, version, w.message)
	}
	owner(t, p, func() error { return w.queryWork(virtual.GridQuery{}) })
	_, _ = ex.StepNext(bg)
	drain(t, p, 7)
	owner(t, p, func() error {
		if e := w.grid.JumpToKey(key(0), column(1)); e != nil {
			return e
		}
		if e := w.begin(); e != nil {
			return e
		}
		if e := w.grid.SetEditText("42"); e != nil {
			return e
		}
		ed, _ := w.grid.Edit()
		if e := w.editWork(ed); e != nil {
			return e
		}
		return w.grid.SetEditText("43")
	})
	old := w.domain
	_, _ = ex.StepNext(bg)
	drain(t, p, 7)
	ed, _ = w.grid.Edit()
	if w.domain != old || ed.Text != "43" || !ed.Active {
		t.Fatal("staledraftoverwritten")
	}
}
func TestWindowCloseRevokesResultButCanceledWorkerRetainsCredit(t *testing.T) {
	p, ex := portable(t, 1)
	w := p.windows[0]
	old := w.domain
	owner(t, p, func() error {
		if e := w.queryWork(virtual.GridQuery{}); e != nil {
			return e
		}
		return p.owner.CloseWindow(w.id)
	})
	if w.manager.Pending().Active != 1 {
		t.Fatal("closehidpendingworker")
	}
	_, e := ex.StepNext(bg)
	if e != nil {
		t.Fatal(e)
	}
	drain(t, p, 1)
	if !w.manager.Pending().Idle() || w.domain != old {
		t.Fatal("closedresultpublishedorcreditstranded")
	}
	if _, e = w.lifetime(); !errors.Is(e, app.ErrWindowClosed) {
		t.Fatal(e)
	}
	if _, e = p.windows[1].lifetime(); e != nil {
		t.Fatal("independentwindowclosed", e)
	}
}

func TestExternalRefusesPendingSettlementWithoutStrandingSave(t *testing.T) {
	for _, refuseExecutor := range []bool{false, true} {
		t.Run(fmt.Sprintf("executor-refused-%t", refuseExecutor), func(t *testing.T) {
			p, ex := portable(t, 7)
			w := p.windows[1]
			var ack func(error)
			owner(t, p, func() error {
				if err := w.grid.JumpToKey(key(0), column(1)); err != nil {
					return err
				}
				if err := w.begin(); err != nil {
					return err
				}
				if err := w.grid.SetEditText("42"); err != nil {
					return err
				}
				w.settle = func(_ context.Context, cb func(error)) error { ack = cb; return nil }
				return w.save()
			})
			if refuseExecutor {
				ex.Close()
			}
			owner(t, p, func() error {
				token := w.editToken
				if !w.settling || w.settleCancel == nil || ack == nil {
					t.Fatal("settlement not pending")
				}
				if err := w.external(); !errors.Is(err, errBusy) {
					t.Fatal("external should not invalidate pending settlement", err)
				}
				if w.editToken != token || !w.settling || w.settleCancel == nil || w.editTask != nil {
					t.Fatal("refused external stranded settlement")
				}
				ack(nil)
				if w.settling || w.settleCancel != nil {
					t.Fatal("ack did not release pending settlement")
				}
				return nil
			})
			if refuseExecutor {
				if w.editTask != nil || !w.manager.Pending().Idle() || len(ex.PendingIDs()) != 0 {
					t.Fatal("refused worker retained credits")
				}
				ed, err := w.grid.Edit()
				if err != nil || !ed.Active || ed.Text != "42" {
					t.Fatal("refusal discarded draft", ed, err)
				}
				return
			}
			if len(ex.PendingIDs()) != 1 {
				t.Fatal("settled Save was not admitted")
			}
			if _, err := ex.StepNext(bg); err != nil {
				t.Fatal(err)
			}
			drain(t, p, 7)
			ed, err := w.grid.Edit()
			if err != nil || ed.Active || w.editTask != nil || !w.manager.Pending().Idle() {
				t.Fatal("Save did not complete", ed, err, w.message)
			}
		})
	}
}

func TestFilterSettlementReadsAcknowledgedValueAndRefusesStaleWork(t *testing.T) {
	for _, scenario := range []string{"ack", "cancel", "query", "domain", "closed-window", "admission-refused"} {
		t.Run(scenario, func(t *testing.T) {
			p, ex := portable(t, 7)
			w := p.windows[0]
			var ack func(error)
			var admitted context.Context
			owner(t, p, func() error {
				w.settleFilter = func(ctx context.Context, cb func(error)) error {
					admitted, ack = ctx, cb
					if scenario == "admission-refused" {
						return errBusy
					}
					return nil
				}
				err := w.applyFilter()
				if scenario == "admission-refused" {
					if !errors.Is(err, errBusy) || w.settlingFilter || w.filterCancel != nil || admitted.Err() == nil {
						t.Fatal("immediate refusal leaked reservation", err)
					}
					return nil
				}
				return err
			})
			if scenario == "admission-refused" {
				return
			}
			owner(t, p, func() error {
				w.filterText = "7" // owner value published by the field before acknowledgment
				switch scenario {
				case "cancel":
					if err := w.cancelWork(); err != nil {
						return err
					}
				case "query":
					if err := w.queryWork(virtual.GridQuery{}); err != nil {
						return err
					}
				case "domain":
					var err error
					w.domain, err = edited(bg, w.domain, key(0), column(1), "42", 1)
					if err != nil {
						return err
					}
				case "closed-window":
					if err := p.owner.CloseWindow(w.id); err != nil {
						return err
					}
				}
				if scenario == "closed-window" {
					ack(context.Canceled)
				} else {
					ack(nil)
				}
				if w.settlingFilter || w.filterCancel != nil || admitted.Err() == nil {
					t.Fatal("terminal filter reservation not released")
				}
				return nil
			})
			ids := ex.PendingIDs()
			if scenario == "ack" || scenario == "query" {
				if len(ids) != 1 {
					t.Fatal("expected one admitted worker", ids)
				}
				if err := ex.Step(bg, ids[0]); err != nil {
					t.Fatal(err)
				}
				drain(t, p, 7)
				owner(t, p, func() error {
					if scenario == "ack" && (len(w.query.Filters) != 1 || w.query.Filters[0].Text != "7") {
						t.Fatal("filter read ahead of acknowledged value", w.query)
					}
					if scenario == "query" && len(w.query.Filters) != 0 {
						t.Fatal("stale acknowledgment replaced newer query")
					}
					return nil
				})
			} else if len(ids) != 0 {
				t.Fatal("stale/canceled field acknowledgment admitted work", ids)
			}
			if !w.manager.Pending().Idle() {
				t.Fatal("terminal filter work retained credit")
			}
		})
	}
}
