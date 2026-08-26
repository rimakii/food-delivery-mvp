package order

import (
	"context"

	"avito-kitchen/internal/domain"
)

// Repository объявлен на стороне usecase (не там, где реализация), чтобы
// его можно было замокать в тестах Service независимо от postgres.
type Repository interface {
	// CreateOrder создаёт заказ и атомарно списывает stock по каждой позиции
	// в рамках одной транзакции. Возвращает domain.ErrMenuItemOutOfStock,
	// если для какой-то позиции не хватает остатка.
	CreateOrder(ctx context.Context, restaurantID int32, items []domain.OrderItem) (domain.Order, error)

	GetOrder(ctx context.Context, id int64) (domain.Order, error)

	ListOrdersByRestaurant(ctx context.Context, restaurantID int32) ([]domain.Order, error)

	UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error
}
