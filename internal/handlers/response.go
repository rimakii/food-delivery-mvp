package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"avito-kitchen/internal/domain"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("handlers: encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, domain.ErrMenuItemOutOfStock):
		status = http.StatusConflict
	case errors.Is(err, domain.ErrOrderNotFound),
		errors.Is(err, domain.ErrRestaurantNotFound),
		errors.Is(err, domain.ErrMenuItemNotFound):
		status = http.StatusNotFound
	case errors.Is(err, domain.ErrEmptyOrder),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidOrderID),
		errors.Is(err, domain.ErrInvalidRestaurantID),
		errors.Is(err, domain.ErrInvalidStatus),
		errors.Is(err, domain.ErrEmptyMenuItemName),
		errors.Is(err, domain.ErrInvalidPrice),
		errors.Is(err, domain.ErrInvalidStock),
		errors.Is(err, domain.ErrInvalidMenuItemID):
		status = http.StatusBadRequest
	default:
		log.Printf("handlers: internal error: %v", err)
	}

	writeJSON(w, status, errorResponse{Error: err.Error()})
}
