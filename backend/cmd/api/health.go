package main

import (
	"net/http"

	"taskflow/backend/internal/response"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.JSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "method not allowed",
		})
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusNotFound, map[string]string{
		"error": "not found",
	})
}
