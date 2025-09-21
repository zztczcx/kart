package server

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	oapimw "github.com/oapi-codegen/nethttp-middleware"

	"kart/internal/openapi"
)

// NewRouter creates and configures a chi Router with OpenAPI request validation.
func NewRouter(handlers openapi.ServerInterface) http.Handler {
	loader := &openapi3.Loader{IsExternalRefsAllowed: true}
	specPath := "api/openapi.yaml"
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		specPath = "/api/openapi.yaml"
	}
	spec, err := loader.LoadFromFile(specPath)
	if err != nil {
		log.Fatalf("failed to load OpenAPI spec: %v", err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		log.Fatalf("invalid OpenAPI spec: %v", err)
	}

	r := chi.NewRouter()
	auth := NewOpenAPIAuthFunc(handlers.(*Server).Cfg.APIKey)
	r.Use(oapimw.OapiRequestValidatorWithOptions(spec, &oapimw.Options{
		SilenceServersWarning: true,
		Options:               openapi3filter.Options{AuthenticationFunc: auth},
	}))

	return openapi.HandlerWithOptions(handlers, openapi.ChiServerOptions{
		BaseRouter: r,
	})
}
