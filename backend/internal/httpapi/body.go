package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// errTrailingData means the body has more data after the first JSON value.
var errTrailingData = errors.New("httpapi: trailing data after JSON body")

// decodeBody reads the whole request body, up to maxBodyBytes, and decodes
// exactly one JSON value into v. A body over the limit fails even when the
// extra bytes come after a valid value, for example whitespace padding.
// Data after the value, other than whitespace, also fails.
func decodeBody(w http.ResponseWriter, r *http.Request, v any) error {
	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return errTrailingData
	}
	return nil
}
