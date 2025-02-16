package handler

import (
	"encoding/json"
	"net/http"

	"github.com/every-base/graphql-go"
)

type HandlerOption func(h *Handler)

func WithExplorer(explorer http.Handler) HandlerOption {
	return func(h *Handler) {
		h.explorer = explorer
	}
}

var _ http.Handler = (*Handler)(nil)

func New(schema *graphql.Schema, opts ...HandlerOption) *Handler {
	h := &Handler{schema: schema}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

type Handler struct {
	schema   *graphql.Schema
	explorer http.Handler
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.exec(w, r)
	default:
		if r.Method == http.MethodGet && h.explorer != nil {
			h.explorer.ServeHTTP(w, r)
		} else {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func (h *Handler) exec(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := h.schema.Exec(r.Context(), params.Query, params.OperationName, params.Variables)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(resp)
}
