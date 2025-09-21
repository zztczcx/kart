package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	repomock "kart/internal/mocks/repo"
	"kart/internal/repo"
)

type _ = repo.Product // ensure repo types referenced

type _ = mock.Arguments // ensure testify/mock imported

func TestOrderService_PlaceOrder(t *testing.T) {
	type tc struct {
		name       string
		in         PlaceOrderInput
		setupMocks func(p *repomock.ProductRepository, c *repomock.CouponRepository, o *repomock.OrderRepository)
		wantErr    bool
		assertGood func(t *testing.T, res PlaceOrderResult)
	}

	// Common inputs
	items := []OrderItemInput{{ProductID: "10", Quantity: 2}, {ProductID: "11", Quantity: 1}}

	cases := []tc{
		{
			name: "success without coupon",
			in:   PlaceOrderInput{CouponCode: "", Items: items},
			setupMocks: func(p *repomock.ProductRepository, _ *repomock.CouponRepository, o *repomock.OrderRepository) {
				p.On("GetMany", mock.Anything, []string{"10", "11"}).
					Return(map[string]repo.Product{"10": {ID: "10"}, "11": {ID: "11"}}, nil)
				o.On("CreateWithItems", mock.Anything, mock.Anything, mock.MatchedBy(func(items []repo.OrderItem) bool { return len(items) == 2 })).
					Return("order-1", nil)
			},
			assertGood: func(t *testing.T, res PlaceOrderResult) {
				require.NotEmpty(t, res.OrderID)
				require.Len(t, res.Products, 2)
				require.Equal(t, items, res.Items)
			},
		},
		{
			name: "success with valid coupon",
			in:   PlaceOrderInput{CouponCode: "SAVE20AA", Items: items},
			setupMocks: func(p *repomock.ProductRepository, c *repomock.CouponRepository, o *repomock.OrderRepository) {
				c.On("Get", mock.Anything, "SAVE20AA").
					Return(repo.Coupon{Code: "SAVE20AA", PresenceMask: 3}, nil)
				p.On("GetMany", mock.Anything, []string{"10", "11"}).
					Return(map[string]repo.Product{"10": {ID: "10"}, "11": {ID: "11"}}, nil)
				o.On("CreateWithItems", mock.Anything, mock.Anything, mock.Anything).
					Return("order-2", nil)
			},
			assertGood: func(t *testing.T, res PlaceOrderResult) {
				require.NotEmpty(t, res.OrderID)
				require.Len(t, res.Products, 2)
			},
		},
		{
			name:    "error coupon too short",
			in:      PlaceOrderInput{CouponCode: "ABC", Items: items},
			wantErr: true,
		},
		{
			name: "error coupon fetch",
			in:   PlaceOrderInput{CouponCode: "SAVE20AA", Items: items},
			setupMocks: func(_ *repomock.ProductRepository, c *repomock.CouponRepository, _ *repomock.OrderRepository) {
				c.On("Get", mock.Anything, "SAVE20AA").
					Return(repo.Coupon{}, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "error coupon insufficient mask",
			in:   PlaceOrderInput{CouponCode: "SAVE20AA", Items: items},
			setupMocks: func(_ *repomock.ProductRepository, c *repomock.CouponRepository, _ *repomock.OrderRepository) {
				c.On("Get", mock.Anything, "SAVE20AA").
					Return(repo.Coupon{Code: "SAVE20AA", PresenceMask: 1}, nil)
			},
			wantErr: true,
		},
		{
			name: "error products get many",
			in:   PlaceOrderInput{CouponCode: "", Items: items},
			setupMocks: func(p *repomock.ProductRepository, _ *repomock.CouponRepository, _ *repomock.OrderRepository) {
				p.On("GetMany", mock.Anything, []string{"10", "11"}).
					Return((map[string]repo.Product)(nil), errors.New("db down"))
			},
			wantErr: true,
		},
		{
			name: "error order insert fail (rollback)",
			in:   PlaceOrderInput{CouponCode: "", Items: items},
			setupMocks: func(p *repomock.ProductRepository, _ *repomock.CouponRepository, o *repomock.OrderRepository) {
				p.On("GetMany", mock.Anything, []string{"10", "11"}).
					Return(map[string]repo.Product{"10": {ID: "10"}, "11": {ID: "11"}}, nil)
				o.On("CreateWithItems", mock.Anything, mock.Anything, mock.Anything).
					Return("", errors.New("bad order"))
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			p := repomock.NewProductRepository(t)
			co := repomock.NewCouponRepository(t)
			o := repomock.NewOrderRepository(t)
			if c.setupMocks != nil {
				c.setupMocks(p, co, o)
			}

			svc := NewOrderService(p, co, o)

			res, err := svc.PlaceOrder(ctx, c.in)
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if c.assertGood != nil {
					c.assertGood(t, res)
				}
			}
		})
	}
}
