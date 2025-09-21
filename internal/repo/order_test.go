package repo

import (
	"context"
	"regexp"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderRepo_CreateWithItems(t *testing.T) {
	type tc struct {
		name              string
		buildExpectations func(mock sqlmock.Sqlmock)
		order             Order
		items             []OrderItem
		wantErr           bool
	}
	cases := []tc{
		{
			name: "success two items",
			buildExpectations: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders (id, coupon_code) VALUES ($1, $2)`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO order_items (id, order_id, product_id, quantity)
SELECT UNNEST($1::text[]), UNNEST($2::text[]), UNNEST($3::text[]), UNNEST($4::int4[])`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(2, 2))
				mock.ExpectCommit()
			},
			order: Order{},
			items: []OrderItem{{ProductID: "10", Quantity: 1}, {ProductID: "11", Quantity: 1}},
		},
		{
			name: "rollback on first item error",
			buildExpectations: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO orders (id, coupon_code) VALUES ($1, $2)`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO order_items (id, order_id, product_id, quantity)
SELECT UNNEST($1::text[]), UNNEST($2::text[]), UNNEST($3::text[]), UNNEST($4::int4[])`)).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(assert.AnError)
				mock.ExpectRollback()
			},
			order:   Order{},
			items:   []OrderItem{{ProductID: "10", Quantity: 0}},
			wantErr: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			r := NewOrderRepo(db)
			c.buildExpectations(mock)
			_, err = r.CreateWithItems(context.Background(), c.order, c.items)
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
