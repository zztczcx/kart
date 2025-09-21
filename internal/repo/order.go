package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	sqldb "kart/internal/sqlc"
)

// ErrCouponRedeemed indicates a single-use coupon has already been redeemed.
var ErrCouponRedeemed = errors.New("coupon redeemed")

type OrderRepo struct{ db *sql.DB }

func NewOrderRepo(db *sql.DB) *OrderRepo { return &OrderRepo{db: db} }

func (r *OrderRepo) CreateWithItems(ctx context.Context, o Order, items []OrderItem) (string, error) {
	if o.ID == "" {
		o.ID = uuid.NewString()
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	q := sqldb.New(tx)

	// Single-use coupon redemption within the same transaction
	if o.CouponCode.Valid {
		if _, err := q.TryRedeemSingleUse(ctx, o.CouponCode.String); err != nil {
			// sqlc returns sql.ErrNoRows when ON CONFLICT DO NOTHING prevented insert
			if err == sql.ErrNoRows {
				return "", ErrCouponRedeemed
			}
			return "", err
		}
	}
	err = q.InsertOrder(ctx, sqldb.InsertOrderParams{
		ID:         o.ID,
		CouponCode: o.CouponCode,
	})
	if err != nil {
		return "", err
	}
	if len(items) > 0 {
		ids := make([]string, len(items))
		orderIDs := make([]string, len(items))
		productIDs := make([]string, len(items))
		quantities := make([]int32, len(items))
		for i := range items {
			if items[i].ID == "" {
				items[i].ID = uuid.NewString()
			}
			items[i].OrderID = o.ID
			ids[i] = items[i].ID
			orderIDs[i] = items[i].OrderID
			productIDs[i] = items[i].ProductID
			quantities[i] = items[i].Quantity
		}
		if err = q.InsertOrderItems(ctx, sqldb.InsertOrderItemsParams{
			Column1: ids,
			Column2: orderIDs,
			Column3: productIDs,
			Column4: quantities,
		}); err != nil {
			return "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return o.ID, nil
}
