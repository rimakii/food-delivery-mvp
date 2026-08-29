package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/usecase/order"
)

type OrdersHandler struct {
	svc *order.Service
}

func NewOrdersHandler(svc *order.Service) *OrdersHandler {
	return &OrdersHandler{svc: svc}
}

type createOrderItemRequest struct {
	MenuItemID int32 `json:"menu_item_id"`
	Quantity   int   `json:"quantity"`
}

type createOrderRequest struct {
	RestaurantID int32                    `json:"restaurant_id"`
	Items        []createOrderItemRequest `json:"items"`
}

type updateStatusRequest struct {
	Status domain.OrderStatus `json:"status"`
}

type orderItemResponse struct {
	ID         int64 `json:"id"`
	MenuItemID int32 `json:"menu_item_id"`
	Quantity   int   `json:"quantity"`
	PriceCents int64 `json:"price_cents"`
}

type orderResponse struct {
	ID           int64               `json:"id"`
	RestaurantID int32               `json:"restaurant_id"`
	Status       domain.OrderStatus  `json:"status"`
	Items        []orderItemResponse `json:"items,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
}

func newOrderResponse(o domain.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, orderItemResponse{
			ID:         item.ID,
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			PriceCents: item.PriceCents,
		})
	}

	return orderResponse{
		ID:           o.ID,
		RestaurantID: o.RestaurantID,
		Status:       o.Status,
		Items:        items,
		CreatedAt:    o.CreatedAt,
	}
}

func (h *OrdersHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	items := make([]domain.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.OrderItem{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
		})
	}

	o, err := h.svc.CreateOrder(r.Context(), req.RestaurantID, items)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, newOrderResponse(o))
}

func (h *OrdersHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt64(r, "orderID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid order id"})
		return
	}

	o, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, newOrderResponse(o))
}

func (h *OrdersHandler) ListByRestaurant(w http.ResponseWriter, r *http.Request) {
	restaurantID, ok := pathInt32(r, "restaurantID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid restaurant id"})
		return
	}

	orders, err := h.svc.ListOrdersByRestaurant(r.Context(), restaurantID)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		resp = append(resp, newOrderResponse(o))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *OrdersHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathInt64(r, "orderID")
	if !ok {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid order id"})
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if err := h.svc.UpdateStatus(r.Context(), id, req.Status); err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
