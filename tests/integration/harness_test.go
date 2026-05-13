//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// bootPostgres starts a Postgres container + runs migrations.
// Returns DSN + cleanup.
func bootPostgres(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	pg, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("ag"),
		postgres.WithUsername("ag"),
		postgres.WithPassword("ag"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready").WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatal(err)
	}
	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	// Run migrations from the repo's migrations/ dir.
	_, file, _, _ := runtime.Caller(0)
	migDir := filepath.Join(filepath.Dir(file), "..", "..", "migrations")
	m, err := migrate.New("file://"+migDir, dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}

	// Sanity: count customers.
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow("SELECT count(*) FROM customers").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 100 {
		t.Fatalf("expected 100 customers after seed, got %d", n)
	}

	return dsn, func() { _ = pg.Terminate(ctx) }
}

func TestHarness_BootsAndSeeds(t *testing.T) {
	_, cleanup := bootPostgres(t)
	defer cleanup()
	// Passes if the harness ran migrations + the seed count check passed inline.
}
