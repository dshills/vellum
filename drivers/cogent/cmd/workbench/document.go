package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/canvas"
	"github.com/dshills/vellum/ui"
	"github.com/dshills/vellum/virtual"
	"io"
	"math"
	"slices"
	"strconv"
	"unicode/utf8"
)

// Document v1 stores domain state; UI selection, viewport, task status and
// projection revision are transient. Geometry is the first N component IDs in
// order, matching the workbench's sparse geometry editor.
type documentImage struct {
	Version    int
	Parts      []part
	Objects    []canvas.Rect
	References []string
	outline    *virtual.Tree
	captured   *partSource
	prepared   partStore
}

func buildOutline(count, total int) (*virtual.Tree, error) {
	nodes := []virtual.TreeNode{{ID: "project", Label: fmt.Sprintf("Outline · %d of %d components", count, total)}}
	for group := 0; group < (count+999)/1000; group++ {
		id := fmt.Sprintf("group:%d", group)
		nodes = append(nodes, virtual.TreeNode{ID: id, Parent: "project", Label: fmt.Sprintf("Group %d", group+1)})
		for j := 0; j < 1000 && group*1000+j < count; j++ {
			index := group*1000 + j
			nodes = append(nodes, virtual.TreeNode{ID: "component:" + strconv.Itoa(index+1), Parent: id, Label: fmt.Sprintf("R%07d", index+1)})
		}
	}
	tree, err := virtual.NewTree(nodes)
	if err != nil {
		return nil, err
	}
	if err = tree.Toggle("project"); err != nil {
		return nil, err
	}
	return tree, nil
}
func captureDocument(source *partSource, board *canvas.Scene, refs *fileReferences) documentImage {
	return documentImage{Version: 1, captured: source.Snapshot(), Objects: slices.Clone(board.Objects), References: slices.Clone(refs.paths)}
}
func validateDocument(image documentImage) error {
	if image.Version != 1 {
		return fmt.Errorf("unsupported document version %d", image.Version)
	}
	if len(image.Parts) < 100000 || len(image.Parts) > 1000000 {
		return fmt.Errorf("document requires 100,000..1,000,000 components")
	}
	for i, p := range image.Parts {
		if p.Rotation >= 360 || p.Rotation%90 != 0 || p.Package > 1 {
			return fmt.Errorf("invalid component %d", i+1)
		}
	}
	if len(image.Objects) > len(image.Parts) {
		return fmt.Errorf("geometry exceeds component count")
	}
	if _, err := canvas.New(image.Objects); err != nil {
		return err
	}
	for i, o := range image.Objects {
		if o.ID != "component:"+strconv.Itoa(i+1) {
			return fmt.Errorf("geometry must follow component identity order")
		}
	}
	if len(image.References) > 0 {
		if _, err := ui.CopyFileDropPaths(image.References); err != nil {
			return err
		}
	}
	return nil
}

// materializeDocument runs outside the application owner for persistence encoding.
func materializeDocument(image documentImage) documentImage {
	if image.captured != nil {
		image.Parts = image.captured.parts.values()
		image.captured = nil
	}
	return image
}
func encodeDocument(image documentImage) ([]byte, error) {
	image = materializeDocument(image)
	if err := validateDocument(image); err != nil {
		return nil, err
	}
	return json.Marshal(image)
}
func decodeDocument(data []byte, treeRows int) (documentImage, error) {
	var image documentImage
	if !utf8.Valid(data) {
		return image, fmt.Errorf("document must be valid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&image); err != nil {
		return documentImage{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return documentImage{}, fmt.Errorf("document contains trailing data")
	}
	if err := validateDocument(image); err != nil {
		return documentImage{}, err
	}
	var err error
	image.outline, err = buildOutline(min(treeRows, len(image.Parts)), len(image.Parts))
	if err == nil {
		image.prepared = storeParts(image.Parts)
	}
	return image, err
}
func installDocument(source *partSource, table *virtual.Table, board *canvas.Scene, refs *fileReferences, image documentImage) error {
	// Decode has validated and built the projection. No assignments precede the
	// remaining admission checks; its detached buffers transfer to this owner.
	if image.outline == nil || image.prepared.count != len(image.Parts) || image.prepared.count < 100000 {
		return fmt.Errorf("document outline is missing")
	}
	if source.revision == math.MaxUint64 || refs.revision == math.MaxUint64 {
		return fmt.Errorf("component revision exhausted")
	}
	source.parts = image.prepared
	source.revision++
	table.Reset()
	board.Objects = image.Objects
	refs.paths = image.References
	refs.revision++
	return nil
}
