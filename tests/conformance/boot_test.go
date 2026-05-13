//go:build integration

package conformance_test

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
	"github.com/lee-mcfaul2/agent-sql-mcp/internal/store"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func bootForConformance(t *testing.T) (store.Pool, func()) {
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
	dsn, _ := pg.ConnectionString(ctx, "sslmode=disable")
	_, file, _, _ := runtime.Caller(0)
	migDir := filepath.Join(filepath.Dir(file), "..", "..", "migrations")
	m, _ := migrate.New("file://"+migDir, dsn)
	_ = m.Up()
	_ = sql.Drivers
	p, err := store.New(ctx, dsn, 8, 5)
	if err != nil {
		t.Fatal(err)
	}
	return p, func() {
		p.Close()
		_ = pg.Terminate(ctx)
	}
}
