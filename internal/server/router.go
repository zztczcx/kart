package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	oapimw "github.com/oapi-codegen/nethttp-middleware"

	"kart/internal/openapi"
)

// NewRouter creates and configures a chi Router with OpenAPI request validation.
// It returns an error instead of exiting the process to enable graceful startup handling.
func NewRouter(apiKey string, handlers openapi.ServerInterface) (http.Handler, error) {
	loader := &openapi3.Loader{IsExternalRefsAllowed: true}
	specPath := "api/openapi.yaml"
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		specPath = "/api/openapi.yaml"
	}
	spec, err := loader.LoadFromFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("load OpenAPI spec: %w", err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("validate OpenAPI spec: %w", err)
	}

	r := chi.NewRouter()
	auth := NewOpenAPIAuthFunc(apiKey)
	r.Use(oapimw.OapiRequestValidatorWithOptions(spec, &oapimw.Options{
		SilenceServersWarning: true,
		Options:               openapi3filter.Options{AuthenticationFunc: auth},
	}))

	h := openapi.HandlerWithOptions(handlers, openapi.ChiServerOptions{
		BaseRouter: r,
	})
	return h, nil
}
