package virtual

import (
 "errors"
 "testing"
)

func TestReviewGridEditHiddenReferenceRefused(t *testing.T) {
 s := newEditableFixture(100)
 g := mustGrid(t, s, gridColumns(4,100), GridConfig{RowHeight:20,HeaderHeight:20,OverscanRows:1,CellEditing:true}, GridViewport{Width:200,Height:60,ScrollY:40})
 snap := gridSnapshot(t,g)
 var hidden GridCellRef
 for _,cell := range snap.Cells { if cell.Clip.Empty() {hidden=cell.Ref; break} }
 if hidden.token==nil {t.Fatal("fixture lacks overscan")}
 before:=g.state()
 if err:=g.BeginEdit(hidden); !errors.Is(err,ErrGridStale) {t.Fatalf("clipped overscan reference accepted: err=%v edit=%+v",err,g.edit.state)}
 if g.state()!=before {t.Fatal("refused edit mutated state")}
}

func TestReviewGridEditCellVersionHistoryRefused(t *testing.T) {
 for _,mode:=range []string{"rollback","same-version-text","same-version-editability"} {
 t.Run(mode,func(t *testing.T) {
  g,s:=editableFixtureGrid(t,100)
  e:=beginFixtureEdit(t,g)
  remote:=s.candidate(e.RowKey,e.ColumnID,"remote value",3)
  if err:=g.ReplaceSource(remote);err!=nil {t.Fatal(err)}
  before:=g.state()
  next:=remote.candidate(e.RowKey,e.ColumnID,"remote value",3)
  switch mode {
  case "rollback": next.values[GridActive{e.RowKey,e.ColumnID}]=GridEditValue{"remote value",2,true}
  case "same-version-text": next.values[GridActive{e.RowKey,e.ColumnID}]=GridEditValue{"different remote value",3,true}
  case "same-version-editability":next.values[GridActive{e.RowKey,e.ColumnID}]=GridEditValue{"remote value",3,false}
  }
  if err:=g.ReplaceSource(next);!errors.Is(err,ErrGridSource) {t.Fatalf("invalid accessed cell version transition accepted: err=%v edit=%+v",err,g.edit.state)}
  if g.state()!=before ||g.source!=remote {t.Fatal("refusal published source or changed draft")}
 })
 }
}

func TestReviewGridEditBaselineEditabilityVersionRefused(t *testing.T) {
 g,s:=editableFixtureGrid(t,100)
 e:=beginFixtureEdit(t,g)
 before:=g.state()
 locked:=s.candidate(e.RowKey,e.ColumnID,e.Original,e.BaseVersion)
 locked.values[GridActive{e.RowKey,e.ColumnID}]=GridEditValue{e.Original,e.BaseVersion,false}
 if err:=g.ReplaceSource(locked);!errors.Is(err,ErrGridSource) {t.Fatalf("editability changed without cell version: err=%v edit=%+v",err,g.edit.state)}
 if g.state()!=before ||g.source!=s {t.Fatal("refusal published source or changed draft")}
}
