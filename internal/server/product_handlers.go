package server

import (
	"fmt"
	"net/http"

	"kart/internal/openapi"
)

// ListProducts GET /product
func (s *Server) ListProducts(w http.ResponseWriter, r *http.Request) {
	ps, err := s.Products.List(r.Context())
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	out := make([]openapi.Product, 0, len(ps))
	for _, p := range ps {
		price := float32(p.PriceCents) / 100.0
		id := p.ID
		name := p.Name
		category := p.Category
		out = append(out, openapi.Product{Id: &id, Name: &name, Category: &category, Price: &price})
	}
	writeJSON(w, http.StatusOK, out)
}

// GetProduct GET /product/{productId}
func (s *Server) GetProduct(w http.ResponseWriter, r *http.Request, productId int64) {
	p, err := s.Products.Get(r.Context(), stringFromInt64(productId))
	if err != nil {
		http.Error(w, "Product not found", http.StatusNotFound)
		return
	}
	price := float32(p.PriceCents) / 100.0
	id := p.ID
	name := p.Name
	category := p.Category
	writeJSON(w, http.StatusOK, openapi.Product{Id: &id, Name: &name, Category: &category, Price: &price})
}

func stringFromInt64(v int64) string { return fmt.Sprintf("%d", v) }
