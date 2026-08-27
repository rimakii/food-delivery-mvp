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
	ID        int32     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

type MenuItem struct {
	ID           int32     `db:"id"`
	RestaurantID int32     `db:"restaurant_id"`
	Name         string    `db:"name"`
	PriceCents   int64     `db:"price_cents"`
	Stock        int32     `db:"stock"`
	CreatedAt    time.Time `db:"created_at"`
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
