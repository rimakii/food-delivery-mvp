package domain

import "time"

type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "new"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
	// промежуточные статусы (in_preparation, delivered)?
)

type Restaurant struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type MenuItem struct {
	ID           int64
	RestaurantID int64
	Name         string
	PriceCents   int64
	Stock        int
	CreatedAt    time.Time
}

type Order struct {
	ID           int64
	RestaurantID int64
	Status       OrderStatus
	Items        []OrderItem
	CreatedAt    time.Time
}

type OrderItem struct {
	ID         int64
	OrderID    int64
	MenuItemID int64
	Quantity   int
	PriceCents int64
}
