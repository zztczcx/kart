package repo

import (
	"context"
	"testing"

	sqlcmock "kart/internal/mocks/sqlc"
	"kart/internal/sqlc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProductRepo_List(t *testing.T) {
	type tc struct {
		name    string
		setup   func(m *sqlcmock.Querier)
		wantLen int
		wantErr bool
	}
	cases := []tc{
		{
			name: "two",
			setup: func(m *sqlcmock.Querier) {
				m.On("ListProducts", mock.Anything).Return([]sqlc.Product{{ID: "1"}, {ID: "2"}}, nil)
			},
			wantLen: 2,
		},
		{
			name:    "empty",
			setup:   func(m *sqlcmock.Querier) { m.On("ListProducts", mock.Anything).Return([]sqlc.Product{}, nil) },
			wantLen: 0,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := sqlcmock.NewQuerier(t)
			c.setup(m)
			r := NewProductRepo(m)
			got, err := r.List(context.Background())
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Len(t, got, c.wantLen)
		})
	}
}

func TestProductRepo_Get(t *testing.T) {
	type tc struct {
		name    string
		id      string
		setup   func(m *sqlcmock.Querier)
		wantID  string
		wantErr bool
	}
	cases := []tc{
		{
			name:   "ok",
			id:     "1",
			setup:  func(m *sqlcmock.Querier) { m.On("GetProduct", mock.Anything, "1").Return(sqlc.Product{ID: "1"}, nil) },
			wantID: "1",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := sqlcmock.NewQuerier(t)
			c.setup(m)
			r := NewProductRepo(m)
			got, err := r.Get(context.Background(), c.id)
			if c.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, c.wantID, got.ID)
		})
	}
}
