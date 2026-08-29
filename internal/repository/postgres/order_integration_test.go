//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/repository/postgres"
)

func TestOrderRepository_CreateOrder_DecrementsStock(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := postgres.NewOrderRepository(pool)

	restaurantID := seedRestaurant(t, pool)
	menuItemID := seedMenuItem(t, pool, restaurantID, 5)

	o, err := repo.CreateOrder(ctx, restaurantID, []domain.OrderItem{{MenuItemID: menuItemID, Quantity: 2}})
	if err != nil {
		t.Fatalf("CreateOrder() error = %v", err)
	}
	if len(o.Items) != 1 || o.Items[0].PriceCents != 100 {
		t.Errorf("unexpected items: %+v", o.Items)
	}

	var stock int32
	if err := pool.QueryRow(ctx, `SELECT stock FROM menu_items WHERE id = $1`, menuItemID).Scan(&stock); err != nil {
		t.Fatalf("read stock: %v", err)
	}
	if stock != 3 {
		t.Errorf("stock = %d, want 3", stock)
	}
}

func TestOrderRepository_CreateOrder_OutOfStock_RollsBack(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := postgres.NewOrderRepository(pool)

	restaurantID := seedRestaurant(t, pool)
	menuItemID := seedMenuItem(t, pool, restaurantID, 1)

	_, err := repo.CreateOrder(ctx, restaurantID, []domain.OrderItem{{MenuItemID: menuItemID, Quantity: 5}})
	if !errors.Is(err, domain.ErrMenuItemOutOfStock) {
		t.Fatalf("CreateOrder() error = %v, want %v", err, domain.ErrMenuItemOutOfStock)
	}

	var orderCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM orders WHERE restaurant_id = $1`, restaurantID).Scan(&orderCount); err != nil {
		t.Fatalf("count orders: %v", err)
	}
	if orderCount != 0 {
		t.Errorf("orders created = %d, want 0 (transaction should roll back)", orderCount)
	}

	var stock int32
	if err := pool.QueryRow(ctx, `SELECT stock FROM menu_items WHERE id = $1`, menuItemID).Scan(&stock); err != nil {
		t.Fatalf("read stock: %v", err)
	}
	if stock != 1 {
		t.Errorf("stock = %d, want 1 (unchanged)", stock)
	}
}

func TestOrderRepository_CreateOrder_RestaurantNotFound(t *testing.T) {
	pool := testPool(t)
	repo := postgres.NewOrderRepository(pool)

	_, err := repo.CreateOrder(context.Background(), 999999999, []domain.OrderItem{{MenuItemID: 1, Quantity: 1}})
	if !errors.Is(err, domain.ErrRestaurantNotFound) {
		t.Fatalf("CreateOrder() error = %v, want %v", err, domain.ErrRestaurantNotFound)
	}
}

func TestOrderRepository_CreateOrder_ConcurrentOrders_NoOverselling(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	repo := postgres.NewOrderRepository(pool)

	const stock = 5
	const attempts = 20

	restaurantID := seedRestaurant(t, pool)
	menuItemID := seedMenuItem(t, pool, restaurantID, stock)

	var wg sync.WaitGroup
	var mu sync.Mutex
	succeeded, outOfStock, unexpected := 0, 0, 0

	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := repo.CreateOrder(ctx, restaurantID, []domain.OrderItem{{MenuItemID: menuItemID, Quantity: 1}})

			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				succeeded++
			case errors.Is(err, domain.ErrMenuItemOutOfStock):
				outOfStock++
			default:
				unexpected++
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	wg.Wait()

	if unexpected != 0 {
		t.Fatalf("%d unexpected errors", unexpected)
	}
	if succeeded != stock {
		t.Errorf("succeeded = %d, want %d", succeeded, stock)
	}
	if outOfStock != attempts-stock {
		t.Errorf("outOfStock = %d, want %d", outOfStock, attempts-stock)
	}

	var finalStock int32
	if err := pool.QueryRow(ctx, `SELECT stock FROM menu_items WHERE id = $1`, menuItemID).Scan(&finalStock); err != nil {
		t.Fatalf("read stock: %v", err)
	}
	if finalStock != 0 {
		t.Errorf("finalStock = %d, want 0 — overselling or underselling happened", finalStock)
	}
}
