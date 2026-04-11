package main

import (
	"net/http"
	"strconv"
	"strings"

	"taskflow/backend/internal/db"
	"taskflow/backend/internal/response"
)

func parsePaginationParams(r *http.Request) (db.Pagination, response.FieldErrors) {
	fields := response.NewFieldErrors()
	query := r.URL.Query()

	pagination := db.Pagination{
		Page:  db.DefaultPage,
		Limit: db.DefaultLimit,
	}

	if rawPage := strings.TrimSpace(query.Get("page")); rawPage != "" {
		page, err := strconv.Atoi(rawPage)
		if err != nil || page < 1 {
			fields.Add("page", "must be a positive integer")
		} else {
			pagination.Page = page
		}
	}

	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit < 1 {
			fields.Add("limit", "must be a positive integer")
		} else {
			pagination.Limit = limit
		}
	}

	return pagination.Normalize(), fields
}

func paginateBounds(total int, pagination db.Pagination) (start int, end int) {
	normalized := pagination.Normalize()
	start = normalized.Offset()
	if start >= total {
		return total, total
	}

	end = start + normalized.Limit
	if end > total {
		end = total
	}

	return start, end
}
