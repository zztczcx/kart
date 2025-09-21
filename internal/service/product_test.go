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
	ctx := context.Background()
	m := sqlcmock.NewQuerier(t)
	m.On("ListProducts", mock.Anything).
		Return([]sqlc.Product{{ID: "1"}, {ID: "2"}}, nil)
	repo := repo.NewProductRepo(m)
	svc := NewProductService(repo)
	got, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2, got %d", len(got))
	}
}

func TestProductService_Get(t *testing.T) {
	ctx := context.Background()
	// ok
	m1 := sqlcmock.NewQuerier(t)
	m1.On("GetProduct", mock.Anything, "1").
		Return(sqlc.Product{ID: "1"}, nil)
	repo1 := repo.NewProductRepo(m1)
	svc1 := NewProductService(repo1)
	p, err := svc1.Get(ctx, "1")
	if err != nil || p.ID != "1" {
		t.Fatalf("unexpected: %+v %v", p, err)
	}
	// err
	m2 := sqlcmock.NewQuerier(t)
	m2.On("GetProduct", mock.Anything, "x").
		Return(sqlc.Product{}, errors.New("not found"))
	repo2 := repo.NewProductRepo(m2)
	svc2 := NewProductService(repo2)
	_, err = svc2.Get(ctx, "x")
	if err == nil {
		t.Fatalf("expected error")
	}
}
