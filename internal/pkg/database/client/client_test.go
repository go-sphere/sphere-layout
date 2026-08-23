package client

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestAutoMigrateDoesNotDropColumnsByDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "migration.db")
	config := Config{Type: "sqlite3", Path: path}
	db, err := NewDataBaseClient(config)
	if err != nil {
		t.Fatalf("create database: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), "ALTER TABLE admins ADD COLUMN legacy_value TEXT"); err != nil {
		_ = db.Close()
		t.Fatalf("add legacy column: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	db, err = NewDataBaseClient(config)
	if err != nil {
		t.Fatalf("migrate with safe defaults: %v", err)
	}
	if !hasColumn(t, db, "admins", "legacy_value") {
		_ = db.Close()
		t.Fatal("safe auto-migration dropped a column")
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}

	config.AutoMigrateDrop = true
	db, err = NewDataBaseClient(config)
	if err != nil {
		t.Fatalf("migrate with destructive option: %v", err)
	}
	defer db.Close()
	if hasColumn(t, db, "admins", "legacy_value") {
		t.Fatal("destructive auto-migration did not drop the legacy column")
	}
}

func hasColumn(t *testing.T, db interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}, table, column string) bool {
	t.Helper()
	rows, err := db.QueryContext(context.Background(), "PRAGMA table_info("+table+")")
	if err != nil {
		t.Fatalf("inspect table: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, kind string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &kind, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan table info: %v", err)
		}
		if name == column {
			return true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate table info: %v", err)
	}
	return false
}
