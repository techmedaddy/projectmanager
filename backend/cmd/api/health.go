package main

import (
	"net/http"

	"taskflow/backend/internal/response"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.MethodNotAllowed(w)
		return
	}

	response.JSON(w, http.StatusOK, response.HealthResponse{
		Status: "ok",
	})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	response.NotFound(w)
}
