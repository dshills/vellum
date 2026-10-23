//go:build darwin && !offscreen
package main
/*
#include <stdlib.h>
#include "control.h"
*/
import "C"
import (
 "encoding/json"
 "errors"
 "unsafe"
 "time"
)
func phaseTarget(sequence int) (float64,error) {target:=float64(C.pc_target(C.int(sequence)));if target<=0{return 0,errors.New("phase anchor deadline/order")};return target,nil}
func phaseWait(target float64) error {if C.pc_wait(C.double(target))!=1{return errors.New("phase wake outside2ms")};return nil}
func phaseConfigure() error {if C.pc_configure()!=1{return errors.New("pending control configuration")};return nil}
func phaseSnapshot() (map[string]any,error) {
 deadline:=time.Now().Add(2*time.Second)
 for {
  raw:=C.pc_snapshot();if raw==nil{return nil,errors.New("pending control snapshot bound")}
  var out map[string]any;err:=json.Unmarshal([]byte(C.GoString(raw)),&out);C.free(unsafe.Pointer(raw))
  if err!=nil{return nil,err};if out["invalid"]!=false||out["overflow"]!=false{return out,errors.New("pending recorder invalid/overflow")}
  if out["pending"]==float64(0)&&out["acquiring"]==float64(0){return out,nil}
  if time.Now().After(deadline){return out,errors.New("pending control callbacks did not drain")};time.Sleep(time.Millisecond)
 }
}
