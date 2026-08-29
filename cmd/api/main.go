package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"avito-kitchen/api"
	"avito-kitchen/internal/config"
	"avito-kitchen/internal/handlers"
	"avito-kitchen/internal/repository/postgres"
	"avito-kitchen/internal/usecase/menu"
	"avito-kitchen/internal/usecase/order"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer pool.Close()

	orderSvc := order.NewService(postgres.NewOrderRepository(pool))
	menuSvc := menu.NewService(postgres.NewMenuRepository(pool))

	ordersHandler := handlers.NewOrdersHandler(orderSvc)
	menuHandler := handlers.NewMenuHandler(menuSvc)
	restaurantsHandler := handlers.NewRestaurantsHandler(menuSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handlers.NewHealthHandler(pool).ServeHTTP)

	mux.HandleFunc("GET /restaurants", restaurantsHandler.List)
	mux.HandleFunc("GET /restaurants/{restaurantID}/menu", menuHandler.Get)
	mux.HandleFunc("POST /restaurants/{restaurantID}/menu-items", menuHandler.Create)
	mux.HandleFunc("PUT /restaurants/{restaurantID}/menu-items/{itemID}", menuHandler.Update)
	mux.HandleFunc("PATCH /restaurants/{restaurantID}/menu-items/{itemID}/stock", menuHandler.AdjustStock)
	mux.HandleFunc("GET /restaurants/{restaurantID}/orders", ordersHandler.ListByRestaurant)

	mux.HandleFunc("POST /orders", ordersHandler.Create)
	mux.HandleFunc("GET /orders/{orderID}", ordersHandler.Get)
	mux.HandleFunc("PATCH /orders/{orderID}/status", ordersHandler.UpdateStatus)

	mux.HandleFunc("GET /docs", handlers.NewDocsHandler())
	mux.HandleFunc("GET /openapi.yaml", handlers.NewOpenAPIHandler(api.OpenAPISpec))

	srv := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on :%s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
