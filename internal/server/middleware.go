package server

import (
	"context"
	"errors"

	"github.com/getkin/kin-openapi/openapi3filter"
)

// NewOpenAPIAuthFunc returns an AuthenticationFunc which validates the
// api_key header for operations using the api_key security scheme.
func NewOpenAPIAuthFunc(expectedAPIKey string) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error {
		if ai == nil || ai.SecurityScheme == nil {
			return nil
		}
		if ai.SecuritySchemeName == "api_key" {
			req := ai.RequestValidationInput.Request
			if req == nil {
				return errors.New("missing request in auth input")
			}
			if req.Header.Get("api_key") != expectedAPIKey {
				return errors.New("invalid api_key")
			}
			return nil
		}
		return nil
	}
}
