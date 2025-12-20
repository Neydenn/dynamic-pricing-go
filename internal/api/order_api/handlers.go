package order_api

import (
    "encoding/json"
    "net/http"

    "dynamic-pricing/internal/services/order"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
)

type Handler struct{ svc *order.Service }

func NewHandler(svc *order.Service) *Handler { return &Handler{svc: svc} }

type createUserReq struct{ Email string `json:"email"` }

type placeOrderReq struct {
    UserID    string `json:"user_id"`
    ProductID string `json:"product_id"`
    Qty       int    `json:"qty"`
}

func (h *Handler) Routes() http.Handler {
    r := chi.NewRouter()
    r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
    r.Post("/users", h.createUser)
    r.Post("/orders", h.placeOrder)
    r.Post("/orders/{id}/cancel", h.cancelOrder)
    r.Get("/orders/{id}", h.getOrder)
    return r
}

func (h *Handler) createUser(w http.ResponseWriter, r *http.Request) {
    var req createUserReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad json", http.StatusBadRequest)
        return
    }
    u, err := h.svc.CreateUser(r.Context(), req.Email)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, u, http.StatusCreated)
}

func (h *Handler) placeOrder(w http.ResponseWriter, r *http.Request) {
    var req placeOrderReq
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad json", http.StatusBadRequest)
        return
    }
    userID, err := uuid.Parse(req.UserID)
    if err != nil {
        http.Error(w, "bad user_id", http.StatusBadRequest)
        return
    }
    productID, err := uuid.Parse(req.ProductID)
    if err != nil {
        http.Error(w, "bad product_id", http.StatusBadRequest)
        return
    }
    o, err := h.svc.PlaceOrder(r.Context(), userID, productID, req.Qty)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, o, http.StatusCreated)
}

func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        http.Error(w, "bad id", http.StatusBadRequest)
        return
    }
    o, err := h.svc.CancelOrder(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    writeJSON(w, o, http.StatusOK)
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
    id, err := uuid.Parse(chi.URLParam(r, "id"))
    if err != nil {
        http.Error(w, "bad id", http.StatusBadRequest)
        return
    }
    o, err := h.svc.GetOrder(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    writeJSON(w, o, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v any, status int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    _ = json.NewEncoder(w).Encode(v)
}

