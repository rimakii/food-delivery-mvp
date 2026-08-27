package order_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/usecase/order"
)

// fakeRepository — ручной мок order.Repository: даёт задать поведение через
// функции-поля и проверить, дошёл ли вызов до репозитория.
type fakeRepository struct {
	createOrderFn  func(ctx context.Context, restaurantID int32, items []domain.OrderItem) (domain.Order, error)
	getOrderFn     func(ctx context.Context, id int64) (domain.Order, error)
	listFn         func(ctx context.Context, restaurantID int32) ([]domain.Order, error)
	updateStatusFn func(ctx context.Context, id int64, status domain.OrderStatus) error

	called bool
}

func (f *fakeRepository) CreateOrder(ctx context.Context, restaurantID int32, items []domain.OrderItem) (domain.Order, error) {
	f.called = true
	return f.createOrderFn(ctx, restaurantID, items)
}

func (f *fakeRepository) GetOrder(ctx context.Context, id int64) (domain.Order, error) {
	f.called = true
	return f.getOrderFn(ctx, id)
}

func (f *fakeRepository) ListOrdersByRestaurant(ctx context.Context, restaurantID int32) ([]domain.Order, error) {
	f.called = true
	return f.listFn(ctx, restaurantID)
}

func (f *fakeRepository) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error {
	f.called = true
	return f.updateStatusFn(ctx, id, status)
}

func TestService_CreateOrder_Validation(t *testing.T) {
	tests := []struct {
		name         string
		restaurantID int32
		items        []domain.OrderItem
		wantErr      error
	}{
		{
			name:         "empty items",
			restaurantID: 1,
			items:        nil,
			wantErr:      domain.ErrEmptyOrder,
		},
		{
			name:         "negative quantity",
			restaurantID: 1,
			items:        []domain.OrderItem{{MenuItemID: 1, Quantity: -1}},
			wantErr:      domain.ErrInvalidQuantity,
		},
		{
			name:         "zero quantity",
			restaurantID: 1,
			items:        []domain.OrderItem{{MenuItemID: 1, Quantity: 0}},
			wantErr:      domain.ErrInvalidQuantity,
		},
		{
			name:         "invalid restaurant id",
			restaurantID: 0,
			items:        []domain.OrderItem{{MenuItemID: 1, Quantity: 1}},
			wantErr:      domain.ErrInvalidRestaurantID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{
				createOrderFn: func(context.Context, int32, []domain.OrderItem) (domain.Order, error) {
					t.Fatal("repository must not be called when validation fails")
					return domain.Order{}, nil
				},
			}
			svc := order.NewService(repo)

			_, err := svc.CreateOrder(context.Background(), tt.restaurantID, tt.items)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CreateOrder() error = %v, want %v", err, tt.wantErr)
			}
			if repo.called {
				t.Error("repository was called, want validation to short-circuit")
			}
		})
	}
}

func TestService_CreateOrder_DelegatesToRepository(t *testing.T) {
	items := []domain.OrderItem{{MenuItemID: 1, Quantity: 2}}
	want := domain.Order{ID: 42, RestaurantID: 1, Status: domain.OrderStatusNew}

	repo := &fakeRepository{
		createOrderFn: func(_ context.Context, restaurantID int32, gotItems []domain.OrderItem) (domain.Order, error) {
			if restaurantID != 1 {
				t.Errorf("restaurantID = %d, want 1", restaurantID)
			}
			if len(gotItems) != 1 {
				t.Errorf("items = %v, want 1 item", gotItems)
			}
			return want, nil
		},
	}
	svc := order.NewService(repo)

	got, err := svc.CreateOrder(context.Background(), 1, items)
	if err != nil {
		t.Fatalf("CreateOrder() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CreateOrder() = %v, want %v", got, want)
	}
	if !repo.called {
		t.Error("repository was not called")
	}
}

func TestService_CreateOrder_PropagatesRepositoryError(t *testing.T) {
	repo := &fakeRepository{
		createOrderFn: func(context.Context, int32, []domain.OrderItem) (domain.Order, error) {
			return domain.Order{}, domain.ErrMenuItemOutOfStock
		},
	}
	svc := order.NewService(repo)

	_, err := svc.CreateOrder(context.Background(), 1, []domain.OrderItem{{MenuItemID: 1, Quantity: 1}})
	if !errors.Is(err, domain.ErrMenuItemOutOfStock) {
		t.Errorf("CreateOrder() error = %v, want %v", err, domain.ErrMenuItemOutOfStock)
	}
}

func TestService_GetOrder(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		repo := &fakeRepository{
			getOrderFn: func(context.Context, int64) (domain.Order, error) {
				t.Fatal("repository must not be called for an invalid id")
				return domain.Order{}, nil
			},
		}
		svc := order.NewService(repo)

		_, err := svc.GetOrder(context.Background(), 0)
		if !errors.Is(err, domain.ErrInvalidOrderID) {
			t.Errorf("GetOrder() error = %v, want %v", err, domain.ErrInvalidOrderID)
		}
	})

	t.Run("delegates to repository", func(t *testing.T) {
		want := domain.Order{ID: 7}
		repo := &fakeRepository{
			getOrderFn: func(_ context.Context, id int64) (domain.Order, error) {
				if id != 7 {
					t.Errorf("id = %d, want 7", id)
				}
				return want, nil
			},
		}
		svc := order.NewService(repo)

		got, err := svc.GetOrder(context.Background(), 7)
		if err != nil {
			t.Fatalf("GetOrder() error = %v, want nil", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetOrder() = %v, want %v", got, want)
		}
	})
}

func TestService_ListOrdersByRestaurant_InvalidID(t *testing.T) {
	repo := &fakeRepository{
		listFn: func(context.Context, int32) ([]domain.Order, error) {
			t.Fatal("repository must not be called for an invalid restaurant id")
			return nil, nil
		},
	}
	svc := order.NewService(repo)

	_, err := svc.ListOrdersByRestaurant(context.Background(), 0)
	if !errors.Is(err, domain.ErrInvalidRestaurantID) {
		t.Errorf("ListOrdersByRestaurant() error = %v, want %v", err, domain.ErrInvalidRestaurantID)
	}
}

func TestService_UpdateStatus(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		status  domain.OrderStatus
		wantErr error
	}{
		{"invalid id", 0, domain.OrderStatusConfirmed, domain.ErrInvalidOrderID},
		{"invalid status", 1, domain.OrderStatus("delivered"), domain.ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{
				updateStatusFn: func(context.Context, int64, domain.OrderStatus) error {
					t.Fatal("repository must not be called when validation fails")
					return nil
				},
			}
			svc := order.NewService(repo)

			err := svc.UpdateStatus(context.Background(), tt.id, tt.status)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UpdateStatus() error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	t.Run("delegates to repository", func(t *testing.T) {
		repo := &fakeRepository{
			updateStatusFn: func(_ context.Context, id int64, status domain.OrderStatus) error {
				if id != 1 || status != domain.OrderStatusConfirmed {
					t.Errorf("got id=%d status=%s, want id=1 status=confirmed", id, status)
				}
				return nil
			},
		}
		svc := order.NewService(repo)

		if err := svc.UpdateStatus(context.Background(), 1, domain.OrderStatusConfirmed); err != nil {
			t.Errorf("UpdateStatus() error = %v, want nil", err)
		}
	})
}
