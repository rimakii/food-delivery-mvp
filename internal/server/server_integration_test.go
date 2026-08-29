//go:build integration

package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"avito-kitchen/internal/repository/postgres"
	"avito-kitchen/internal/server"
)

func TestE2E_OrderJourney(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping e2e test")
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(pool.Close)

	ts := httptest.NewServer(server.New(pool))
	t.Cleanup(ts.Close)
	client := ts.Client()

	var restaurantID int32
	if err := pool.QueryRow(ctx,
		`INSERT INTO restaurants (name) VALUES ($1) RETURNING id`, "e2e-"+t.Name(),
	).Scan(&restaurantID); err != nil {
		t.Fatalf("seed restaurant: %v", err)
	}

	var menuItemID int32
	if err := pool.QueryRow(ctx,
		`INSERT INTO menu_items (restaurant_id, name, price_cents, stock) VALUES ($1, $2, $3, $4) RETURNING id`,
		restaurantID, "Roll", 500, 2,
	).Scan(&menuItemID); err != nil {
		t.Fatalf("seed menu item: %v", err)
	}

	// 1. заведение видно в общем списке
	var restaurants []struct {
		ID int32 `json:"id"`
	}
	getJSON(t, client, ts.URL+"/restaurants", http.StatusOK, &restaurants)
	if !containsID(restaurants, restaurantID) {
		t.Fatalf("restaurant %d not found in %+v", restaurantID, restaurants)
	}

	// 2. в меню видна засеянная позиция с корректным stock
	menuURL := fmt.Sprintf("%s/restaurants/%d/menu", ts.URL, restaurantID)
	var menuItems []struct {
		ID    int32 `json:"id"`
		Stock int32 `json:"stock"`
	}
	getJSON(t, client, menuURL, http.StatusOK, &menuItems)
	if len(menuItems) != 1 || menuItems[0].Stock != 2 {
		t.Fatalf("unexpected menu: %+v", menuItems)
	}

	// 3. заказ больше, чем есть в наличии -> 409, stock не тронут
	postJSON(t, client, ts.URL+"/orders", map[string]any{
		"restaurant_id": restaurantID,
		"items":         []map[string]any{{"menu_item_id": menuItemID, "quantity": 5}},
	}, http.StatusConflict, nil)

	getJSON(t, client, menuURL, http.StatusOK, &menuItems)
	if menuItems[0].Stock != 2 {
		t.Fatalf("stock after failed order = %d, want 2 (unchanged)", menuItems[0].Stock)
	}

	// 4. пустой заказ -> 400
	postJSON(t, client, ts.URL+"/orders", map[string]any{
		"restaurant_id": restaurantID,
		"items":         []map[string]any{},
	}, http.StatusBadRequest, nil)

	// 5. валидный заказ -> 201, списывает stock целиком
	var created struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	postJSON(t, client, ts.URL+"/orders", map[string]any{
		"restaurant_id": restaurantID,
		"items":         []map[string]any{{"menu_item_id": menuItemID, "quantity": 2}},
	}, http.StatusCreated, &created)
	if created.Status != "new" {
		t.Errorf("status = %q, want new", created.Status)
	}

	getJSON(t, client, menuURL, http.StatusOK, &menuItems)
	if menuItems[0].Stock != 0 {
		t.Errorf("stock after order = %d, want 0", menuItems[0].Stock)
	}

	// 6. заведение видит заказ в своём списке (без позиций — по дизайну)
	ordersURL := fmt.Sprintf("%s/restaurants/%d/orders", ts.URL, restaurantID)
	var restaurantOrders []struct {
		ID int64 `json:"id"`
	}
	getJSON(t, client, ordersURL, http.StatusOK, &restaurantOrders)
	if len(restaurantOrders) != 1 || restaurantOrders[0].ID != created.ID {
		t.Fatalf("unexpected restaurant orders: %+v", restaurantOrders)
	}

	// 7. заведение подтверждает заказ
	statusURL := fmt.Sprintf("%s/orders/%d/status", ts.URL, created.ID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, statusURL, bytes.NewReader(mustJSON(t, map[string]string{"status": "confirmed"})))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("patch status: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("patch status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	// 8. клиент видит подтверждённый заказ с позициями
	var order struct {
		Status string `json:"status"`
		Items  []struct {
			Quantity int `json:"quantity"`
		} `json:"items"`
	}
	getJSON(t, client, fmt.Sprintf("%s/orders/%d", ts.URL, created.ID), http.StatusOK, &order)
	if order.Status != "confirmed" {
		t.Errorf("final status = %q, want confirmed", order.Status)
	}
	if len(order.Items) != 1 || order.Items[0].Quantity != 2 {
		t.Errorf("unexpected items: %+v", order.Items)
	}
}

func containsID(restaurants []struct {
	ID int32 `json:"id"`
}, id int32,
) bool {
	for _, r := range restaurants {
		if r.ID == id {
			return true
		}
	}

	return false
}

func getJSON(t *testing.T, client *http.Client, url string, wantStatus int, out any) {
	t.Helper()

	resp, err := client.Get(url) //nolint:noctx // тестовый хелпер, контекст не нужен
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		t.Fatalf("GET %s: status = %d, want %d", url, resp.StatusCode, wantStatus)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("GET %s: decode: %v", url, err)
		}
	}
}

func postJSON(t *testing.T, client *http.Client, url string, body any, wantStatus int, out any) {
	t.Helper()

	resp, err := client.Post(url, "application/json", bytes.NewReader(mustJSON(t, body))) //nolint:noctx // тестовый хелпер
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		t.Fatalf("POST %s: status = %d, want %d", url, resp.StatusCode, wantStatus)
	}
	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			t.Fatalf("POST %s: decode: %v", url, err)
		}
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()

	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	return b
}
