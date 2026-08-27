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

	ErrRestaurantNotFound = errors.New("restaurant not found")
	ErrMenuItemNotFound   = errors.New("menu item not found")
	ErrEmptyMenuItemName  = errors.New("menu item name must not be empty")
	ErrInvalidPrice       = errors.New("price must not be negative")
	ErrInvalidStock       = errors.New("stock must not be negative")
	ErrInvalidMenuItemID  = errors.New("menu item ID must be positive")
)
