package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/usecase/menu"
)

const (
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
)

var _ menu.Repository = (*MenuRepository)(nil)

type MenuRepository struct {
	pool *pgxpool.Pool
}

func NewMenuRepository(pool *pgxpool.Pool) *MenuRepository {
	return &MenuRepository{pool: pool}
}

func (r *MenuRepository) ListRestaurants(ctx context.Context) ([]domain.Restaurant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, created_at FROM restaurants ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("postgres: list restaurants: %w", err)
	}
	defer rows.Close()

	restaurants, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Restaurant])
	if err != nil {
		return nil, fmt.Errorf("postgres: scan restaurants: %w", err)
	}

	return restaurants, nil
}

func (r *MenuRepository) GetMenu(ctx context.Context, restaurantID int32) ([]domain.MenuItem, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM restaurants WHERE id = $1)`,
		restaurantID,
	).Scan(&exists); err != nil {
		return nil, fmt.Errorf("postgres: check restaurant exists: %w", err)
	}
	if !exists {
		return nil, domain.ErrRestaurantNotFound
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, restaurant_id, name, price_cents, stock, created_at
		 FROM menu_items
		 WHERE restaurant_id = $1
		 ORDER BY id`,
		restaurantID,
	)
	if err != nil {
		return nil, fmt.Errorf("postgres: get menu: %w", err)
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.MenuItem])
	if err != nil {
		return nil, fmt.Errorf("postgres: scan menu items: %w", err)
	}

	return items, nil
}

func (r *MenuRepository) UpsertMenuItem(ctx context.Context, item domain.MenuItem) (domain.MenuItem, error) {
	if item.ID == 0 {
		return r.insertMenuItem(ctx, item)
	}

	return r.updateMenuItem(ctx, item)
}

func (r *MenuRepository) insertMenuItem(ctx context.Context, item domain.MenuItem) (domain.MenuItem, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO menu_items (restaurant_id, name, price_cents, stock)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, restaurant_id, name, price_cents, stock, created_at`,
		item.RestaurantID, item.Name, item.PriceCents, item.Stock,
	).Scan(&item.ID, &item.RestaurantID, &item.Name, &item.PriceCents, &item.Stock, &item.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.MenuItem{}, domain.ErrRestaurantNotFound
		}
		return domain.MenuItem{}, fmt.Errorf("postgres: insert menu item: %w", err)
	}

	return item, nil
}

func (r *MenuRepository) updateMenuItem(ctx context.Context, item domain.MenuItem) (domain.MenuItem, error) {
	err := r.pool.QueryRow(ctx,
		`UPDATE menu_items SET name = $1, price_cents = $2
		 WHERE id = $3 AND restaurant_id = $4
		 RETURNING id, restaurant_id, name, price_cents, stock, created_at`,
		item.Name, item.PriceCents, item.ID, item.RestaurantID,
	).Scan(&item.ID, &item.RestaurantID, &item.Name, &item.PriceCents, &item.Stock, &item.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.MenuItem{}, domain.ErrMenuItemNotFound
		}
		return domain.MenuItem{}, fmt.Errorf("postgres: update menu item: %w", err)
	}

	return item, nil
}

func (r *MenuRepository) AdjustStock(ctx context.Context, restaurantID, menuItemID int32, delta int32) (domain.MenuItem, error) {
	var item domain.MenuItem
	err := r.pool.QueryRow(ctx,
		`UPDATE menu_items SET stock = stock + $1
		 WHERE id = $2 AND restaurant_id = $3
		 RETURNING id, restaurant_id, name, price_cents, stock, created_at`,
		delta, menuItemID, restaurantID,
	).Scan(&item.ID, &item.RestaurantID, &item.Name, &item.PriceCents, &item.Stock, &item.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.MenuItem{}, domain.ErrMenuItemNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgCheckViolation {
			return domain.MenuItem{}, domain.ErrMenuItemOutOfStock
		}
		return domain.MenuItem{}, fmt.Errorf("postgres: adjust stock: %w", err)
	}

	return item, nil
}
