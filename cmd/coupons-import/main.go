package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"kart/internal/config"
	"kart/internal/sqlc"
	"kart/internal/store"
)

// Importer ingests coupon codes from a newline-delimited file into the database.
// It streams the file to avoid excessive memory use and batches inserts for throughput.
type Importer struct {
	db        *sql.DB
	batchSize int
}

func newImporter(db *sql.DB, batchSize int) *Importer {
	if batchSize <= 0 {
		batchSize = 1000
	}
	return &Importer{db: db, batchSize: batchSize}
}

func (im *Importer) insertBatch(ctx context.Context, codes []string) error {
	if len(codes) == 0 {
		return nil
	}

	tx, err := im.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		// Ensure rollback if not committed
		_ = tx.Rollback()
	}()

	// Insert with initial presence bit; on conflict, increment 8-bit counter (wrap at 256)
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO coupons (code, presence_mask) VALUES ($1, B'00000001') ON CONFLICT (code) DO UPDATE SET presence_mask = (((coupons.presence_mask)::int + 1) % 256)::bit(8)")
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	for _, code := range codes {
		if code == "" {
			continue
		}
		if _, err := stmt.ExecContext(ctx, code); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func run(ctx context.Context, path string, batchSize int) error {
	if path == "" {
		return errors.New("-file is required")
	}

	cfg := config.Load()
	sdb, err := store.Open(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer func() { _ = sdb.Close() }()

	// Simple health check to fail fast
	if err := sdb.PingContext(ctx); err != nil {
		return fmt.Errorf("db ping: %w", err)
	}

	// Touch sqlc to ensure it's linked and available for migrations in other cmds
	_ = sqlc.New(sdb)

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	reader := bufio.NewReaderSize(f, 1024*1024) // 1MB buffer to handle long lines efficiently
	scanner := bufio.NewScanner(reader)
	// Increase the scanner buffer to accommodate very long lines
	const maxCapacity = 1024 * 1024 // 1MB per line; adjust if codes can exceed this
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, maxCapacity)

	importer := newImporter(sdb.DB, batchSize)

	var (
		batch   []string
		total   int64
		lastLog = time.Now()
	)

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		if err := importer.insertBatch(ctx, batch); err != nil {
			return err
		}
		total += int64(len(batch))
		batch = batch[:0]
		now := time.Now()
		if now.Sub(lastLog) >= 5*time.Second {
			log.Printf("imported %d coupons...", total)
			lastLog = now
		}
		return nil
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		batch = append(batch, line)
		if len(batch) >= importer.batchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan: %w", err)
	}
	if err := flush(); err != nil {
		return err
	}
	log.Printf("completed import: %d coupons", total)
	return nil
}

// nextPresenceMask allocates the next available presence bit [0..7] for this import run.
// It uses a small coordination table and an exclusive table lock to avoid races.
// presence logic removed per request; only importing coupon codes now

func main() {
	var (
		filePath  string
		batchSize int
	)
	flag.StringVar(&filePath, "file", "", "path to newline-delimited coupon codes file")
	flag.IntVar(&batchSize, "batch", 2000, "number of rows per transaction")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := run(ctx, filePath, batchSize); err != nil {
		log.Fatalf("import failed: %v", err)
	}
}
