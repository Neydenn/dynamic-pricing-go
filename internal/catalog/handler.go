package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createReq struct {
	Name      string  `json:"name"`
	BasePrice float64 `json:"base_price"`
	Stock     int     `json:"stock"`
}

type updateReq struct {
	Name      string  `json:"name"`
	BasePrice float64 `json:"base_price"`
}

type stockReq struct {
	Stock int `json:"stock"`
}

func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	r.Post("/products", h.create)
	r.Get("/products/{id}", h.get)
	r.Put("/products/{id}", h.update)
	r.Patch("/products/{id}/stock", h.updateStock)
	return r
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	p, err := h.svc.Create(r.Context(), req.Name, req.BasePrice, req.Stock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, p, http.StatusCreated)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	p, err := h.svc.Get(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, p, http.StatusOK)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	var req updateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	p, err := h.svc.Update(r.Context(), id, req.Name, req.BasePrice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, p, http.StatusOK)
}

func (h *Handler) updateStock(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	var req stockReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	p, err := h.svc.UpdateStock(r.Context(), id, req.Stock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, p, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
