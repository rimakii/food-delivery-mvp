package server

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"avito-kitchen/api"
	"avito-kitchen/internal/handlers"
	"avito-kitchen/internal/repository/postgres"
	"avito-kitchen/internal/usecase/menu"
	"avito-kitchen/internal/usecase/order"
)

func New(pool *pgxpool.Pool) http.Handler {
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

	return mux
}
