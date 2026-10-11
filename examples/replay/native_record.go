package replayconsumer

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/dshills/vellum/replay"
)

//go:embed native-record.json
var nativeRecordBytes []byte

// NativeRecord loads this application's fixed, bundled fixture. This is not an
// external automation transport. Decode and portable validation finish before
// constructing the owner/driver or applying any recorded work.
func NativeRecord() (replay.Record, error) {
	var r replay.Record
	if len(nativeRecordBytes) > 8<<20 {
		return r, replay.ErrCapacity
	}
	d := json.NewDecoder(bytes.NewReader(nativeRecordBytes))
	d.DisallowUnknownFields()
	if err := d.Decode(&r); err != nil {
		return r, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return r, fmt.Errorf("%w: trailing fixture record data", replay.ErrInvalid)
	}
	return r, r.Validate(replay.Limits{})
}
func NativeRecordSHA256() string {
	h := sha256.Sum256(nativeRecordBytes)
	return hex.EncodeToString(h[:])
}
