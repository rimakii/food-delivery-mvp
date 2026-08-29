package handlers

import (
	"net/http"
	"time"

	"avito-kitchen/internal/usecase/menu"
)

type RestaurantsHandler struct {
	svc *menu.Service
}

func NewRestaurantsHandler(svc *menu.Service) *RestaurantsHandler {
	return &RestaurantsHandler{svc: svc}
}

type restaurantResponse struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *RestaurantsHandler) List(w http.ResponseWriter, r *http.Request) {
	restaurants, err := h.svc.ListRestaurants(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]restaurantResponse, 0, len(restaurants))
	for _, res := range restaurants {
		resp = append(resp, restaurantResponse{ID: res.ID, Name: res.Name, CreatedAt: res.CreatedAt})
	}

	writeJSON(w, http.StatusOK, resp)
}
