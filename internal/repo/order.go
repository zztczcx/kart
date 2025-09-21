package repo

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	sqldb "kart/internal/sqlc"
	"kart/internal/store"
)

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
				return "", store.ErrNotFound
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
	for _, it := range items {
		if it.ID == "" {
			it.ID = uuid.NewString()
		}
		it.OrderID = o.ID
		err = q.InsertOrderItem(ctx, sqldb.InsertOrderItemParams{
			ID:        it.ID,
			OrderID:   it.OrderID,
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
		})
		if err != nil {
			return "", err
		}
	}
	if err = tx.Commit(); err != nil {
		return "", err
	}
	return o.ID, nil
}
