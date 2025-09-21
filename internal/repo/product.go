package repo

import (
	"context"

	sqldb "kart/internal/sqlc"
	"kart/internal/store"
)

type ProductRepo struct{ q sqldb.Querier }

func NewProductRepo(q sqldb.Querier) *ProductRepo { return &ProductRepo{q: q} }

func (r *ProductRepo) List(ctx context.Context) ([]Product, error) {
	rows, err := r.q.ListProducts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Product, 0, len(rows))
	for _, p := range rows {
		out = append(out, Product(p))
	}
	return out, nil
}

func (r *ProductRepo) Get(ctx context.Context, id string) (Product, error) {
	pr, err := r.q.GetProduct(ctx, id)
	if err != nil {
		// sqlc uses database/sql errors; map no rows at call site if desired
		return Product{}, err
	}
	// Map not found to domain error if needed by checking store.ErrNotFound equivalence higher up
	if pr.ID == "" {
		return Product{}, store.ErrNotFound
	}
	return Product(pr), nil
}

// GetMany returns products for the given IDs. Missing IDs are not included.
func (r *ProductRepo) GetMany(ctx context.Context, ids []string) (map[string]Product, error) {
	rows, err := r.q.GetProductsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[string]Product, len(rows))
	for _, p := range rows {
		out[p.ID] = Product(p)
	}
	return out, nil
}
