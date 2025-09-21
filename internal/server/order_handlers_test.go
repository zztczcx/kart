package server

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"kart/internal/config"
	servermock "kart/internal/mocks/server"
	"kart/internal/openapi"
	"kart/internal/repo"
	"kart/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPlaceOrder_Handler(t *testing.T) {
	type tc struct {
		name       string
		body       []byte
		setupMock  func(m *servermock.OrderService)
		wantStatus int
	}

	type item struct {
		ProductId string `json:"productId"`
		Quantity  int    `json:"quantity"`
	}
	mkBody := func(items []item) []byte {
		anon := make([]struct {
			ProductId string `json:"productId"`
			Quantity  int    `json:"quantity"`
		}, len(items))
		for i, it := range items {
			anon[i] = struct {
				ProductId string `json:"productId"`
				Quantity  int    `json:"quantity"`
			}{it.ProductId, it.Quantity}
		}
		b := openapi.PlaceOrderJSONRequestBody{Items: anon}
		buf, _ := json.Marshal(b)
		return buf
	}

	cases := []tc{
		{
			name: "ok",
			body: mkBody([]item{{"10", 1}, {"11", 2}}),
			setupMock: func(m *servermock.OrderService) {
				m.On("PlaceOrder", mock.Anything, mock.MatchedBy(func(in service.PlaceOrderInput) bool {
					return len(in.Items) == 2 && in.Items[0].ProductID == "10" && in.Items[0].Quantity == 1 && in.Items[1].ProductID == "11" && in.Items[1].Quantity == 2
				})).Return(service.PlaceOrderResult{
					OrderID: "order-id",
					Items: []service.OrderItemInput{
						{ProductID: "10", Quantity: 1},
						{ProductID: "11", Quantity: 2},
					},
					Products: []repo.Product{
						{ID: "10", Name: "Product 10", Category: "Category A", PriceCents: 1000},
						{ID: "11", Name: "Product 11", Category: "Category B", PriceCents: 2000},
					},
				}, nil)
			},
			wantStatus: 200,
		},
		{
			name:       "bad json",
			body:       []byte("{"),
			setupMock:  func(m *servermock.OrderService) {},
			wantStatus: 400,
		},
		{
			name:       "invalid items (empty)",
			body:       mkBody([]item{}),
			setupMock:  func(m *servermock.OrderService) {},
			wantStatus: 422,
		},
		{
			name:       "invalid items (qty<=0)",
			body:       mkBody([]item{{"10", 0}}),
			setupMock:  func(m *servermock.OrderService) {},
			wantStatus: 422,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := servermock.NewOrderService(t)
			c.setupMock(m)
			s := &Server{Orders: m, Cfg: config.Config{APIKey: "apitest"}}

			rr := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/order", bytes.NewReader(c.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("api_key", "apitest")

			s.PlaceOrder(rr, req)
			assert.Equal(t, c.wantStatus, rr.Code)
		})
	}
}
