package store

import (
	"database/sql"
	"errors"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	*sql.DB
}

func Open(databaseURL string) (*DB, error) {
	var driver string
	url := databaseURL
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		driver = "pgx"
	} else {
		// default to postgres
		driver = "pgx"
		url = "postgres://user:pass@localhost:5432/kart?sslmode=disable"
	}

	sdb, err := sql.Open(driver, url)
	if err != nil {
		return nil, err
	}
	if err := sdb.Ping(); err != nil {
		_ = sdb.Close()
		return nil, err
	}
	return &DB{DB: sdb}, nil
}

func (db *DB) Close() error {
	if db == nil || db.DB == nil {
		return nil
	}
	return db.DB.Close()
}

// Migrations are handled externally (docker-compose migrate service)

// Seed replaced by goose dev migration

var ErrNotFound = errors.New("not found")
