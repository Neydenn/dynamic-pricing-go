package pricing

import (
	"encoding/json"
	"net/http"

	"errors"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	repo   *Repository
	engine *Engine
}

func NewHandler(repo *Repository, engine *Engine) *Handler {
	return &Handler{repo: repo, engine: engine}
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Get("/prices/{product_id}", h.getPrice)
	return r
}

func (h *Handler) getPrice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "product_id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	p, err := h.repo.GetPrice(r.Context(), id)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(p)
		return
	}
	if errors.Is(err, pgx.ErrNoRows) {
		p2, err2 := h.engine.ComputeAndPersistCurrentPrice(r.Context(), id)
		if err2 == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(p2)
			return
		}
		if errors.Is(err2, ErrUnknownProduct) {
			http.Error(w, "price not found (unknown product)", http.StatusNotFound)
			return
		}
		http.Error(w, err2.Error(), http.StatusInternalServerError)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
