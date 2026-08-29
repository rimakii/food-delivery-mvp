package postgres_test

import (
	"context"
	"errors"
	"testing"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/repository/postgres"
)

func TestMenuRepository_UpsertMenuItem_CreateThenUpdate(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := postgres.NewMenuRepository(pool)

	restaurantID := seedRestaurant(t, pool)

	created, err := repo.UpsertMenuItem(ctx, domain.MenuItem{
		RestaurantID: restaurantID,
		Name:         "Roll",
		PriceCents:   500,
		Stock:        3,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("created.ID = 0, want non-zero")
	}

	updated, err := repo.UpsertMenuItem(ctx, domain.MenuItem{
		ID:           created.ID,
		RestaurantID: restaurantID,
		Name:         "Roll XL",
		PriceCents:   600,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Roll XL" || updated.PriceCents != 600 {
		t.Errorf("updated = %+v, want name=Roll XL price=600", updated)
	}
	if updated.Stock != 3 {
		t.Errorf("updated.Stock = %d, want 3 (update must not touch stock)", updated.Stock)
	}
}

func TestMenuRepository_UpsertMenuItem_RestaurantNotFound(t *testing.T) {
	pool := testPool(t)
	repo := postgres.NewMenuRepository(pool)

	_, err := repo.UpsertMenuItem(context.Background(), domain.MenuItem{
		RestaurantID: 999999999,
		Name:         "Roll",
		PriceCents:   500,
	})
	if !errors.Is(err, domain.ErrRestaurantNotFound) {
		t.Fatalf("error = %v, want %v", err, domain.ErrRestaurantNotFound)
	}
}

func TestMenuRepository_AdjustStock_CannotGoNegative(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := postgres.NewMenuRepository(pool)

	restaurantID := seedRestaurant(t, pool)
	menuItemID := seedMenuItem(t, pool, restaurantID, 2)

	_, err := repo.AdjustStock(ctx, restaurantID, menuItemID, -5)
	if !errors.Is(err, domain.ErrMenuItemOutOfStock) {
		t.Fatalf("error = %v, want %v", err, domain.ErrMenuItemOutOfStock)
	}

	var stock int32
	if err := pool.QueryRow(ctx, `SELECT stock FROM menu_items WHERE id = $1`, menuItemID).Scan(&stock); err != nil {
		t.Fatalf("read stock: %v", err)
	}
	if stock != 2 {
		t.Errorf("stock = %d, want 2 (unchanged)", stock)
	}
}

func TestMenuRepository_GetMenu_RestaurantNotFound(t *testing.T) {
	pool := testPool(t)
	repo := postgres.NewMenuRepository(pool)

	_, err := repo.GetMenu(context.Background(), 999999999)
	if !errors.Is(err, domain.ErrRestaurantNotFound) {
		t.Fatalf("error = %v, want %v", err, domain.ErrRestaurantNotFound)
	}
}
