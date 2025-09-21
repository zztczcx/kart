package store

import (
	"database/sql"
	"strings"
	"time"

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
	// reasonable pool defaults
	sdb.SetMaxOpenConns(20)
	sdb.SetMaxIdleConns(4)
	sdb.SetConnMaxLifetime(30 * time.Minute)
	sdb.SetConnMaxIdleTime(10 * time.Minute)

	if err := sdb.Ping(); err != nil { // Ping doesn't take context in database/sql
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
