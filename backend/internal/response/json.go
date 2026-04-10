package response

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response body with the provided status code.
func JSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if payload == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(payload)
}
