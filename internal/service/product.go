package service

import (
	"context"
	"kart/internal/repo"
)

type ProductService struct{ Products repo.ProductRepository }

func NewProductService(p repo.ProductRepository) *ProductService { return &ProductService{Products: p} }

func (s *ProductService) List(ctx context.Context) ([]repo.Product, error) {
	return s.Products.List(ctx)
}

func (s *ProductService) Get(ctx context.Context, id string) (repo.Product, error) {
	return s.Products.Get(ctx, id)
}
