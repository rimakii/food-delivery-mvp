//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestE2E_OrderJourney(t *testing.T) {
	baseURL := os.Getenv("E2E_API_URL")
	if baseURL == "" {
		t.Skip("E2E_API_URL not set, skipping black-box e2e test")
	}

	client := &http.Client{Timeout: 5 * time.Second}
	waitForHealth(t, client, baseURL)

	const restaurantID = 1

	var menuItems []struct {
		ID    int32 `json:"id"`
		Stock int32 `json:"stock"`
	}
	getJSON(t, client, baseURL+fmt.Sprintf("/restaurants/%d/menu", restaurantID), http.StatusOK, &menuItems)
	if len(menuItems) == 0 {
		t.Fatal("seeded menu is empty — did migrate-e2e run?")
	}
	item := menuItems[0]

	postJSON(t, client, baseURL+"/orders", map[string]any{
		"restaurant_id": restaurantID,
		"items":         []map[string]any{{"menu_item_id": item.ID, "quantity": item.Stock + 1000}},
	}, http.StatusConflict, nil)

	var created struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	postJSON(t, client, baseURL+"/orders", map[string]any{
		"restaurant_id": restaurantID,
		"items":         []map[string]any{{"menu_item_id": item.ID, "quantity": 1}},
	}, http.StatusCreated, &created)
	if created.Status != "new" {
		t.Errorf("status = %q, want new", created.Status)
	}

	ctx := context.Background()
	statusURL := fmt.Sprintf("%s/orders/%d/status", baseURL, created.ID)
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

	var order struct {
		Status string `json:"status"`
	}
	getJSON(t, client, fmt.Sprintf("%s/orders/%d", baseURL, created.ID), http.StatusOK, &order)
	if order.Status != "confirmed" {
		t.Errorf("final status = %q, want confirmed", order.Status)
	}
}

func waitForHealth(t *testing.T, client *http.Client, baseURL string) {
	t.Helper()

	ctx := context.Background()
	deadline := time.Now().Add(30 * time.Second)

	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health", nil)
		if err != nil {
			t.Fatalf("build health request: %v", err)
		}

		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		time.Sleep(500 * time.Millisecond)
	}

	t.Fatalf("api at %s did not become healthy within 30s", baseURL)
}

func getJSON(t *testing.T, client *http.Client, url string, wantStatus int, out any) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := client.Do(req)
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

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(mustJSON(t, body)))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
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
