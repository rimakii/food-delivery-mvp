package domain

import "errors"

var (
	ErrMenuItemOutOfStock  = errors.New("menu item out of stock")
	ErrOrderNotFound       = errors.New("order not found")
	ErrEmptyOrder          = errors.New("order must contain at least one item")
	ErrInvalidQuantity     = errors.New("quantity must be positive")
	ErrInvalidOrderID      = errors.New("order ID must be positive")
	ErrInvalidRestaurantID = errors.New("restaurant ID must be positive")
	ErrInvalidStatus       = errors.New("invalid order status")
)
