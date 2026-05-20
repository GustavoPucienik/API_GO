package response

import (
	"encoding/json"
	"net/http"
)

type success[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
}

type successMsg struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type errBody struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type paginated[T any] struct {
	Success bool `json:"success"`
	Data    T    `json:"data"`
	Meta    Meta `json:"meta"`
}

type Meta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	TotalPages int   `json:"totalPages"`
	Limit      int   `json:"limit"`
}

func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func Ok[T any](w http.ResponseWriter, data T) {
	write(w, http.StatusOK, success[T]{Success: true, Data: data})
}

func Created[T any](w http.ResponseWriter, data T) {
	write(w, http.StatusCreated, success[T]{Success: true, Data: data})
}

func Msg(w http.ResponseWriter, message string) {
	write(w, http.StatusOK, successMsg{Success: true, Message: message})
}

func Err(w http.ResponseWriter, status int, message string) {
	write(w, status, errBody{Success: false, Message: message})
}

func Page[T any](w http.ResponseWriter, data T, meta Meta) {
	write(w, http.StatusOK, paginated[T]{Success: true, Data: data, Meta: meta})
}
