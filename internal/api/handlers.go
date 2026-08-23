package api

import (
	"encoding/json"
	"net/http"
	"strings"
)

func JSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]any{"code": status, "message": message})
}

type Page struct {
	Items    any `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func Paginate[T any](items []T, page, size int) Page {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	start := (page - 1) * size
	if start > len(items) {
		start = len(items)
	}
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return Page{Items: items[start:end], Total: len(items), Page: page, PageSize: size}
}

func DecodeBody(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func WritePage(w http.ResponseWriter, page Page) {
	write(w, page)
}

func statusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if strings.Contains(err.Error(), "not found") {
		return http.StatusNotFound
	}
	if strings.Contains(err.Error(), "conflict") {
		return http.StatusConflict
	}
	return http.StatusBadRequest
}

func PathID(r *http.Request) int64 {
	p := strings.Trim(r.URL.Path, "/")
	var n int64
	for _, c := range p {
		if c >= '0' && c <= '9' {
			n = n*10 + int64(c-'0')
		}
	}
	return n
}
