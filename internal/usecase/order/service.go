package order

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

func (s *Service) CreateOrder(ctx context.Context, restaurantID int32, items []domain.OrderItem) (domain.Order, error) {
	if restaurantID <= 0 {
		return domain.Order{}, domain.ErrInvalidRestaurantID
	}
	if len(items) == 0 {
		return domain.Order{}, domain.ErrEmptyOrder
	}
	for _, item := range items {
		if item.Quantity <= 0 {
			return domain.Order{}, domain.ErrInvalidQuantity
		}
	}

	return s.repo.CreateOrder(ctx, restaurantID, items)
}

func (s *Service) GetOrder(ctx context.Context, id int64) (domain.Order, error) {
	if id <= 0 {
		return domain.Order{}, domain.ErrInvalidOrderID
	}

	return s.repo.GetOrder(ctx, id)
}

func (s *Service) ListOrdersByRestaurant(ctx context.Context, restaurantID int32) ([]domain.Order, error) {
	if restaurantID <= 0 {
		return nil, domain.ErrInvalidRestaurantID
	}

	return s.repo.ListOrdersByRestaurant(ctx, restaurantID)
}

func (s *Service) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error {
	if id <= 0 {
		return domain.ErrInvalidOrderID
	}
	if !status.Valid() {
		return domain.ErrInvalidStatus
	}

	return s.repo.UpdateStatus(ctx, id, status)
}
