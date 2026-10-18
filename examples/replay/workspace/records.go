package workspace

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dshills/vellum/replay"
	"io"
)

//go:embed records/*.json
var recordFiles embed.FS

func FixtureBytes(rows int) []byte {
	return []byte(fmt.Sprintf("{\"schema\":1,\"rows\":%d,\"columns\":20,\"tree_nodes\":100000,\"values\":\"(r*97+c*13+7)%%10007\",\"parent\":\"(node-1)/100\",\"seed\":1}", rows))
}
func FixtureSHA256(rows int) string {
	h := sha256.Sum256(FixtureBytes(rows))
	return hex.EncodeToString(h[:])
}
func windowWidth(id string) int {
	if id == "inspector" {
		return 500
	}
	return Width
}
func windowHeight(id string) int {
	if id == "inspector" {
		return 320
	}
	return Height
}
func Record(rows int) (replay.Record, error) {
	var record replay.Record
	if rows != 10000 && rows != MaxRows {
		return record, replay.ErrInvalid
	}
	data, err := recordFiles.ReadFile(fmt.Sprintf("records/%d.json", rows))
	if err != nil {
		return record, err
	}
	if len(data) > 8<<20 {
		return record, replay.ErrCapacity
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&record); err != nil {
		return record, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return record, replay.ErrInvalid
	}
	return record, record.Validate(replay.Limits{})
}

// ExpectedCaptures reads the immutable authored goldens. Native integration uses
// their domain assertions while retaining its own actual native geometry/pixels.
// It does not rewrite the record or substitute native observations for goldens.
func ExpectedCaptures(rows int) ([]replay.Checkpoint, error) {
	r, err := Record(rows)
	if err != nil {
		return nil, err
	}
	data, err := recordFiles.ReadFile(fmt.Sprintf("records/%d.golden.json", rows))
	if err != nil {
		return nil, err
	}
	if len(data) > 8<<20 {
		return nil, replay.ErrCapacity
	}
	var out []replay.Checkpoint
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if err = d.Decode(&out); err != nil {
		return nil, err
	}
	if err = d.Decode(new(any)); err != io.EOF {
		return nil, replay.ErrInvalid
	}
	expected := 0
	for _, o := range r.Operations {
		if o.Kind != replay.Capture {
			continue
		}
		if expected >= len(out) {
			return nil, replay.ErrInvalid
		}
		c := out[expected]
		if c.Sequence != o.Sequence || c.Label != o.Label || c.Window != o.Window || len(c.Data) == 0 {
			return nil, replay.ErrInvalid
		}
		expected++
	}
	if expected != len(out) {
		return nil, replay.ErrInvalid
	}
	return out, nil
}
