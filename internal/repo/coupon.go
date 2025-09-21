package repo

import (
	"context"

	sqldb "kart/internal/sqlc"
)

type CouponRepo struct{ q sqldb.Querier }

func NewCouponRepo(q sqldb.Querier) *CouponRepo { return &CouponRepo{q: q} }

func (r *CouponRepo) Get(ctx context.Context, code string) (Coupon, error) {
	c, err := r.q.GetCoupon(ctx, code)
	if err != nil {
		return Coupon{}, err
	}

	return Coupon(c), nil
}
