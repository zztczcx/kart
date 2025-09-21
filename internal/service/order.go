package service

import (
	"context"
	"database/sql"
	"errors"
	"kart/internal/repo"
	"math/bits"
)

type OrderService struct {
	Products *repo.ProductRepo
	Coupons  *repo.CouponRepo
	Orders   *repo.OrderRepo
}

func NewOrderService(p *repo.ProductRepo, c *repo.CouponRepo, o *repo.OrderRepo) *OrderService {
	return &OrderService{Products: p, Coupons: c, Orders: o}
}

type OrderItemInput struct {
	ProductID string
	Quantity  int32
}

type PlaceOrderInput struct {
	CouponCode string
	Items      []OrderItemInput
}

type PlaceOrderResult struct {
	OrderID  string
	Items    []OrderItemInput
	Products []repo.Product
}

func (s *OrderService) PlaceOrder(ctx context.Context, in PlaceOrderInput) (PlaceOrderResult, error) {
	valid, err := s.validateCoupon(ctx, in.CouponCode)
	if err != nil || !valid {
		return PlaceOrderResult{}, err
	}

	productsByID, err := s.fetchProductsMap(ctx, in.Items)
	if err != nil {
		return PlaceOrderResult{}, err
	}

	items, err := s.buildOrderItems(productsByID, in.Items)
	if err != nil {
		return PlaceOrderResult{}, err
	}

	coupon := sql.NullString{String: in.CouponCode, Valid: in.CouponCode != ""}
	orderID, err := s.Orders.CreateWithItems(
		ctx,
		repo.Order{CouponCode: coupon},
		items,
	)
	if err != nil {
		return PlaceOrderResult{}, err
	}

	// Collect products for response
	ps := make([]repo.Product, 0, len(productsByID))
	for _, p := range productsByID {
		ps = append(ps, p)
	}

	return PlaceOrderResult{
		OrderID:  orderID,
		Items:    in.Items,
		Products: ps,
	}, nil
}

func (s *OrderService) fetchProductsMap(ctx context.Context, items []OrderItemInput) (map[string]repo.Product, error) {
	uniq := make(map[string]struct{}, len(items))
	ids := make([]string, 0, len(items))
	for _, it := range items {
		if _, ok := uniq[it.ProductID]; !ok {
			uniq[it.ProductID] = struct{}{}
			ids = append(ids, it.ProductID)
		}
	}
	return s.Products.GetMany(ctx, ids)
}

func (s *OrderService) buildOrderItems(productsByID map[string]repo.Product, inputs []OrderItemInput) ([]repo.OrderItem, error) {
	items := make([]repo.OrderItem, 0, len(inputs))
	for _, in := range inputs {
		p, ok := productsByID[in.ProductID]
		if !ok {
			return nil, errors.New("product not found")
		}
		items = append(items, repo.OrderItem{
			ProductID: p.ID,
			Quantity:  in.Quantity,
		})
	}
	return items, nil
}

func (s *OrderService) validateCoupon(ctx context.Context, couponCode string) (bool, error) {
	if couponCode == "" {
		return true, nil
	}
	// Must be a string of length between 8 and 10 characters
	if len(couponCode) < 8 || len(couponCode) > 10 {
		return false, errors.New("coupon code must be between 8 and 10 characters")
	}

	c, err := s.Coupons.Get(ctx, couponCode)
	if err != nil {
		return false, err
	}

	// Require coupon to apply to at least 2 categories
	n := bits.OnesCount8(c.PresenceMask)
	if n < 2 {
		return false, errors.New("coupon must apply to at least two categories")
	}
	return true, nil
}
