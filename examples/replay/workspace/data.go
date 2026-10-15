// Package workspace is an independent professional-workspace consumer using
// public Vellum APIs. Immutable stored data and indexes are built on workers.
package workspace

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/dshills/vellum/virtual"
)

const Columns = 20
const TreeNodes = 100_000
const MaxRows = 1_000_000

type row struct{ Values [Columns]uint32 }
type source struct {
	rows   []row
	digest string
}

func rowKey(i int) string          { return fmt.Sprintf("row:%07d", i) }
func columnID(i int) string        { return fmt.Sprintf("c%02d", i) }
func (s *source) Identity() string { return "workspace:stored-records" }
func (s *source) Revision() uint64 {
	if len(s.rows) == 0 {
		return 1
	}
	return 2
}
func (s *source) RowCount() int { return len(s.rows) }
func (s *source) RowKey(i int) (string, error) {
	if i < 0 || i >= len(s.rows) {
		return "", virtual.ErrGridRange
	}
	return rowKey(i), nil
}
func (s *source) IndexOf(k string) (int, bool, error) {
	if len(k) != 11 || !strings.HasPrefix(k, "row:") {
		return 0, false, nil
	}
	i, e := strconv.Atoi(k[4:])
	return i, e == nil && i >= 0 && i < len(s.rows) && rowKey(i) == k, nil
}
func (s *source) Cell(k, c string) (string, error) {
	i, ok, e := s.IndexOf(k)
	if e != nil {
		return "", e
	}
	if !ok || len(c) != 3 || c[0] != 'c' {
		return "", virtual.ErrGridRange
	}
	j, e := strconv.Atoi(c[1:])
	if e != nil || j < 0 || j >= Columns || columnID(j) != c {
		return "", virtual.ErrGridRange
	}
	return strconv.FormatUint(uint64(s.rows[i].Values[j]), 10), nil
}
func (s *source) SelectionSnapshot() (virtual.GridRowIndex, error) { return s, nil }
func generatedValue(r, c int) uint32                               { return uint32((uint64(r)*97 + uint64(c)*13 + 7) % 10007) }
func buildData(ctx context.Context, count int) (*source, *virtual.TreeData, *virtual.TreeIndex, error) {
	if count != 10_000 && count != MaxRows {
		return nil, nil, nil, fmt.Errorf("workspace: unsupported row count")
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, nil, err
	}
	data := &source{rows: make([]row, count)}
	hash := sha256.New()
	var bytes [Columns * 4]byte
	for i := range data.rows {
		if i%256 == 0 {
			if e := ctx.Err(); e != nil {
				return nil, nil, nil, e
			}
		}
		for j := range data.rows[i].Values {
			v := generatedValue(i, j)
			data.rows[i].Values[j] = v
			binary.LittleEndian.PutUint32(bytes[j*4:], v)
		}
		hash.Write(bytes[:])
	}
	data.digest = hex.EncodeToString(hash.Sum(nil))
	nodes := make([]virtual.TreeNode, TreeNodes)
	for i := range nodes {
		if i%256 == 0 {
			if e := ctx.Err(); e != nil {
				return nil, nil, nil, e
			}
		}
		parent := ""
		if i > 0 {
			parent = fmt.Sprintf("node:%06d", (i-1)/100)
		}
		nodes[i] = virtual.TreeNode{ID: fmt.Sprintf("node:%06d", i), Parent: parent, Label: fmt.Sprintf("Component %06d", i)}
	}
	tree, e := virtual.NewTreeData(ctx, "workspace:stored-tree", 2, nodes, virtual.TreeLimits{MaxNodes: TreeNodes, MaxStorageBytes: 64 << 20, MaxWorkOperations: 64_000_000})
	if e != nil {
		return nil, nil, nil, e
	}
	index, e := virtual.BuildTreeIndex(ctx, tree, virtual.TreeIndexOptions{ExpandAll: true})
	if e != nil {
		return nil, nil, nil, e
	}
	return data, tree, index, nil
}
func projectionLimits() virtual.GridProjectionLimits {
	return virtual.GridProjectionLimits{MaxRows: MaxRows, MaxIndexBytes: 32 << 20, MaxWorkingBytes: 256 << 20}
}
