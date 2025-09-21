package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	sqlcmock "kart/internal/mocks/sqlc"
	"kart/internal/repo"
	"kart/internal/sqlc"
)

func TestOrderService_PlaceOrder(t *testing.T) {
	type tc struct {
		name       string
		in         PlaceOrderInput
		setupSQLC  func(m *sqlcmock.Querier)
		setupDB    func(m sqlmock.Sqlmock)
		wantErr    bool
		assertGood func(t *testing.T, res PlaceOrderResult)
	}

	// Common inputs
	items := []OrderItemInput{{ProductID: "10", Quantity: 2}, {ProductID: "11", Quantity: 1}}

	cases := []tc{
		{
			name: "success without coupon",
			in:   PlaceOrderInput{CouponCode: "", Items: items},
			setupSQLC: func(m *sqlcmock.Querier) {
				m.On("GetProductsByIDs", mock.Anything, []string{"10", "11"}).
					Return([]sqlc.Product{{ID: "10"}, {ID: "11"}}, nil)
			},
			setupDB: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(`.*INSERT INTO orders \(id, coupon_code\) VALUES \(\$1, \$2\)`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(`.*INSERT INTO order_items \(id, order_id, product_id, quantity\)\s+SELECT UNNEST\(\$1::text\[\]\), UNNEST\(\$2::text\[\]\), UNNEST\(\$3::text\[\]\), UNNEST\(\$4::int4\[\]\)`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(2, 2))
				m.ExpectCommit()
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
			setupSQLC: func(m *sqlcmock.Querier) {
				m.On("GetCoupon", mock.Anything, "SAVE20AA").
					Return(sqlc.Coupon{Code: "SAVE20AA", PresenceMask: 3}, nil)
				m.On("GetProductsByIDs", mock.Anything, []string{"10", "11"}).
					Return([]sqlc.Product{{ID: "10"}, {ID: "11"}}, nil)
			},
			setupDB: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(`.*INSERT INTO coupon_redemptions \(code\)\s+VALUES \(\$1\)\s+ON CONFLICT DO NOTHING\s+RETURNING code`).
					WithArgs("SAVE20AA").
					WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("SAVE20AA"))
				m.ExpectExec(`.*INSERT INTO orders \(id, coupon_code\) VALUES \(\$1, \$2\)`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(`.*INSERT INTO order_items \(id, order_id, product_id, quantity\)\s+SELECT UNNEST\(\$1::text\[\]\), UNNEST\(\$2::text\[\]\), UNNEST\(\$3::text\[\]\), UNNEST\(\$4::int4\[\]\)`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(2, 2))
				m.ExpectCommit()
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
			setupSQLC: func(m *sqlcmock.Querier) {
				m.On("GetCoupon", mock.Anything, "SAVE20AA").
					Return(sqlc.Coupon{}, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "error coupon insufficient mask",
			in:   PlaceOrderInput{CouponCode: "SAVE20AA", Items: items},
			setupSQLC: func(m *sqlcmock.Querier) {
				m.On("GetCoupon", mock.Anything, "SAVE20AA").
					Return(sqlc.Coupon{Code: "SAVE20AA", PresenceMask: 1}, nil)
			},
			wantErr: true,
		},
		{
			name: "error products get many",
			in:   PlaceOrderInput{CouponCode: "", Items: items},
			setupSQLC: func(m *sqlcmock.Querier) {
				m.On("GetProductsByIDs", mock.Anything, []string{"10", "11"}).
					Return(([]sqlc.Product)(nil), errors.New("db down"))
			},
			wantErr: true,
		},
		{
			name: "error order insert items fail (rollback)",
			in:   PlaceOrderInput{CouponCode: "", Items: items},
			setupSQLC: func(m *sqlcmock.Querier) {
				m.On("GetProductsByIDs", mock.Anything, []string{"10", "11"}).
					Return([]sqlc.Product{{ID: "10"}, {ID: "11"}}, nil)
			},
			setupDB: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(`.*INSERT INTO orders \(id, coupon_code\) VALUES \(\$1, \$2\)`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				m.ExpectExec(`.*INSERT INTO order_items \(id, order_id, product_id, quantity\)\s+SELECT UNNEST\(\$1::text\[\]\), UNNEST\(\$2::text\[\]\), UNNEST\(\$3::text\[\]\), UNNEST\(\$4::int4\[\]\)`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("bad insert"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "error order insert fail (rollback)",
			in:   PlaceOrderInput{CouponCode: "", Items: items},
			setupSQLC: func(m *sqlcmock.Querier) {
				m.On("GetProductsByIDs", mock.Anything, []string{"10", "11"}).
					Return([]sqlc.Product{{ID: "10"}, {ID: "11"}}, nil)
			},
			setupDB: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectExec(`.*INSERT INTO orders \(id, coupon_code\) VALUES \(\$1, \$2\)`).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("bad order"))
				m.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "error coupon already redeemed",
			in:   PlaceOrderInput{CouponCode: "SAVE20AA", Items: items},
			setupSQLC: func(m *sqlcmock.Querier) {
				m.On("GetCoupon", mock.Anything, "SAVE20AA").
					Return(sqlc.Coupon{Code: "SAVE20AA", PresenceMask: 3}, nil)
				m.On("GetProductsByIDs", mock.Anything, []string{"10", "11"}).
					Return([]sqlc.Product{{ID: "10"}, {ID: "11"}}, nil)
			},
			setupDB: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(`.*INSERT INTO coupon_redemptions \(code\)\s+VALUES \(\$1\)\s+ON CONFLICT DO NOTHING\s+RETURNING code`).
					WithArgs("SAVE20AA").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			// sqlc mock for product & coupon repos
			qm := sqlcmock.NewQuerier(t)
			if c.setupSQLC != nil {
				c.setupSQLC(qm)
			}

			// sqlmock DB for order repo
			db, m, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			if c.setupDB != nil {
				c.setupDB(m)
			}

			pRepo := repo.NewProductRepo(qm)
			cRepo := repo.NewCouponRepo(qm)
			oRepo := repo.NewOrderRepo(db)
			svc := NewOrderService(pRepo, cRepo, oRepo)

			res, err := svc.PlaceOrder(ctx, c.in)
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if c.assertGood != nil {
					c.assertGood(t, res)
				}
			}
			require.NoError(t, m.ExpectationsWereMet())
		})
	}
}
