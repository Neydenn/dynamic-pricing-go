package pricing_api

import (
    "encoding/json"
    "errors"
    "net/http"

    "dynamic-pricing/internal/services/pricing"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
)

type Handler struct {
    repo pricing.PriceRepository
    eng  *pricing.Engine
}

func NewHandler(repo pricing.PriceRepository, eng *pricing.Engine) *Handler {
    return &Handler{repo: repo, eng: eng}
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
        writeJSON(w, p, http.StatusOK)
        return
    }
    if errors.Is(err, pgx.ErrNoRows) {
        p2, err2 := h.eng.ComputeAndPersistCurrentPrice(r.Context(), id)
        if err2 == nil {
            writeJSON(w, p2, http.StatusOK)
            return
        }
        if errors.Is(err2, pricing.ErrUnknownProduct) {
            http.Error(w, "price not found (unknown product)", http.StatusNotFound)
            return
        }
        http.Error(w, err2.Error(), http.StatusInternalServerError)
        return
    }
    http.Error(w, err.Error(), http.StatusInternalServerError)
}

func writeJSON(w http.ResponseWriter, v any, status int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

