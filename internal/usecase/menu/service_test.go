package menu_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/usecase/menu"
)

type fakeRepository struct {
	listRestaurantsFn func(ctx context.Context) ([]domain.Restaurant, error)
	getMenuFn         func(ctx context.Context, restaurantID int32) ([]domain.MenuItem, error)
	upsertMenuItemFn  func(ctx context.Context, item domain.MenuItem) (domain.MenuItem, error)
	adjustStockFn     func(ctx context.Context, restaurantID, menuItemID int32, delta int32) (domain.MenuItem, error)

	called bool
}

func (f *fakeRepository) ListRestaurants(ctx context.Context) ([]domain.Restaurant, error) {
	f.called = true
	return f.listRestaurantsFn(ctx)
}

func (f *fakeRepository) GetMenu(ctx context.Context, restaurantID int32) ([]domain.MenuItem, error) {
	f.called = true
	return f.getMenuFn(ctx, restaurantID)
}

func (f *fakeRepository) UpsertMenuItem(ctx context.Context, item domain.MenuItem) (domain.MenuItem, error) {
	f.called = true
	return f.upsertMenuItemFn(ctx, item)
}

func (f *fakeRepository) AdjustStock(ctx context.Context, restaurantID, menuItemID int32, delta int32) (domain.MenuItem, error) {
	f.called = true
	return f.adjustStockFn(ctx, restaurantID, menuItemID, delta)
}

func TestService_ListRestaurants_DelegatesToRepository(t *testing.T) {
	want := []domain.Restaurant{{ID: 1, Name: "Sushi"}}
	repo := &fakeRepository{
		listRestaurantsFn: func(context.Context) ([]domain.Restaurant, error) {
			return want, nil
		},
	}
	svc := menu.NewService(repo)

	got, err := svc.ListRestaurants(context.Background())
	if err != nil {
		t.Fatalf("ListRestaurants() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListRestaurants() = %v, want %v", got, want)
	}
}

func TestService_GetMenu(t *testing.T) {
	t.Run("invalid restaurant id", func(t *testing.T) {
		repo := &fakeRepository{
			getMenuFn: func(context.Context, int32) ([]domain.MenuItem, error) {
				t.Fatal("repository must not be called for an invalid restaurant id")
				return nil, nil
			},
		}
		svc := menu.NewService(repo)

		_, err := svc.GetMenu(context.Background(), 0)
		if !errors.Is(err, domain.ErrInvalidRestaurantID) {
			t.Errorf("GetMenu() error = %v, want %v", err, domain.ErrInvalidRestaurantID)
		}
	})

	t.Run("delegates to repository", func(t *testing.T) {
		want := []domain.MenuItem{{ID: 1, RestaurantID: 1, Name: "Roll"}}
		repo := &fakeRepository{
			getMenuFn: func(_ context.Context, restaurantID int32) ([]domain.MenuItem, error) {
				if restaurantID != 1 {
					t.Errorf("restaurantID = %d, want 1", restaurantID)
				}
				return want, nil
			},
		}
		svc := menu.NewService(repo)

		got, err := svc.GetMenu(context.Background(), 1)
		if err != nil {
			t.Fatalf("GetMenu() error = %v, want nil", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("GetMenu() = %v, want %v", got, want)
		}
	})
}

func TestService_UpsertMenuItem_Validation(t *testing.T) {
	tests := []struct {
		name    string
		item    domain.MenuItem
		wantErr error
	}{
		{
			name:    "invalid restaurant id",
			item:    domain.MenuItem{RestaurantID: 0, Name: "Roll", PriceCents: 100},
			wantErr: domain.ErrInvalidRestaurantID,
		},
		{
			name:    "empty name",
			item:    domain.MenuItem{RestaurantID: 1, Name: "", PriceCents: 100},
			wantErr: domain.ErrEmptyMenuItemName,
		},
		{
			name:    "negative price",
			item:    domain.MenuItem{RestaurantID: 1, Name: "Roll", PriceCents: -1},
			wantErr: domain.ErrInvalidPrice,
		},
		{
			name:    "negative stock",
			item:    domain.MenuItem{RestaurantID: 1, Name: "Roll", PriceCents: 100, Stock: -1},
			wantErr: domain.ErrInvalidStock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{
				upsertMenuItemFn: func(context.Context, domain.MenuItem) (domain.MenuItem, error) {
					t.Fatal("repository must not be called when validation fails")
					return domain.MenuItem{}, nil
				},
			}
			svc := menu.NewService(repo)

			_, err := svc.UpsertMenuItem(context.Background(), tt.item)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UpsertMenuItem() error = %v, want %v", err, tt.wantErr)
			}
			if repo.called {
				t.Error("repository was called, want validation to short-circuit")
			}
		})
	}
}

func TestService_UpsertMenuItem_DelegatesToRepository(t *testing.T) {
	item := domain.MenuItem{RestaurantID: 1, Name: "Roll", PriceCents: 500, Stock: 3}
	want := domain.MenuItem{ID: 9, RestaurantID: 1, Name: "Roll", PriceCents: 500, Stock: 3}

	repo := &fakeRepository{
		upsertMenuItemFn: func(_ context.Context, gotItem domain.MenuItem) (domain.MenuItem, error) {
			if !reflect.DeepEqual(gotItem, item) {
				t.Errorf("item passed to repository = %v, want %v", gotItem, item)
			}
			return want, nil
		},
	}
	svc := menu.NewService(repo)

	got, err := svc.UpsertMenuItem(context.Background(), item)
	if err != nil {
		t.Fatalf("UpsertMenuItem() error = %v, want nil", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UpsertMenuItem() = %v, want %v", got, want)
	}
}

func TestService_UpsertMenuItem_PropagatesRepositoryError(t *testing.T) {
	repo := &fakeRepository{
		upsertMenuItemFn: func(context.Context, domain.MenuItem) (domain.MenuItem, error) {
			return domain.MenuItem{}, domain.ErrRestaurantNotFound
		},
	}
	svc := menu.NewService(repo)

	_, err := svc.UpsertMenuItem(context.Background(), domain.MenuItem{RestaurantID: 1, Name: "Roll", PriceCents: 100})
	if !errors.Is(err, domain.ErrRestaurantNotFound) {
		t.Errorf("UpsertMenuItem() error = %v, want %v", err, domain.ErrRestaurantNotFound)
	}
}

func TestService_AdjustStock(t *testing.T) {
	tests := []struct {
		name         string
		restaurantID int32
		menuItemID   int32
		wantErr      error
	}{
		{"invalid restaurant id", 0, 1, domain.ErrInvalidRestaurantID},
		{"invalid menu item id", 1, 0, domain.ErrInvalidMenuItemID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepository{
				adjustStockFn: func(context.Context, int32, int32, int32) (domain.MenuItem, error) {
					t.Fatal("repository must not be called when validation fails")
					return domain.MenuItem{}, nil
				},
			}
			svc := menu.NewService(repo)

			_, err := svc.AdjustStock(context.Background(), tt.restaurantID, tt.menuItemID, 1)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("AdjustStock() error = %v, want %v", err, tt.wantErr)
			}
		})
	}

	t.Run("delegates to repository, negative delta allowed", func(t *testing.T) {
		want := domain.MenuItem{ID: 1, RestaurantID: 1, Stock: 3}
		repo := &fakeRepository{
			adjustStockFn: func(_ context.Context, restaurantID, menuItemID, delta int32) (domain.MenuItem, error) {
				if restaurantID != 1 || menuItemID != 1 || delta != -2 {
					t.Errorf("got restaurantID=%d menuItemID=%d delta=%d, want 1,1,-2", restaurantID, menuItemID, delta)
				}
				return want, nil
			},
		}
		svc := menu.NewService(repo)

		got, err := svc.AdjustStock(context.Background(), 1, 1, -2)
		if err != nil {
			t.Fatalf("AdjustStock() error = %v, want nil", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("AdjustStock() = %v, want %v", got, want)
		}
	})

	t.Run("zero delta is allowed", func(t *testing.T) {
		want := domain.MenuItem{ID: 1, RestaurantID: 1, Stock: 3}
		repo := &fakeRepository{
			adjustStockFn: func(_ context.Context, _, _, delta int32) (domain.MenuItem, error) {
				if delta != 0 {
					t.Errorf("delta = %d, want 0", delta)
				}
				return want, nil
			},
		}
		svc := menu.NewService(repo)

		got, err := svc.AdjustStock(context.Background(), 1, 1, 0)
		if err != nil {
			t.Fatalf("AdjustStock() error = %v, want nil", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("AdjustStock() = %v, want %v", got, want)
		}
		if !repo.called {
			t.Error("repository was not called for delta == 0")
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		repo := &fakeRepository{
			adjustStockFn: func(context.Context, int32, int32, int32) (domain.MenuItem, error) {
				return domain.MenuItem{}, domain.ErrMenuItemOutOfStock
			},
		}
		svc := menu.NewService(repo)

		_, err := svc.AdjustStock(context.Background(), 1, 1, -100)
		if !errors.Is(err, domain.ErrMenuItemOutOfStock) {
			t.Errorf("AdjustStock() error = %v, want %v", err, domain.ErrMenuItemOutOfStock)
		}
	})
}
