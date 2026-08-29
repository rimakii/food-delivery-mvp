package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	orderStatusNew       = "new"
	orderStatusConfirmed = "confirmed"
)

type order struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	apiBaseURL := getEnv("API_BASE_URL", "http://api:8080")
	restaurantID := getEnv("RESTAURANT_ID", "1")
	pollInterval := 5 * time.Second

	client := &http.Client{Timeout: 5 * time.Second}

	log.Printf("restaurant-service: polling %s for restaurant %s every %s", apiBaseURL, restaurantID, pollInterval)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("restaurant-service: shutting down")
			return
		case <-ticker.C:
			processNewOrders(ctx, client, apiBaseURL, restaurantID)
		}
	}
}

func processNewOrders(ctx context.Context, client *http.Client, apiBaseURL, restaurantID string) {
	orders, err := fetchOrders(ctx, client, apiBaseURL, restaurantID)
	if err != nil {
		log.Printf("restaurant-service: fetch orders: %v", err)
		return
	}

	for _, o := range orders {
		if o.Status != orderStatusNew {
			continue
		}

		if err := confirmOrder(ctx, client, apiBaseURL, o.ID); err != nil {
			log.Printf("restaurant-service: confirm order %d: %v", o.ID, err)
			continue
		}

		log.Printf("restaurant-service: order %d confirmed", o.ID)
	}
}

func fetchOrders(ctx context.Context, client *http.Client, apiBaseURL, restaurantID string) ([]order, error) {
	url := fmt.Sprintf("%s/restaurants/%s/orders", apiBaseURL, restaurantID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var orders []order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return orders, nil
}

func confirmOrder(ctx context.Context, client *http.Client, apiBaseURL string, orderID int64) error {
	body, err := json.Marshal(updateStatusRequest{Status: orderStatusConfirmed})
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}

	url := fmt.Sprintf("%s/orders/%d/status", apiBaseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
