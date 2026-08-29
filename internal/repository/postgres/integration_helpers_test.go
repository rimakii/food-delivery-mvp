//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"avito-kitchen/internal/repository/postgres"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	pool, err := postgres.NewPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

func seedRestaurant(t *testing.T, pool *pgxpool.Pool) int32 {
	t.Helper()

	var id int32
	err := pool.QueryRow(context.Background(),
		`INSERT INTO restaurants (name) VALUES ($1) RETURNING id`, t.Name(),
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed restaurant: %v", err)
	}

	return id
}

func seedMenuItem(t *testing.T, pool *pgxpool.Pool, restaurantID, stock int32) int32 {
	t.Helper()

	var id int32
	err := pool.QueryRow(context.Background(),
		`INSERT INTO menu_items (restaurant_id, name, price_cents, stock) VALUES ($1, $2, $3, $4) RETURNING id`,
		restaurantID, "item", 100, stock,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed menu item: %v", err)
	}

	return id
}
