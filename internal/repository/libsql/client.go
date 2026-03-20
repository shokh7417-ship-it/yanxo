package libsqlrepo

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tursodatabase/libsql-client-go/libsql"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type DB struct {
	SQL *sql.DB
}

func Open(ctx context.Context, databaseURL, authToken string) (*DB, error) {
	// IMPORTANT:
	// The libsql driver does NOT read auth token from env here.
	// Use a connector with WithAuthToken for Turso remote auth.
	authToken = strings.TrimSpace(authToken)
	conn, err := libsql.NewConnector(databaseURL, libsql.WithAuthToken(authToken))
	if err != nil {
		return nil, err
	}
	db := sql.OpenDB(conn)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return &DB{SQL: db}, nil
}

func (d *DB) Close() error { return d.SQL.Close() }

func RunMigrations(ctx context.Context, db *sql.DB, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(strings.ToLower(name), ".sql") {
			files = append(files, filepath.Join(migrationsDir, name))
		}
	}
	sort.Strings(files)

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY);`); err != nil {
		return err
	}

	for _, p := range files {
		version := filepath.Base(p)
		var exists string
		err := db.QueryRowContext(ctx, `SELECT version FROM schema_migrations WHERE version = ?`, version).Scan(&exists)
		if err == nil {
			continue
		}
		if err != sql.ErrNoRows {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		sqlText := strings.TrimSpace(string(b))
		if sqlText == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, sqlText); err != nil {
			return fmt.Errorf("migration %s: %w", version, err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES (?)`, version); err != nil {
			return err
		}
	}
	return nil
}

