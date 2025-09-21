package sqlc

// Pin mockery version for reproducible mocks
//go:generate go run github.com/vektra/mockery/v2@v2.53.5 --name Querier --dir . --output ../mocks/sqlc --outpkg sqlcmock --filename querier_mock.go
