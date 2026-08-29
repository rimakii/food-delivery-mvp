package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/usecase/menu"
)

type MenuHandler struct {
	svc *menu.Service
}

func NewMenuHandler(svc *menu.Service) *MenuHandler {
	return &MenuHandler{svc: svc}
}

type createMenuItemRequest struct {
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
	Stock      int32  `json:"stock,omitempty"`
}

type updateMenuItemRequest struct {
	Name       string `json:"name"`
	PriceCents int64  `json:"price_cents"`
}

type adjustStockRequest struct {
	Delta int32 `json:"delta"`
}

type menuItemResponse struct {
	ID           int32     `json:"id"`
	RestaurantID int32     `json:"restaurant_id"`
	Name         string    `json:"name"`
	PriceCents   int64     `json:"price_cents"`
	Stock        int32     `json:"stock"`
	CreatedAt    time.Time `json:"created_at"`
}

func newMenuItemResponse(item domain.MenuItem) menuItemResponse {
	return menuItemResponse{
		ID:           item.ID,
		RestaurantID: item.RestaurantID,
		Name:         item.Name,
		PriceCents:   item.PriceCents,
		Stock:        item.Stock,
		CreatedAt:    item.CreatedAt,
	}
}

func (h *MenuHandler) Get(w http.ResponseWriter, r *http.Request) {
	restaurantID, ok := pathInt32(r, "restaurantID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid restaurant id"})
		return
	}

	items, err := h.svc.GetMenu(r.Context(), restaurantID)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]menuItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, newMenuItemResponse(item))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *MenuHandler) Create(w http.ResponseWriter, r *http.Request) {
	restaurantID, ok := pathInt32(r, "restaurantID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid restaurant id"})
		return
	}

	var req createMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	item, err := h.svc.UpsertMenuItem(r.Context(), domain.MenuItem{
		RestaurantID: restaurantID,
		Name:         req.Name,
		PriceCents:   req.PriceCents,
		Stock:        req.Stock,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newMenuItemResponse(item))
}

func (h *MenuHandler) Update(w http.ResponseWriter, r *http.Request) {
	restaurantID, ok := pathInt32(r, "restaurantID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid restaurant id"})
		return
	}

	itemID, ok := pathInt32(r, "itemID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid menu item id"})
		return
	}

	var req updateMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	item, err := h.svc.UpsertMenuItem(r.Context(), domain.MenuItem{
		ID:           itemID,
		RestaurantID: restaurantID,
		Name:         req.Name,
		PriceCents:   req.PriceCents,
	})
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newMenuItemResponse(item))
}

func (h *MenuHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	restaurantID, ok := pathInt32(r, "restaurantID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid restaurant id"})
		return
	}

	menuItemID, ok := pathInt32(r, "itemID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid menu item id"})
		return
	}

	var req adjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	item, err := h.svc.AdjustStock(r.Context(), restaurantID, menuItemID, req.Delta)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newMenuItemResponse(item))
}
