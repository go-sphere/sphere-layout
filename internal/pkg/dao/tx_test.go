package dao

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-sphere/sphere-layout/internal/pkg/database/client"
	"github.com/go-sphere/sphere-layout/internal/pkg/database/ent"
)

func testClient(t *testing.T) *ent.Client {
	t.Helper()
	db, err := client.NewDataBaseClient(client.Config{
		Type: "sqlite3",
		Path: fmt.Sprintf("file:dao-tx-test-%d?mode=memory&cache=shared", time.Now().UnixNano()),
	})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestWithTxExPanicRollsBackAndRepanics(t *testing.T) {
	db := testClient(t)
	ctx := context.Background()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic to be re-raised after rollback")
		}
	}()

	_ = WithTxEx(ctx, db, func(ctx context.Context, tx *ent.Client) error {
		if err := tx.Admin.Create().
			SetUsername("panic-admin").
			SetPassword("x").
			SetRoles([]string{"all"}).
			Exec(ctx); err != nil {
			t.Fatalf("insert in tx: %v", err)
		}
		panic("boom")
	})
}

func TestWithTxExPanicDidNotCommit(t *testing.T) {
	db := testClient(t)
	ctx := context.Background()

	func() {
		defer func() { _ = recover() }()
		_ = WithTxEx(ctx, db, func(ctx context.Context, tx *ent.Client) error {
			if err := tx.Admin.Create().
				SetUsername("panic-admin").
				SetPassword("x").
				SetRoles([]string{"all"}).
				Exec(ctx); err != nil {
				return err
			}
			panic("boom")
		})
	}()

	count, err := db.Admin.Query().Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("panic path committed %d admins, want 0", count)
	}
}

func TestWithTxPanicRollsBackAndRepanics(t *testing.T) {
	db := testClient(t)
	ctx := t.Context()

	panicked := false
	func() {
		defer func() {
			panicked = recover() != nil
		}()
		_, _ = WithTx(ctx, db, func(ctx context.Context, tx *ent.Client) (*int, error) {
			if err := tx.Admin.Create().
				SetUsername("panic-admin").
				SetPassword("x").
				SetRoles([]string{"all"}).
				Exec(ctx); err != nil {
				return nil, err
			}
			panic("boom")
		})
	}()
	if !panicked {
		t.Fatal("expected panic to be re-raised after rollback")
	}

	count, err := db.Admin.Query().Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("panic path committed %d admins, want 0", count)
	}
}
