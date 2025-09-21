package server

import (
	"net/http"
	"strconv"

	"kart/internal/openapi"
)

// ListProducts GET /product
func (s *Server) ListProducts(w http.ResponseWriter, r *http.Request) {
	ps, err := s.Products.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	out := make([]openapi.Product, 0, len(ps))
	for _, p := range ps {
		price := float32(p.PriceCents) / 100.0
		out = append(out, openapi.Product{
			Id:       ptr(p.ID),
			Name:     ptr(p.Name),
			Category: ptr(p.Category),
			Price:    ptr(price),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// GetProduct GET /product/{productId}
func (s *Server) GetProduct(w http.ResponseWriter, r *http.Request, productId int64) {
	p, err := s.Products.Get(r.Context(), strconv.FormatInt(productId, 10))
	if err != nil {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	price := float32(p.PriceCents) / 100.0
	writeJSON(w, http.StatusOK, openapi.Product{
		Id:       ptr(p.ID),
		Name:     ptr(p.Name),
		Category: ptr(p.Category),
		Price:    ptr(price),
	})
}
