package server

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	servermock "kart/internal/mocks/server"
	"kart/internal/openapi"
	"kart/internal/sqlc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListProducts_Handler(t *testing.T) {
	type tc struct {
		name       string
		mockSetup  func(m *servermock.ProductService)
		wantStatus int
		wantLen    int
	}
	cases := []tc{
		{
			name: "ok two products",
			mockSetup: func(m *servermock.ProductService) {
				m.On("List", mock.Anything).Return([]sqlc.Product{{ID: "1"}, {ID: "2"}}, nil)
			},
			wantStatus: 200,
			wantLen:    2,
		},
		{
			name:       "ok empty",
			mockSetup:  func(m *servermock.ProductService) { m.On("List", mock.Anything).Return([]sqlc.Product{}, nil) },
			wantStatus: 200,
			wantLen:    0,
		},
		{
			name:       "service error",
			mockSetup:  func(m *servermock.ProductService) { m.On("List", mock.Anything).Return(nil, assert.AnError) },
			wantStatus: 500,
			wantLen:    0,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := servermock.NewProductService(t)
			c.mockSetup(m)
			s := &Server{Products: m}

			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/product", nil)
			s.ListProducts(rr, req)

			assert.Equal(t, c.wantStatus, rr.Code)
			var got []openapi.Product
			_ = json.Unmarshal(rr.Body.Bytes(), &got)
			if c.wantStatus == 200 {
				assert.Len(t, got, c.wantLen)
			}
		})
	}
}

func TestGetProduct_Handler(t *testing.T) {
	type tc struct {
		name      string
		productID int64
		mockSetup func(m *servermock.ProductService)
		want      int
	}
	cases := []tc{
		{
			name:      "ok",
			productID: 1,
			mockSetup: func(m *servermock.ProductService) { m.On("Get", mock.Anything, "1").Return(sqlc.Product{ID: "1"}, nil) },
			want:      200,
		},
		{
			name:      "not found",
			productID: 2,
			mockSetup: func(m *servermock.ProductService) {
				m.On("Get", mock.Anything, "2").Return(sqlc.Product{}, assert.AnError)
			},
			want: 404,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := servermock.NewProductService(t)
			c.mockSetup(m)
			s := &Server{Products: m}
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/product/x", nil)
			s.GetProduct(rr, req, c.productID)
			assert.Equal(t, c.want, rr.Code)
		})
	}
}
