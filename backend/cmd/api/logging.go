package main

import (
	"net/http"
)

func (app *application) logRequestError(r *http.Request, message string, err error) {
	app.logger.Error(
		message,
		"request_id", requestIDFromContext(r.Context()),
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)
}
