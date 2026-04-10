package main

import (
	"log/slog"
	"net/http"
)

func newRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/", notFoundHandler)

	return chain(
		mux,
		requestLoggerMiddleware(logger),
		jsonMiddleware,
	)
}

func chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	wrapped := handler

	for i := len(middlewares) - 1; i >= 0; i-- {
		wrapped = middlewares[i](wrapped)
	}

	return wrapped
}
