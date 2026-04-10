package main

import (
	"net/http"
)

func newRouter(app *application) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/auth/register", app.registerHandler)
	mux.HandleFunc("/auth/login", app.loginHandler)
	mux.Handle("/auth/me", app.requireAuth(http.HandlerFunc(app.meHandler)))
	mux.Handle("/projects", app.requireAuth(http.HandlerFunc(app.projectsHandler)))
	mux.Handle("/projects/{id}", app.requireAuth(http.HandlerFunc(app.projectByIDHandler)))
	mux.Handle("/projects/{id}/tasks", app.requireAuth(http.HandlerFunc(app.tasksByProjectHandler)))
	mux.Handle("/tasks/{id}", app.requireAuth(http.HandlerFunc(app.taskByIDHandler)))
	mux.HandleFunc("/", notFoundHandler)

	return chain(
		mux,
		requestIDMiddleware,
		requestLoggerMiddleware(app.logger),
		corsMiddleware,
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
