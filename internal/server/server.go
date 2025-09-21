package server

import (
	"context"
	"kart/internal/config"
	"kart/internal/openapi"
	"kart/internal/repo"
	"kart/internal/service"
)

//go:generate mockery --name ProductService --dir . --output ../mocks/server --outpkg servermock --filename product_service_mock.go
//go:generate mockery --name OrderService --dir . --output ../mocks/server --outpkg servermock --filename service_mocks.go

// ProductService is the minimal interface the handlers need.
type ProductService interface {
	List(ctx context.Context) ([]repo.Product, error)
	Get(ctx context.Context, id string) (repo.Product, error)
}

// OrderService is the minimal interface the handlers need.
type OrderService interface {
	PlaceOrder(ctx context.Context, in service.PlaceOrderInput) (service.PlaceOrderResult, error)
}

// Server holds dependencies for HTTP handlers.
type Server struct {
	Cfg      config.Config
	Products ProductService
	Orders   OrderService
}

// Ensure Server implements the generated interface.
var _ openapi.ServerInterface = (*Server)(nil)
