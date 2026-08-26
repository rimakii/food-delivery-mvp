package domain

import "time"

type OrderStatus string

const (
	OrderStatusNew       OrderStatus = "new"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
	// промежуточные статусы (in_preparation, delivered)?
)

func (s OrderStatus) Valid() bool {
	switch s {
	case OrderStatusNew, OrderStatusConfirmed, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

type Restaurant struct {
	ID        int32
	Name      string
	CreatedAt time.Time
}

type MenuItem struct {
	ID           int32
	RestaurantID int32
	Name         string
	PriceCents   int64
	Stock        int
	CreatedAt    time.Time
}

type Order struct {
	ID           int64
	RestaurantID int32
	Status       OrderStatus
	Items        []OrderItem
	CreatedAt    time.Time
}

type OrderItem struct {
	ID         int64 `db:"id"`
	OrderID    int64 `db:"order_id"`
	MenuItemID int32 `db:"menu_item_id"`
	Quantity   int   `db:"quantity"`
	PriceCents int64 `db:"price_cents"`
}
