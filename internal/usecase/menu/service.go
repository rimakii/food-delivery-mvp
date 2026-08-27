package menu

import (
	"context"

	"avito-kitchen/internal/domain"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListRestaurants(ctx context.Context) ([]domain.Restaurant, error) {
	return s.repo.ListRestaurants(ctx)
}

func (s *Service) GetMenu(ctx context.Context, restaurantID int32) ([]domain.MenuItem, error) {
	if restaurantID <= 0 {
		return nil, domain.ErrInvalidRestaurantID
	}

	return s.repo.GetMenu(ctx, restaurantID)
}

func (s *Service) UpsertMenuItem(ctx context.Context, item domain.MenuItem) (domain.MenuItem, error) {
	if item.RestaurantID <= 0 {
		return domain.MenuItem{}, domain.ErrInvalidRestaurantID
	}
	if item.Name == "" {
		return domain.MenuItem{}, domain.ErrEmptyMenuItemName
	}
	if item.PriceCents < 0 {
		return domain.MenuItem{}, domain.ErrInvalidPrice
	}
	if item.Stock < 0 {
		return domain.MenuItem{}, domain.ErrInvalidStock
	}

	return s.repo.UpsertMenuItem(ctx, item)
}

func (s *Service) AdjustStock(ctx context.Context, restaurantID, menuItemID int32, delta int32) (domain.MenuItem, error) {
	if restaurantID <= 0 {
		return domain.MenuItem{}, domain.ErrInvalidRestaurantID
	}
	if menuItemID <= 0 {
		return domain.MenuItem{}, domain.ErrInvalidMenuItemID
	}

	return s.repo.AdjustStock(ctx, restaurantID, menuItemID, delta)
}
