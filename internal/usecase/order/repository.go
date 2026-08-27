package order

import (
	"context"

	"avito-kitchen/internal/domain"
)

type Repository interface {
	CreateOrder(ctx context.Context, restaurantID int32, items []domain.OrderItem) (domain.Order, error)

	GetOrder(ctx context.Context, id int64) (domain.Order, error)

	ListOrdersByRestaurant(ctx context.Context, restaurantID int32) ([]domain.Order, error)

	UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error
}
