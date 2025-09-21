package server

import (
	"encoding/json"
	"net/http"

	"kart/internal/openapi"
	"kart/internal/service"
)

// PlaceOrder POST /order
func (s *Server) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req openapi.OrderReq
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// basic input validation at the edge
	if len(req.Items) == 0 {
		http.Error(w, "validation error: no items", http.StatusUnprocessableEntity)
		return
	}
	in := make([]service.OrderItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		if it.ProductId == "" || it.Quantity <= 0 {
			http.Error(w, "validation error: invalid item", http.StatusUnprocessableEntity)
			return
		}
		in = append(in, service.OrderItemInput{ProductID: it.ProductId, Quantity: int32(it.Quantity)})
	}

	result, err := s.Orders.PlaceOrder(r.Context(), service.PlaceOrderInput{
		CouponCode: deref(req.CouponCode),
		Items:      in,
	})
	if err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	// Build response according to OpenAPI spec
	items := make([]openapi.OrderItem, 0, len(result.Items))
	for _, item := range result.Items {
		qty := int(item.Quantity)
		items = append(items, openapi.OrderItem{
			ProductId: ptr(item.ProductID),
			Quantity:  ptr(qty),
		})
	}

	products := make([]openapi.Product, 0, len(result.Products))
	for _, p := range result.Products {
		price := float32(p.PriceCents) / 100.0
		products = append(products, openapi.Product{
			Id:       ptr(p.ID),
			Name:     ptr(p.Name),
			Category: ptr(p.Category),
			Price:    ptr(price),
		})
	}

	writeJSON(w, http.StatusOK, openapi.Order{
		Id:       &result.OrderID,
		Items:    &items,
		Products: &products,
	})
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
