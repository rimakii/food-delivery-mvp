package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"avito-kitchen/internal/domain"
	"avito-kitchen/internal/usecase/order"
)

var _ order.Repository = (*OrderRepository)(nil)

type OrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{pool: pool}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, restaurantID int32, items []domain.OrderItem) (domain.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("postgres: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var orderID int64
	var createdAt time.Time
	err = tx.QueryRow(ctx,
		`INSERT INTO orders (restaurant_id, status) VALUES ($1, $2) RETURNING id, created_at`,
		restaurantID, domain.OrderStatusNew,
	).Scan(&orderID, &createdAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolation {
			return domain.Order{}, domain.ErrRestaurantNotFound
		}
		return domain.Order{}, fmt.Errorf("postgres: insert order: %w", err)
	}

	order := domain.Order{
		ID:           orderID,
		RestaurantID: restaurantID,
		Status:       domain.OrderStatusNew,
		Items:        make([]domain.OrderItem, 0, len(items)),
		CreatedAt:    createdAt,
	}

	for _, item := range items {
		var priceCents int64
		err = tx.QueryRow(ctx,
			`UPDATE menu_items SET stock = stock - $1
			 WHERE id = $2 AND restaurant_id = $3 AND stock >= $1
			 RETURNING price_cents`,
			item.Quantity, item.MenuItemID, restaurantID,
		).Scan(&priceCents)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.Order{}, fmt.Errorf("postgres: menu item %d: %w", item.MenuItemID, domain.ErrMenuItemOutOfStock)
			}
			return domain.Order{}, fmt.Errorf("postgres: decrement stock: %w", err)
		}

		var orderItemID int64
		err = tx.QueryRow(ctx,
			`INSERT INTO order_items (order_id, menu_item_id, quantity, price_cents)
			 VALUES ($1, $2, $3, $4) RETURNING id`,
			orderID, item.MenuItemID, item.Quantity, priceCents,
		).Scan(&orderItemID)
		if err != nil {
			return domain.Order{}, fmt.Errorf("postgres: insert order item: %w", err)
		}

		order.Items = append(order.Items, domain.OrderItem{
			ID:         orderItemID,
			OrderID:    orderID,
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			PriceCents: priceCents,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("postgres: commit: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) GetOrder(ctx context.Context, id int64) (domain.Order, error) {
	var order domain.Order
	err := r.pool.QueryRow(ctx,
		`SELECT id, restaurant_id, status, created_at FROM orders WHERE id = $1`,
		id,
	).Scan(&order.ID, &order.RestaurantID, &order.Status, &order.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, domain.ErrOrderNotFound
		}
		return domain.Order{}, fmt.Errorf("postgres: get order: %w", err)
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, order_id, menu_item_id, quantity, price_cents
		 FROM order_items
		 WHERE order_id = $1
		 ORDER BY id`,
		order.ID,
	)
	if err != nil {
		return domain.Order{}, fmt.Errorf("postgres: get order items: %w", err)
	}
	defer rows.Close()

	order.Items, err = pgx.CollectRows(rows, pgx.RowToStructByName[domain.OrderItem])
	if err != nil {
		return domain.Order{}, fmt.Errorf("postgres: scan order items: %w", err)
	}

	return order, nil
}

func (r *OrderRepository) ListOrdersByRestaurant(ctx context.Context, restaurantID int32) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, restaurant_id, status, created_at
		 FROM orders
		 WHERE restaurant_id = $1
		 ORDER BY id`,
		restaurantID,
	)
	if err != nil {
		return nil, fmt.Errorf("postgres: list orders: %w", err)
	}
	defer rows.Close()

	orders := make([]domain.Order, 0)
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.RestaurantID, &o.Status, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("postgres: scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list orders: %w", err)
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id int64, status domain.OrderStatus) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE orders SET status = $1 WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("postgres: update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}
