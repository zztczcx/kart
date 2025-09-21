package repo

import (
	"context"
	"kart/internal/sqlc"
)

type Product = sqlc.Product
type Coupon = sqlc.Coupon
type Order = sqlc.Order
type OrderItem = sqlc.OrderItem

//go:generate mockery --name ProductRepository --dir . --output ../mocks/repo --outpkg repomock --filename product_repository_mock.go
//go:generate mockery --name CouponRepository --dir . --output ../mocks/repo --outpkg repomock --filename coupon_repository_mock.go
//go:generate mockery --name OrderRepository --dir . --output ../mocks/repo --outpkg repomock --filename order_repository_mock.go

type ProductRepository interface {
	List(ctx context.Context) ([]Product, error)
	Get(ctx context.Context, id string) (Product, error)
	GetMany(ctx context.Context, ids []string) (map[string]Product, error)
}

type CouponRepository interface {
	Get(ctx context.Context, code string) (Coupon, error)
}

type OrderRepository interface {
	CreateWithItems(ctx context.Context, o Order, items []OrderItem) (string, error)
}
