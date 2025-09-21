package service

import (
	"context"
	"errors"
	"testing"

	sqlcmock "kart/internal/mocks/sqlc"
	"kart/internal/repo"
	"kart/internal/sqlc"

	"github.com/stretchr/testify/mock"
)

func TestProductService_List(t *testing.T) {
	type tc struct {
		name      string
		setupMock func(m *sqlcmock.Querier)
		wantLen   int
		wantErr   bool
	}

	cases := []tc{
		{
			name: "success two products",
			setupMock: func(m *sqlcmock.Querier) {
				m.On("ListProducts", mock.Anything).
					Return([]sqlc.Product{{ID: "1"}, {ID: "2"}}, nil)
			},
			wantLen: 2,
		},
		{
			name: "error from ListProducts",
			setupMock: func(m *sqlcmock.Querier) {
				m.On("ListProducts", mock.Anything).
					Return(([]sqlc.Product)(nil), errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			m := sqlcmock.NewQuerier(t)
			if c.setupMock != nil {
				c.setupMock(m)
			}
			repo := repo.NewProductRepo(m)
			svc := NewProductService(repo)
			got, err := svc.List(ctx)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != c.wantLen {
				t.Fatalf("want %d, got %d", c.wantLen, len(got))
			}
		})
	}
}

func TestProductService_Get(t *testing.T) {
	type tc struct {
		name      string
		id        string
		setupMock func(m *sqlcmock.Querier)
		wantID    string
		wantErr   bool
	}

	cases := []tc{
		{
			name: "found",
			id:   "1",
			setupMock: func(m *sqlcmock.Querier) {
				m.On("GetProduct", mock.Anything, "1").
					Return(sqlc.Product{ID: "1"}, nil)
			},
			wantID: "1",
		},
		{
			name: "not found",
			id:   "x",
			setupMock: func(m *sqlcmock.Querier) {
				m.On("GetProduct", mock.Anything, "x").
					Return(sqlc.Product{}, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()
			m := sqlcmock.NewQuerier(t)
			if c.setupMock != nil {
				c.setupMock(m)
			}
			repo := repo.NewProductRepo(m)
			svc := NewProductService(repo)
			p, err := svc.Get(ctx, c.id)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.ID != c.wantID {
				t.Fatalf("want %s, got %s", c.wantID, p.ID)
			}
		})
	}
}
