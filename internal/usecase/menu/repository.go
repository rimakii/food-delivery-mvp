package menu

import (
	"context"

	"avito-kitchen/internal/domain"
)

type Repository interface {
	ListRestaurants(ctx context.Context) ([]domain.Restaurant, error)

	GetMenu(ctx context.Context, restaurantID int32) ([]domain.MenuItem, error)

	UpsertMenuItem(ctx context.Context, item domain.MenuItem) (domain.MenuItem, error)

	// AdjustStock меняет остаток на delta (может быть отрицательным) одним
	// атомарным UPDATE — не read-modify-write, поэтому безопасно относительно
	// параллельного списания в CreateOrder. Уход в минус отклоняется
	// CHECK-constraint'ом в БД, возвращает domain.ErrMenuItemOutOfStock.
	AdjustStock(ctx context.Context, restaurantID, menuItemID int32, delta int32) (domain.MenuItem, error)
}
