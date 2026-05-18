package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"

	"smart-pet-monitoring/backend/internal/config"
)

func main() {
	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if command != "up" {
		log.Fatalf("unsupported migration command %q", command)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := runUp(cfg.DatabaseURL, migrationsDir()); err != nil {
		log.Fatalf("run migrations: %v", err)
	}
}

func migrationsDir() string {
	if value := os.Getenv("MIGRATIONS_PATH"); value != "" {
		return value
	}
	return "migrations"
}

func runUp(databaseURL, dir string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	connConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("parse database url: %w", err)
	}
	connConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	conn, err := pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer conn.Close(context.Background())

	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("list migration files: %w", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		return fmt.Errorf("no migration files found in %s", dir)
	}

	for _, file := range files {
		version := filepath.Base(file)
		applied, err := alreadyApplied(ctx, conn, version)
		if err != nil {
			return err
		}
		if applied {
			log.Printf("migration %s already applied", version)
			continue
		}
		if err := applyMigration(ctx, conn, file, version); err != nil {
			return err
		}
		log.Printf("migration %s applied", version)
	}

	return nil
}

func alreadyApplied(ctx context.Context, conn *pgx.Conn, version string) (bool, error) {
	var exists bool
	err := conn.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return exists, nil
}

func applyMigration(ctx context.Context, conn *pgx.Conn, file, version string) error {
	sql, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", version, err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}
	defer func() {
		if err := tx.Rollback(context.Background()); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			log.Printf("rollback migration %s: %v", version, err)
		}
	}()

	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("execute migration %s: %w", version, err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}
	return nil
}
