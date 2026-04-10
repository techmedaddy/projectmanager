package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const maxRequestBodyBytes = 1 << 20

// DecodeJSON decodes a single JSON object from the request body and rejects
// unknown fields so request contracts stay explicit.
func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("request body is required")
	}

	decoder := json.NewDecoder(io.LimitReader(r.Body, maxRequestBodyBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	if decoder.More() {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}
