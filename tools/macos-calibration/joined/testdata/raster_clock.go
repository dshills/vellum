// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
// Private diagnostic only. No command/image/widget references are retained.
package rasterx

import (
	"image"
	"math"
	"sync"
)

const vellumRasterCapacity = 512

type VellumRasterRecord struct {
	Ordinal                   int
	Target                    image.Rectangle
	Begin, End                float64
	Begun, Complete, Returned bool
}
type VellumRasterSnapshot struct {
	Configured, Overflow, Invalid bool
	Pending                       int
	Records                       []VellumRasterRecord
}

var vellumRasterJournal struct {
	sync.Mutex
	configured, overflow, invalid bool
	clock                         func() float64
	count, pending                int
	rows                          [vellumRasterCapacity]VellumRasterRecord
}

func VellumRasterConfigure(clock func() float64) bool {
	j := &vellumRasterJournal
	j.Lock()
	defer j.Unlock()
	if clock == nil || j.configured || j.pending != 0 {
		return false
	}
	j.configured = true
	j.clock = clock
	return true
}
func vellumRasterBegin(img *image.RGBA) int {
	j := &vellumRasterJournal
	j.Lock()
	if !j.configured {
		j.Unlock()
		return -1
	}
	if j.count == vellumRasterCapacity {
		j.overflow = true
		j.Unlock()
		return -1
	}
	i := j.count
	j.count++
	j.pending++
	j.rows[i].Ordinal = i
	if img != nil {
		j.rows[i].Target = img.Bounds()
	}
	clock := j.clock
	j.Unlock()
	// Native clock is immutable while pending; never call it under the mutex.
	at := clock()
	j.Lock()
	j.rows[i].Begin = at
	j.rows[i].Begun = true
	if math.IsNaN(at) || math.IsInf(at, 0) || at <= 0 {
		j.invalid = true
	}
	j.Unlock()
	return i
}
func vellumRasterEnd(index int, returned bool) {
	if index == -1 {
		return
	}
	j := &vellumRasterJournal
	j.Lock()
	if index < 0 || index >= j.count || !j.rows[index].Begun || j.rows[index].Complete {
		j.invalid = true
		j.Unlock()
		return
	}
	clock := j.clock
	j.Unlock()
	at := clock()
	j.Lock()
	defer j.Unlock()
	// End is private and runs once from Render's defer. The duplicate guard also
	// keeps fault-injection tests from accepting a rewritable completed record.
	if j.rows[index].Complete {
		j.invalid = true
		return
	}
	row := &j.rows[index]
	row.End = at
	row.Complete = true
	row.Returned = returned
	j.pending--
	if math.IsNaN(at) || math.IsInf(at, 0) || at <= row.Begin {
		j.invalid = true
	}
}
func VellumRasterRead() VellumRasterSnapshot {
	j := &vellumRasterJournal
	j.Lock()
	defer j.Unlock()
	return VellumRasterSnapshot{Configured: j.configured, Overflow: j.overflow, Invalid: j.invalid, Pending: j.pending, Records: append([]VellumRasterRecord(nil), j.rows[:j.count]...)}
}

// Reset is only legal after every bracket retires. It drops the clock reference
// and scalar metadata; callers cannot erase an in-flight observation.
func VellumRasterReset() bool {
	j := &vellumRasterJournal
	j.Lock()
	defer j.Unlock()
	if j.pending != 0 {
		return false
	}
	j.configured = false
	j.overflow = false
	j.invalid = false
	j.clock = nil
	j.count = 0
	clear(j.rows[:])
	return true
}
