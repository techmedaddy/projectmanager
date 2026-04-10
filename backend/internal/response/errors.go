package response

import "net/http"

// ErrorBody is the standard JSON envelope for non-validation errors.
type ErrorBody struct {
	Error string `json:"error"`
}

// ValidationErrorBody is the standard JSON envelope for field-level validation
// failures.
type ValidationErrorBody struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields"`
}

// HealthResponse is the response body returned by the health endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// Error writes a standard JSON error response.
func Error(w http.ResponseWriter, statusCode int, message string) {
	JSON(w, statusCode, ErrorBody{
		Error: message,
	})
}

// Validation writes the assignment-required validation error shape.
func Validation(w http.ResponseWriter, fields map[string]string) {
	JSON(w, http.StatusBadRequest, ValidationErrorBody{
		Error:  "validation failed",
		Fields: fields,
	})
}

// BadRequest writes a standard 400 error response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

// MethodNotAllowed writes a standard 405 error response.
func MethodNotAllowed(w http.ResponseWriter) {
	Error(w, http.StatusMethodNotAllowed, "method not allowed")
}

// Unauthenticated writes a standard 401 error response.
func Unauthenticated(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "unauthenticated")
}

// Forbidden writes a standard 403 error response.
func Forbidden(w http.ResponseWriter) {
	Error(w, http.StatusForbidden, "forbidden")
}

// NotFound writes a standard 404 error response.
func NotFound(w http.ResponseWriter) {
	Error(w, http.StatusNotFound, "not found")
}

// Internal writes a standard 500 error response.
func Internal(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal server error")
}
